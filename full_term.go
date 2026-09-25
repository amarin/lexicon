package lexicon

import "github.com/amarin/lexicon/textnorm"

// fullTerm is the ModeFull term of one token: numbers and non-Cyrillic words
// get an unknown lemma (no dictionary lookup), punctuation and symbols no
// lemmas; a dotted word is first looked up as an abbreviation with its dot, a
// hyphenated word as a whole abbreviation; stop words keep FlagStop lemmas.
func (a *Analyzer) fullTerm(tok textnorm.Token, p Profile) Term {
	switch {
	case tok.Kind == textnorm.TokenNumber:
		return unknownTerm(tok, tok.Form)
	case tok.Kind != textnorm.TokenWord || tok.Form == "":
		return Term{Token: tok, Form: tok.Form}
	case tok.Script != textnorm.ScriptCyrillic:
		return unknownTerm(tok, tok.Form)
	}

	if tok.Dotted {
		if ls := a.dottedAbbrev(tok.Form); ls != nil {
			return Term{Token: tok, Form: tok.Form + ".", Lemmas: ls}
		}
	}

	if textnorm.SplitHyphen(tok) != nil {
		if ls := a.abbrev(tok.Form); ls != nil {
			return Term{Token: tok, Form: tok.Form, Lemmas: ls}
		}
	}

	return Term{Token: tok, Form: tok.Form, Lemmas: a.word(tok.Form, p).lemmas}
}
