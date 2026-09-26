package ner

import (
	"slices"
	"sort"
)

// resolver runs weighted interval scheduling with nesting. all is sorted
// by end position; memo caches the best selection inside each candidate.
type resolver struct {
	nesting map[string]map[string]bool
	all     []*candidate
	memo    map[*candidate]*selection
}

// schedule returns the best selection of mutually non-overlapping cs
// (sorted by end). A candidate's value includes its nested selection.
func (r *resolver) schedule(cs []*candidate) *selection {
	n := len(cs)
	if n == 0 {
		return &selection{}
	}
	best := make([]float64, n+1)
	take := make([]bool, n+1)
	prev := make([]int, n+1)
	for j := 1; j <= n; j++ {
		c := cs[j-1]
		p := sort.Search(j-1, func(i int) bool { return cs[i].end > c.start })
		incl := c.score + r.inner(c).value + best[p]
		if incl > best[j-1] {
			best[j], take[j], prev[j] = incl, true, p
		} else {
			best[j] = best[j-1]
		}
	}
	sel := &selection{value: best[n]}
	for j := n; j > 0; {
		if take[j] {
			sel.items = append(sel.items, cs[j-1])
			j = prev[j]
		} else {
			j--
		}
	}
	slices.Reverse(sel.items)
	return sel
}

// inner returns the best selection strictly inside c among the types
// allowed by Nesting[c.typ].
func (r *resolver) inner(c *candidate) *selection {
	if sel, ok := r.memo[c]; ok {
		return sel
	}
	var in []*candidate
	if allowed := r.nesting[c.typ]; len(allowed) > 0 {
		// all is sorted by end: a candidate inside c ends in (c.start, c.end],
		// so only that slice of all is scanned, in all's order.
		lo := sort.Search(len(r.all), func(i int) bool { return r.all[i].end > c.start })
		hi := sort.Search(len(r.all), func(i int) bool { return r.all[i].end > c.end })
		for _, x := range r.all[lo:hi] {
			if x != c && allowed[x.typ] && x.start >= c.start && x.end <= c.end && x.words() < c.words() {
				in = append(in, x)
			}
		}
	}
	sel := r.schedule(in)
	r.memo[c] = sel
	return sel
}
