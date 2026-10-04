package ner

import (
	"sort"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/rules"
)

// lattice exposes one sentence (terms [t0, t1)) and the live candidates
// inside it to rules patterns. Span IDs are indexes into state.cands;
// starts is shared by the sentences of one pattern run.
type lattice struct {
	s      *state
	t0, t1 int
	starts [][]rules.LatticeSpan // by absolute term index; End relative to the sentence
}

// latticeSpans indexes the live candidates by their first term, longest
// first. A candidate never crosses a sentence end.
func (s *state) latticeSpans() [][]rules.LatticeSpan {
	from := make([]int, len(s.terms)) // sentence start of each term
	to := make([]int, len(s.terms))   // sentence end of each term
	for _, seg := range s.tx.Sentences() {
		for t := seg[0]; t < seg[1]; t++ {
			from[t], to[t] = seg[0], seg[1]
		}
	}
	starts := make([][]rules.LatticeSpan, len(s.terms))
	for i, c := range s.cands {
		if c.removed {
			continue
		}
		ts, te := s.tx.TermIndex(c.start), s.tx.TermIndex(c.end-1)+1
		if te > to[ts] {
			continue
		}
		starts[ts] = append(starts[ts], rules.LatticeSpan{ID: i, Type: c.typ, End: te - from[ts]})
	}
	for _, sp := range starts {
		sort.SliceStable(sp, func(a, b int) bool {
			if sp[a].End != sp[b].End {
				return sp[a].End > sp[b].End
			}
			return sp[a].ID < sp[b].ID
		})
	}
	return starts
}

func (l *lattice) Len() int                          { return l.t1 - l.t0 }
func (l *lattice) Term(i int) *lexicon.Term          { return &l.s.terms[l.t0+i] }
func (l *lattice) SpansAt(i int) []rules.LatticeSpan { return l.starts[l.t0+i] }
