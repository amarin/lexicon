package ner

import (
	"cmp"
	"slices"
	"sort"
	"strings"

	"github.com/amarin/lexicon"
)

// span converts a candidate into an output span.
func (s *state) span(text string, c *candidate) Span {
	t0 := s.term(c.start).Token
	t1 := s.term(c.end - 1).Token
	sp := Span{
		Start: t0.Start, End: t1.End,
		RuneStart: t0.RuneStart, RuneEnd: t1.RuneEnd,
		Surface:      text[t0.Start:t1.End],
		Type:         c.typ,
		Normal:       c.normalForms(),
		Refs:         c.refs(),
		Attrs:        c.attrs(),
		Flags:        c.flags,
		Score:        float32(c.score),
		Alternatives: s.alternatives(c),
	}
	if s.explain {
		sp.Evidence = c.evidence
	}
	if len(sp.Refs) > 1 || len(sp.Normal) > 1 {
		sp.Flags |= Ambiguous
	}
	for p := c.start; p < c.end; p++ {
		t := s.term(p)
		if hasLemmaFlag(t, lexicon.FlagAbbrev) {
			sp.Flags |= Abbrev
		}
		// A keyword the rule absorbed was disambiguated by that rule (an
		// «с.» a village hint absorbed reads as «село»), so it is not a
		// source of ambiguity; it still marks the span as abbreviated.
		if hasLemmaFlag(t, lexicon.FlagAmbiguous) && !slices.Contains(c.absorbed, p) {
			sp.Flags |= Ambiguous
		}
		if c.origin != originSurface && onlyPredicted(t) {
			sp.Flags |= Predicted
		}
	}
	return sp
}

// output orders spans (Start asc, End desc, Type), applies the Types filter
// and returns the index of every kept candidate in the result. Extract
// ignores the index map for now; it is reserved for v0.3 facts, which
// refer to spans by their position in Result.Spans.
func (s *state) output(text string, chosen []*candidate, nested map[*candidate]bool, types []string) ([]Span, map[*candidate]int) {
	all := make([]Span, len(chosen))
	for i, c := range chosen {
		all[i] = s.span(text, c)
		if nested[c] {
			all[i].Flags |= Nested
		}
	}
	order := make([]int, len(chosen))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool {
		x, y := &all[order[a]], &all[order[b]]
		if x.Start != y.Start {
			return x.Start < y.Start
		}
		if x.End != y.End {
			return x.End > y.End
		}
		return x.Type < y.Type
	})
	var spans []Span
	index := map[*candidate]int{}
	for _, i := range order {
		if len(types) > 0 && !slices.Contains(types, all[i].Type) {
			continue
		}
		index[chosen[i]] = len(spans)
		spans = append(spans, all[i])
	}
	return spans, index
}

func hasLemmaFlag(t *lexicon.Term, f lexicon.Flag) bool {
	for _, l := range t.Lemmas {
		if l.Flags&f != 0 {
			return true
		}
	}
	return false
}

// onlyPredicted: t has known lemmas and all of them are predicted.
func onlyPredicted(t *lexicon.Term) bool {
	known := 0
	for _, l := range t.Lemmas {
		if l.Flags&lexicon.FlagUnknown != 0 {
			continue
		}
		if l.Flags&lexicon.FlagPredicted == 0 {
			return false
		}
		known++
	}
	return known > 0
}

// alternatives converts c.alts, best first: score desc, type asc, then
// attachment order (decision D14).
func (s *state) alternatives(c *candidate) []Alternative {
	if len(c.alts) == 0 {
		return nil
	}
	alts := slices.Clone(c.alts)
	slices.SortStableFunc(alts, func(a, b *candidate) int {
		if a.score != b.score {
			return cmp.Compare(b.score, a.score)
		}
		return strings.Compare(a.typ, b.typ)
	})
	out := make([]Alternative, len(alts))
	for i, a := range alts {
		out[i] = Alternative{
			Type: a.typ, Refs: a.refs(), Normal: a.normalForms(), Attrs: a.attrs(),
			Score: float32(a.score),
		}
		if s.explain {
			out[i].Evidence = a.evidence
		}
	}
	return out
}
