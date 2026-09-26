package ner

import "slices"

// posIndex buckets candidates by start and by end content position, so the
// rule stages look only at the few positions a keyword window can reach
// instead of every candidate of the document. Buckets keep insertion order.
type posIndex struct {
	byStart  [][]*candidate // byStart[p]: candidates with start == p
	byEnd    [][]*candidate // byEnd[p]: candidates with end == p
	maxWords int            // longest indexed candidate, in content words
}

// newPosIndex indexes cs over n content positions.
func newPosIndex(n int, cs []*candidate) *posIndex {
	x := &posIndex{byStart: make([][]*candidate, n+1), byEnd: make([][]*candidate, n+1)}
	for _, c := range cs {
		x.add(c)
	}
	return x
}

func (x *posIndex) add(c *candidate) {
	x.byStart[c.start] = append(x.byStart[c.start], c)
	x.byEnd[c.end] = append(x.byEnd[c.end], c)
	x.maxWords = max(x.maxWords, c.words())
}

// move re-indexes c after its range changed from [start, end).
func (x *posIndex) move(c *candidate, start, end int) {
	if c.start != start {
		x.byStart[start] = slices.DeleteFunc(x.byStart[start], func(o *candidate) bool { return o == c })
		x.byStart[c.start] = append(x.byStart[c.start], c)
	}
	if c.end != end {
		x.byEnd[end] = slices.DeleteFunc(x.byEnd[end], func(o *candidate) bool { return o == c })
		x.byEnd[c.end] = append(x.byEnd[c.end], c)
	}
	x.maxWords = max(x.maxWords, c.words())
}

// startingIn appends to out the candidates whose start lies in [a, b].
func (x *posIndex) startingIn(a, b int, out []*candidate) []*candidate {
	for p := max(a, 0); p <= min(b, len(x.byStart)-1); p++ {
		out = append(out, x.byStart[p]...)
	}
	return out
}

// endingIn appends to out the candidates whose end lies in [a, b].
func (x *posIndex) endingIn(a, b int, out []*candidate) []*candidate {
	for p := max(a, 0); p <= min(b, len(x.byEnd)-1); p++ {
		out = append(out, x.byEnd[p]...)
	}
	return out
}

// overlapping appends to out the candidates that overlap [a, b).
func (x *posIndex) overlapping(a, b int, out []*candidate) []*candidate {
	for p := max(a-x.maxWords+1, 0); p < min(b, len(x.byStart)); p++ {
		for _, c := range x.byStart[p] {
			if c.end > a {
				out = append(out, c)
			}
		}
	}
	return out
}
