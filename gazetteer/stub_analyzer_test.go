package gazetteer

import (
	"strings"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/textnorm"
)

// stubAnalyzer splits text on spaces. A word "Form/l1/l2" becomes a Word term
// with form strings.ToLower("Form") and lemmas l1, l2; the word "." is a
// Punct term. It keeps gazetteer unit tests independent of v0.1 behaviour.
type stubAnalyzer struct{ version string }

func (s stubAnalyzer) Version() string { return s.version }

func (s stubAnalyzer) Analyze(text string, _ lexicon.Profile, _ lexicon.Mode) []lexicon.Term {
	var out []lexicon.Term
	off := 0
	for _, w := range strings.Fields(text) {
		start := strings.Index(text[off:], w) + off
		off = start + len(w)
		parts := strings.Split(w, "/")
		tok := textnorm.Token{Raw: w, Start: start, End: off, Kind: textnorm.TokenWord}
		if parts[0] == "." {
			tok.Kind = textnorm.TokenPunct
			out = append(out, lexicon.Term{Token: tok})
			continue
		}
		term := lexicon.Term{Token: tok, Form: strings.ToLower(parts[0])}
		for _, l := range parts[1:] {
			term.Lemmas = append(term.Lemmas, lexicon.Lemma{Text: l})
		}
		out = append(out, term)
	}
	return out
}
