package ner

import "math"

// scoreAll computes candidate scores (decision D6). Alternatives already
// attached to a candidate (v0.3 Relabel) are scored too; they may be
// removed candidates. Scores are quantized so that two candidates whose
// formulas are mathematically equal always compare equal, regardless of
// float64 summation order.
func (s *state) scoreAll() {
	for _, c := range s.cands {
		if c.removed {
			continue
		}
		c.score = quantize(s.score(c))
		for _, a := range c.alts {
			a.score = quantize(s.score(a))
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

// quantize rounds x to 1e-6 so equal-by-design scores compare equal exactly.
func quantize(x float64) float64 {
	return math.Round(x*1e6) / 1e6
}
