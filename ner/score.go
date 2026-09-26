package ner

// scoreAll computes candidate scores (decision D6). Alternatives already
// attached to a candidate (v0.3 Relabel) are scored too; they may be
// removed candidates.
func (s *state) scoreAll() {
	for _, c := range s.cands {
		if c.removed {
			continue
		}
		c.score = s.score(c)
		for _, a := range c.alts {
			a.score = s.score(a)
		}
	}
}

// score is formula D6 for one candidate.
func (s *state) score(c *candidate) float64 {
	w := s.p.weights
	words := float64(c.words())
	per := c.origin.weight(w) + float64(w.Types[c.typ])
	return per*words + float64(w.LengthBonus)*(words-1) + c.bonus -
		float64(w.AmbiguityPenalty)*float64(c.readings()-1)
}
