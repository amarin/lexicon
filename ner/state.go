package ner

import (
	"fmt"
	"strings"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/gazetteer"
)

// state holds one Extract call.
type state struct {
	p       *Pipeline
	tx      *gazetteer.Text
	terms   []lexicon.Term
	explain bool
	cands   []*candidate
}

func (s *state) term(pos int) *lexicon.Term { return &s.terms[s.tx.TermIndex(pos)] }

// note appends evidence when Explain is on.
func (s *state) note(c *candidate, format string, args ...any) {
	if s.explain {
		c.evidence = append(c.evidence, fmt.Sprintf(format, args...))
	}
}

// lemmaNormals returns the lemma sequences of positions [a, b), capped at
// gazetteer.MaxLemmaKeys (decision D13).
func (s *state) lemmaNormals(a, b int) []string {
	alts := make([][]string, 0, b-a)
	for p := a; p < b; p++ {
		alts = append(alts, gazetteer.Alternatives(s.term(p)))
	}
	combos, _ := gazetteer.Combine(alts, gazetteer.MaxLemmaKeys)
	out := make([]string, len(combos))
	for i, c := range combos {
		out[i] = strings.Join(c, " ")
	}
	return out
}
