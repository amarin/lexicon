package rules

import (
	"fmt"
	"slices"
	"strings"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/textnorm"
)

var tokenKinds = map[string]textnorm.TokenKind{
	"word":   textnorm.TokenWord,
	"number": textnorm.TokenNumber,
	"punct":  textnorm.TokenPunct,
	"symbol": textnorm.TokenSymbol,
}

// termCond is the compiled per-term conditions of an element (lemma,
// grammeme, token kind, shape); every condition that is set must hold.
type termCond struct {
	lemmas    keyword
	hasLemmas bool
	grammemes []string
	kinds     []textnorm.TokenKind
	shape     shapeCheck
}

func newTermCond(e Element) (termCond, error) {
	var tc termCond
	if e.Lemma != "" {
		k, err := newKeyword(e.Lemma, false)
		if err != nil {
			return termCond{}, err
		}
		tc.lemmas, tc.hasLemmas = k, true
	}
	if e.Grammeme != "" {
		if tc.grammemes = splitAlt(e.Grammeme); len(tc.grammemes) == 0 {
			return termCond{}, fmt.Errorf("empty grammeme %q", e.Grammeme)
		}
	}
	if e.Token != "" {
		for _, name := range splitAlt(e.Token) {
			k, ok := tokenKinds[name]
			if !ok {
				return termCond{}, fmt.Errorf("unknown token kind %q (want word, number, punct or symbol)", name)
			}
			tc.kinds = append(tc.kinds, k)
		}
		if len(tc.kinds) == 0 {
			return termCond{}, fmt.Errorf("empty token kind %q", e.Token)
		}
	}
	shape, err := compileShape(e.Shape)
	if err != nil {
		return termCond{}, err
	}
	tc.shape = shape
	return tc, nil
}

// empty reports whether no condition is set.
func (tc termCond) empty() bool {
	return !tc.hasLemmas && tc.grammemes == nil && tc.kinds == nil && !tc.shape.hasCase && !tc.shape.hasScript
}

// ok reports whether t satisfies every condition that is set.
func (tc termCond) ok(t *lexicon.Term) bool {
	if tc.hasLemmas && !tc.lemmas.match(t) {
		return false
	}
	if tc.grammemes != nil && !hasAnyGrammeme(t, tc.grammemes) {
		return false
	}
	if tc.kinds != nil && !slices.Contains(tc.kinds, t.Token.Kind) {
		return false
	}
	return tc.shape.ok(t)
}

// hasAnyGrammeme: a known lemma of t carries one of gs in its tag.
func hasAnyGrammeme(t *lexicon.Term, gs []string) bool {
	for _, l := range t.Lemmas {
		if l.Flags&lexicon.FlagUnknown != 0 {
			continue
		}
		for _, f := range strings.FieldsFunc(l.Tag, func(r rune) bool { return r == ',' || r == ' ' }) {
			if slices.Contains(gs, f) {
				return true
			}
		}
	}
	return false
}

// splitAlt splits "a|b" into its non-empty, trimmed alternatives.
func splitAlt(s string) []string {
	var out []string
	for _, p := range strings.Split(s, "|") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
