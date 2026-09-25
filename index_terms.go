package lexicon

import "github.com/amarin/lexicon/textnorm"

// indexable: ModeIndex considers Cyrillic words with a form, and numbers.
func indexable(t textnorm.Token) bool {
	switch t.Kind {
	case textnorm.TokenNumber:
		return true
	case textnorm.TokenWord:
		return t.Script == textnorm.ScriptCyrillic && t.Form != ""
	}

	return false
}

// indexTerms are the ModeIndex terms of one token (genodex P1 rules): a number
// is an unknown lemma; a dotted word is first looked up as an abbreviation
// with its dot; a hyphenated word is first looked up whole as an
// abbreviation, otherwise indexed whole and by parts; stop words are dropped.
func (a *Analyzer) indexTerms(tok textnorm.Token, p Profile) []Term {
	if !indexable(tok) {
		return nil
	}

	if tok.Kind == textnorm.TokenNumber {
		return []Term{unknownTerm(tok, tok.Form)}
	}

	if tok.Dotted {
		if ls := a.dottedAbbrev(tok.Form); ls != nil {
			return []Term{{Token: tok, Form: tok.Form + ".", Lemmas: ls}}
		}
	}

	parts := textnorm.SplitHyphen(tok)
	if parts == nil {
		if r := a.word(tok.Form, p); !r.stop {
			return []Term{{Token: tok, Form: tok.Form, Lemmas: r.lemmas}}
		}

		return nil
	}

	if ls := a.abbrev(tok.Form); ls != nil {
		return []Term{{Token: tok, Form: tok.Form, Lemmas: ls}}
	}

	var out []Term

	for _, t := range append([]textnorm.Token{tok}, parts...) {
		if t.Form == "" {
			continue
		}

		if r := a.word(t.Form, p); !r.stop {
			out = append(out, Term{Token: t, Form: t.Form, Lemmas: r.lemmas})
		}
	}

	return out
}
