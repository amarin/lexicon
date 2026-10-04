package ner

import (
	"cmp"
	"slices"
	"sort"
	"strings"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/textnorm"
)

// span converts a candidate into an output span.
func (s *state) span(text string, c *candidate) Span {
	t0 := s.term(c.start).Token
	t1 := s.term(c.end - 1).Token
	end, runeEnd := t1.End, t1.RuneEnd
	if dot := s.abbrevDot(c); dot != nil {
		end, runeEnd = dot.End, dot.RuneEnd
	}
	sp := Span{
		Start: t0.Start, End: end,
		RuneStart: t0.RuneStart, RuneEnd: runeEnd,
		Surface:      text[t0.Start:end],
		Type:         c.typ,
		Normal:       c.normalForms(),
		Refs:         c.refs(),
		Attrs:        c.attrs(),
		Flags:        c.flags,
		Score:        float32(c.score),
		Alternatives: s.alternatives(c),
	}
	if c.labelled && len(c.parts) > 0 {
		sp.Normal = []string{s.composeNormal(c)}
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
		// Of the covered words, only an ambiguous abbreviation («с.»: село
		// or сын) makes the span ambiguous; ordinary homonymy of a word
		// («стали») is resolved by the alias match itself. A keyword the
		// rule absorbed was disambiguated by that rule (an «с.» a village
		// hint absorbed reads as «село»); it still marks the span Abbrev.
		if ambiguousAbbrev(t) && !slices.Contains(c.absorbed, p) {
			sp.Flags |= Ambiguous
		}
		if c.origin != originSurface && onlyPredicted(t) {
			sp.Flags |= Predicted
		}
	}
	return sp
}

// abbrevDot returns the '.' token that belongs to c's last word: the word
// is immediately followed by a dot and is an abbreviation — by a lemma, or
// because a matched alias is written with a trailing dot (decision P16).
// A sentence dot after an ordinary word is not part of the span.
func (s *state) abbrevDot(c *candidate) *textnorm.Token {
	i := s.tx.TermIndex(c.end - 1)
	last := &s.terms[i]
	if !last.Token.Dotted || i+1 >= len(s.terms) {
		return nil
	}
	dot := &s.terms[i+1].Token
	if dot.Kind != textnorm.TokenPunct || dot.Raw != "." {
		return nil
	}
	if hasLemmaFlag(last, lexicon.FlagAbbrev) {
		return dot
	}
	for _, h := range c.hits {
		if strings.HasSuffix(h.alias.Entry.Alias, ".") {
			return dot
		}
	}
	return nil
}

// composeNormal builds the normal form of a labelled span from the spans
// nested in it: each part gives its first normal form, every other word
// its first lemma (decision P12).
func (s *state) composeNormal(c *candidate) string {
	var words []string
	parts := c.parts
	for p := c.start; p < c.end; {
		for len(parts) > 0 && parts[0].start < p {
			parts = parts[1:]
		}
		if len(parts) > 0 && parts[0].start == p {
			if n := parts[0].normalForms(); len(n) > 0 {
				words = append(words, n[0])
			}
			p = parts[0].end
			parts = parts[1:]
			continue
		}
		if n := s.lemmaNormals(p, p+1); len(n) > 0 {
			words = append(words, n[0])
		}
		p++
	}
	return strings.Join(words, " ")
}

// output orders spans (Start asc, End desc, a labelled span before its
// same-range part, Type), applies the Types filter
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
		// One range: a labelled span is the outer one of its nested part.
		if lx, ly := chosen[order[a]].labelled, chosen[order[b]].labelled; lx != ly {
			return lx
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

// ambiguousAbbrev: t has an abbreviation lemma that is ambiguous (several
// expansions, or a one-letter dotted abbreviation).
func ambiguousAbbrev(t *lexicon.Term) bool {
	const f = lexicon.FlagAbbrev | lexicon.FlagAmbiguous
	for _, l := range t.Lemmas {
		if l.Flags&f == f {
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
