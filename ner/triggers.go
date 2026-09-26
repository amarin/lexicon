package ner

import "github.com/amarin/lexicon/rules"

// applyTriggers proposes candidates, or boosts/penalizes overlapping
// gazetteer candidates of the trigger's type (decision D5).
func (s *state) applyTriggers(ts []*rules.TriggerRule) {
	for _, tr := range ts {
		for k := 0; k < s.tx.Len(); k++ {
			kw := s.term(k)
			if !tr.Keyword(kw) {
				continue
			}
			a, b, ok := s.collect(tr, k)
			if !ok {
				continue
			}
			w := float64(tr.Weight)
			boosted := false
			for _, c := range s.cands {
				if c.removed || c.typ != tr.Type || c.origin == originTrigger || c.end <= a || c.start >= b {
					continue
				}
				if tr.Negative {
					c.bonus -= w
					s.note(c, "trigger «%s» → %s %+g", kw.Token.Raw, tr.Type, -w)
					continue
				}
				c.bonus += w
				c.context = true
				boosted = true
				s.note(c, "trigger «%s» → %s %+g", kw.Token.Raw, tr.Type, w)
			}
			if tr.Negative || boosted {
				continue
			}
			nc := &candidate{
				start: a, end: b, typ: tr.Type, origin: originTrigger,
				normals: s.lemmaNormals(a, b), context: true, flags: Candidate,
			}
			s.note(nc, "trigger «%s» → %s candidate", kw.Token.Raw, tr.Type)
			if tr.Absorb {
				s.absorb(nc, k)
			}
			s.cands = append(s.cands, nc)
		}
	}
}

// collect returns the content positions [a, b) the trigger at k covers:
// up to Window.Max accepted, joined words in its direction, at least Min.
func (s *state) collect(tr *rules.TriggerRule, k int) (int, int, bool) {
	count := 0
	if tr.Dir == rules.Left {
		a := k
		for p := k - 1; p >= 0 && count < tr.Window.Max; p-- {
			if s.tx.Break(p) || !tr.Accepts(s.term(p)) {
				break
			}
			a = p
			count++
		}
		return a, k, count >= tr.Window.Min
	}
	b := k + 1
	for p := k + 1; p < s.tx.Len() && count < tr.Window.Max; p++ {
		if s.tx.Break(p-1) || !tr.Accepts(s.term(p)) {
			break
		}
		b = p + 1
		count++
	}
	return k + 1, b, count >= tr.Window.Min
}
