package ner

import (
	"github.com/amarin/lexicon/gazetteer"
	"github.com/amarin/lexicon/rules"
)

// applyHints boosts candidates of each hint's type near its keyword,
// gives them context and optionally absorbs the adjacent keyword.
//
// Only candidates starting or ending within the window of a keyword can
// qualify, so each keyword looks at O(window) index buckets. The effects on
// one candidate do not depend on the other candidates, so visiting them in
// bucket order instead of creation order changes nothing.
func (s *state) applyHints(hs []*rules.HintRule) {
	if len(hs) == 0 {
		return
	}
	idx := s.index()
	var near []*candidate
	for _, h := range hs {
		for k := 0; k < s.tx.Len(); k++ {
			kw := s.term(k)
			if !h.Keyword(kw) {
				continue
			}
			near = near[:0]
			if h.Dir == rules.Right || h.Dir == rules.Both {
				near = idx.startingIn(k+1, k+h.Window, near)
			}
			if h.Dir == rules.Left || h.Dir == rules.Both {
				near = idx.endingIn(k+1-h.Window, k, near)
			}
			for _, c := range near {
				if c.removed || c.typ != h.Type || !within(s.tx, k, c, h.Dir, h.Window) {
					continue
				}
				c.bonus += float64(h.Weight)
				c.context = true
				s.note(c, "hint «%s» → %s %+g", kw.Token.Raw, h.Type, float64(h.Weight))
				if h.Absorb {
					start, end := c.start, c.end
					s.absorb(c, k)
					idx.move(c, start, end)
				}
			}
		}
	}
}

// within reports whether keyword position k lies outside c, within window
// content words in direction dir, with no break in between.
func within(tx *gazetteer.Text, k int, c *candidate, dir rules.Direction, window int) bool {
	if (dir == rules.Right || dir == rules.Both) && k < c.start && c.start-k <= window && tx.Joined(k, c.start) {
		return true
	}
	if (dir == rules.Left || dir == rules.Both) && k >= c.end && k-(c.end-1) <= window && tx.Joined(c.end-1, k) {
		return true
	}
	return false
}

// absorb extends c over the keyword at k when it is directly adjacent and
// records k as absorbed.
func (s *state) absorb(c *candidate, k int) {
	switch {
	case k == c.start-1 && !s.tx.Break(k):
		c.start = k
	case k == c.end && !s.tx.Break(c.end-1):
		c.end = k + 1
	default:
		return
	}
	c.absorbed = append(c.absorbed, k)
}
