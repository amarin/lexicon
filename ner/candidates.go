package ner

import "github.com/amarin/lexicon/gazetteer"

// fromMatches turns gazetteer matches into one candidate per (range, type).
func (s *state) fromMatches(ms []gazetteer.Match) {
	idx := map[rangeKey]*candidate{}
	for _, m := range ms {
		for _, a := range m.Aliases {
			k := rangeKey{m.Start, m.End, a.Entry.Type}
			c := idx[k]
			if c == nil {
				c = &candidate{start: m.Start, end: m.End, typ: a.Entry.Type, origin: originLemma}
				idx[k] = c
				s.cands = append(s.cands, c)
			}
			c.addHit(a, m.Kind)
		}
	}
	for _, c := range s.cands {
		for _, h := range c.hits {
			kind := "lemma"
			if h.surface {
				kind = "surface"
				c.origin = originSurface
			}
			s.note(c, "%s match «%s» (%s)", kind, h.alias.Entry.Alias, h.alias.Source)
		}
	}
}
