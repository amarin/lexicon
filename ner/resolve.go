package ner

import (
	"math"
	"sort"
)

// resolve selects the output candidates. Order of equal-end candidates is
// deterministic: start desc, score desc, type asc, creation order.
func (s *state) resolve() ([]*candidate, map[*candidate]bool) {
	order := make(map[*candidate]int, len(s.cands))
	var live []*candidate
	for i, c := range s.cands {
		order[c] = i
		if !c.removed && c.score > 0 {
			live = append(live, c)
		}
	}
	sort.SliceStable(live, func(i, j int) bool {
		a, b := live[i], live[j]
		if a.end != b.end {
			return a.end < b.end
		}
		if a.start != b.start {
			return a.start > b.start
		}
		if a.score != b.score {
			return a.score > b.score
		}
		if a.typ != b.typ {
			return a.typ < b.typ
		}
		return order[a] < order[b]
	})
	r := &resolver{nesting: s.p.nesting, all: live, memo: map[*candidate]*selection{}}
	var chosen []*candidate
	nested := map[*candidate]bool{}
	var walk func(sel *selection, inside bool)
	walk = func(sel *selection, inside bool) {
		for _, c := range sel.items {
			chosen = append(chosen, c)
			if inside {
				nested[c] = true
			}
			walk(r.inner(c), true)
		}
	}
	walk(r.schedule(live), false)
	s.attachAlternatives(chosen, live)
	return chosen, nested
}

// attachAlternatives keeps every live loser on a winner's exact range with
// another type as an alternative of the winner (D14) and flags ties (D7).
// Two candidates on one range never both win: nesting needs a strictly
// shorter inner candidate.
func (s *state) attachAlternatives(chosen, live []*candidate) {
	for _, c := range chosen {
		for _, o := range live {
			if o == c || o.typ == c.typ || o.start != c.start || o.end != c.end {
				continue
			}
			c.alts = append(c.alts, o)
			if math.Abs(o.score-c.score) < epsilon {
				c.flags |= Ambiguous
				s.note(c, "tie with %s", o.typ)
			}
		}
	}
}
