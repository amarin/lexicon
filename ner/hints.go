package ner

import (
	"github.com/amarin/lexicon/gazetteer"
	"github.com/amarin/lexicon/rules"
)

// applyHints boosts candidates of each hint's type near its keyword,
// gives them context and optionally absorbs the adjacent keyword.
func (s *state) applyHints(hs []*rules.HintRule) {
	for _, h := range hs {
		for k := 0; k < s.tx.Len(); k++ {
			kw := s.term(k)
			if !h.Keyword(kw) {
				continue
			}
			for _, c := range s.cands {
				if c.removed || c.typ != h.Type || !within(s.tx, k, c, h.Dir, h.Window) {
					continue
				}
				c.bonus += float64(h.Weight)
				c.context = true
				s.note(c, "hint «%s» → %s %+g", kw.Token.Raw, h.Type, float64(h.Weight))
				if h.Absorb {
					s.absorb(c, k)
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

// absorb extends c over the keyword at k when it is directly adjacent.
func (s *state) absorb(c *candidate, k int) {
	switch {
	case k == c.start-1 && !s.tx.Break(k):
		c.start = k
	case k == c.end && !s.tx.Break(c.end-1):
		c.end = k + 1
	}
}
