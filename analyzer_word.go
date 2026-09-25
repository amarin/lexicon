package lexicon

import (
	"slices"
	"unicode/utf8"

	"github.com/amarin/lexicon/textnorm"
)

// word is the analysis of a single form under profile p (cached from Task 10).
func (a *Analyzer) word(form string, p Profile) wordResult {
	return a.lookupWord(form, p)
}

// lookupWord: exact readings (trying pre-reform ending variants when there are
// none), otherwise base predictions, otherwise the form as an unknown lemma.
// One-letter forms never take abbreviation readings (they need a dot, D14).
func (a *Analyzer) lookupWord(form string, p Profile) wordResult {
	exact, predicted := a.parse(form, p.Kinds)
	if utf8.RuneCountInString(form) == 1 {
		exact = dropKind(exact, KindAbbrev)
	}

	var extra Flag

	if len(exact) == 0 {
		for _, v := range textnorm.ReformVariants(form) {
			if e, _ := a.parse(v, p.Kinds); len(e) > 0 {
				exact, extra = e, FlagReform

				break
			}
		}
	}

	if allStop(exact) {
		return wordResult{lemmas: a.lemmas(exact, FlagStop|extra), stop: true}
	}

	ls := a.lemmas(p.filter(exact), extra)
	if ls == nil {
		ls = a.lemmas(predicted, FlagPredicted)
	}

	if ls == nil {
		ls = []Lemma{{Text: form, Flags: FlagUnknown}}
	}

	return wordResult{lemmas: ls}
}

// parse splits readings into exact and predicted ones. Predictions are kept
// only from the base dictionary: gomorphy predicts for every built dictionary,
// and a small user dictionary (surnames, abbreviations) would "predict" a
// lemma for almost any word. The Registry already enforces this; the Analyzer
// re-checks so any Dictionaries implementation behaves the same (D11).
func (a *Analyzer) parse(form string, kinds []Kind) (exact, predicted []Reading) {
	for _, r := range a.dicts.Parse(form, kinds) {
		switch {
		case !r.Predicted:
			exact = append(exact, r)
		case r.Kind == KindBase:
			predicted = append(predicted, r)
		}
	}

	return exact, predicted
}

// lemmas are the distinct lemmas of readings (normalized like words), each
// with extra, FlagAbbrev for abbreviation dictionaries and FlagAmbiguous when
// there are several. nil when rs is empty.
func (a *Analyzer) lemmas(rs []Reading, extra Flag) []Lemma {
	var out []Lemma

	for _, r := range rs {
		text := textnorm.NormalizeWord(a.rules, r.Normal)
		if slices.ContainsFunc(out, func(l Lemma) bool { return l.Text == text }) {
			continue
		}

		f := extra
		if r.Kind == KindAbbrev {
			f |= FlagAbbrev
		}

		out = append(out, Lemma{Text: text, Flags: f, Tag: r.Tag, Kind: r.Kind, Dict: r.Dict})
	}

	if len(out) > 1 {
		for i := range out {
			out[i].Flags |= FlagAmbiguous
		}
	}

	return out
}

// dropKind removes readings of kind k.
func dropKind(rs []Reading, k Kind) []Reading {
	var out []Reading

	for _, r := range rs {
		if r.Kind != k {
			out = append(out, r)
		}
	}

	return out
}

// unknownTerm is a term without readings: the lemma is the form.
func unknownTerm(tok textnorm.Token, form string) Term {
	return Term{Token: tok, Form: form, Lemmas: []Lemma{{Text: form, Flags: FlagUnknown}}}
}
