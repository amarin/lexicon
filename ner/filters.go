package ner

import (
	"unicode/utf8"

	"github.com/amarin/lexicon/gazetteer"
)

// filterEarly applies the filters that need no rule evidence.
func (s *state) filterEarly() {
	for _, c := range s.cands {
		if c.removed {
			continue
		}
		kept := c.hits[:0]
		blocked := false
		for _, h := range c.hits {
			e := &h.alias.Entry
			switch {
			case e.Flags.Has(gazetteer.Blocked):
				blocked = true
			case e.Flags.Has(gazetteer.CaseSensitive) && !s.caseMatches(c, h.alias):
			case !h.surface && c.words() == 1 && utf8.RuneCountInString(h.alias.SurfaceKey) < s.p.minRunes:
			default:
				kept = append(kept, h)
			}
		}
		c.hits = kept
		if blocked {
			// A blocked alias means "these words are not of this type": remember
			// the range so a later trigger does not propose a fresh candidate of
			// the same, vetoed type over it.
			s.blocked = append(s.blocked, rangeKey{c.start, c.end, c.typ})
		}
		if blocked || len(kept) == 0 {
			c.removed = true
			continue
		}
		c.recomputeOrigin()
	}
}

// blockedOverlap reports whether [a, b) overlaps a filterEarly-vetoed range
// of typ.
func (s *state) blockedOverlap(typ string, a, b int) bool {
	for _, rk := range s.blocked {
		if rk.typ == typ && rk.start < b && a < rk.end {
			return true
		}
	}
	return false
}

// caseMatches compares the letter case of every covered word with the alias.
func (s *state) caseMatches(c *candidate, a *gazetteer.Alias) bool {
	if len(a.Cases) != c.words() {
		return false
	}
	for i, want := range a.Cases {
		if s.term(c.start+i).Token.Case != want {
			return false
		}
	}
	return true
}

// filterContext drops RequiresContext hits of candidates no rule supported.
func (s *state) filterContext() {
	for _, c := range s.cands {
		if c.removed || c.context || len(c.hits) == 0 {
			continue
		}
		kept := c.hits[:0]
		for _, h := range c.hits {
			if !h.alias.Entry.Flags.Has(gazetteer.RequiresContext) {
				kept = append(kept, h)
			}
		}
		c.hits = kept
		if len(kept) == 0 {
			c.removed = true
			continue
		}
		c.recomputeOrigin()
	}
}
