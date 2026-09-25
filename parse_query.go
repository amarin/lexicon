package lexicon

import "github.com/amarin/lexicon/textnorm"

// ParseQuery parses a search query with the same analysis as ModeIndex. The
// last indexable word, if nothing follows it (End == len(q)), becomes Partial.
// A compound at the end is taken whole, without parts.
func (a *Analyzer) ParseQuery(q string, p Profile) Query {
	var toks []textnorm.Token

	for _, t := range textnorm.Tokenize(a.rules, q) {
		if indexable(t) {
			toks = append(toks, t)
		}
	}

	var out Query

	if k := len(toks); k > 0 {
		last := toks[k-1]
		if last.End == len(q) && last.Kind != textnorm.TokenNumber {
			r := a.word(last.Form, p)

			t := Term{Token: last, Form: last.Form, Lemmas: r.lemmas}
			if r.stop {
				t.Lemmas = nil
			}

			out.Partial = &t
			toks = toks[:k-1]
		}
	}

	for _, tok := range toks {
		out.Terms = append(out.Terms, a.indexTerms(tok, p)...)
	}

	return out
}
