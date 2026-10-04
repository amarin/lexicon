package ner

import (
	"slices"
	"sort"

	"github.com/amarin/lexicon/rules"
)

// applyPatterns runs every active pattern over every sentence, in book
// order, each on the candidates the previous one left (decision P10). The
// candidates are indexed once per pattern: a pattern's matches never cross
// a sentence, so its actions in one sentence cannot change another.
func (s *state) applyPatterns(progs []*rules.Program) {
	if len(progs) == 0 {
		return
	}
	sentences := s.tx.Sentences()
	for _, prog := range progs {
		starts := s.latticeSpans()
		for _, seg := range sentences {
			l := &lattice{s: s, t0: seg[0], t1: seg[1], starts: starts}
			for _, m := range prog.FindAll(l) {
				s.act(prog, l, m)
			}
		}
	}
}

// act runs the actions of one match in order.
func (s *state) act(prog *rules.Program, l *lattice, m rules.PatternMatch) {
	bound := map[string]*candidate{}
	for role, cp := range m.Roles {
		if cp.Span >= 0 && !s.cands[cp.Span].removed {
			bound[role] = s.cands[cp.Span]
		}
	}
	for _, a := range prog.Actions() {
		switch {
		case a.Relabel != nil:
			if c := bound[a.Relabel.Role]; c != nil {
				bound[a.Relabel.Role] = s.relabel(prog.Name(), c, a.Relabel)
			}
		case a.Boost != nil:
			c := bound[a.Boost.Role]
			if c == nil {
				continue
			}
			c.bonus += float64(a.Boost.Weight)
			c.context = true
			s.note(c, "pattern %s boost %+g", prog.Name(), float64(a.Boost.Weight))
		case a.Label != nil:
			cp, ok := m.Roles[a.Label.Role]
			if !ok {
				continue
			}
			if c := s.label(prog.Name(), l, m, cp, a.Label); c != nil {
				bound[a.Label.Role] = c
			}
		case a.Emit != nil:
			f := pendingFact{kind: a.Emit.Kind, rule: prog.Name(), args: map[string]*candidate{}}
			complete := true
			for factRole, role := range a.Emit.Args {
				c := bound[role]
				if c == nil {
					complete = false
					break
				}
				f.args[factRole] = c
			}
			if complete {
				s.facts = append(s.facts, f)
			}
		}
	}
}

// relabel implements decision P6 for the candidate c a type element
// captured. It returns the candidate now bound to the role.
func (s *state) relabel(rule string, c *candidate, r *rules.Relabel) *candidate {
	w := float64(r.Weight)
	if c.typ == r.Type {
		c.bonus += w
		c.context = true
		s.note(c, "pattern %s keeps %s", rule, r.Type)
		return c
	}
	for _, o := range s.cands {
		if o != c && !o.removed && o.typ == r.Type && o.start == c.start && o.end == c.end {
			// The target type's own gazetteer hits (refs) win; the source
			// reading becomes an alternative.
			c.removed = true
			s.note(c, "relabelled by pattern %s", rule)
			o.alts = append(o.alts, c)
			o.bonus += w
			o.context = true
			s.note(o, "pattern %s relabel %s → %s (existing)", rule, c.typ, r.Type)
			return o
		}
	}
	if s.vetoed(r.Type, c.start, c.end) {
		s.note(c, "pattern %s: %s is blocked here", rule, r.Type)
		return c
	}
	// Keep the previous reading as an alternative: a detached copy (not in
	// s.cands, so it never competes in resolution).
	prev := *c
	prev.evidence = slices.Clone(c.evidence)
	prev.alts = nil
	s.note(&prev, "relabelled by pattern %s", rule)
	c.alts = append(c.alts, &prev)
	s.note(c, "pattern %s relabel %s → %s", rule, c.typ, r.Type)
	c.typ, c.hits, c.normals = r.Type, nil, s.lemmaNormals(c.start, c.end)
	c.flags |= Candidate
	c.bonus += w
	c.context = true
	return c
}

// label implements decision P12: a candidate of lb.Type over the terms cp
// captured. It returns the candidate now bound to the role, or nil when
// the range has no content word or the type is blocked there.
func (s *state) label(rule string, l *lattice, m rules.PatternMatch, cp rules.Capture, lb *rules.Label) *candidate {
	a, b, ok := s.positions(l.t0+cp.Start, l.t0+cp.End)
	if !ok || s.vetoed(lb.Type, a, b) {
		return nil
	}
	w := float64(lb.Weight)
	for _, o := range s.cands {
		if !o.removed && o.typ == lb.Type && o.start == a && o.end == b {
			o.bonus += w
			o.context = true
			s.note(o, "pattern %s keeps %s", rule, lb.Type)
			return o
		}
	}
	nc := &candidate{
		start: a, end: b, typ: lb.Type, origin: originPattern,
		normals: s.lemmaNormals(a, b), flags: Candidate, bonus: w, context: true, labelled: true,
	}
	// A span over candidates the match consumed is a composite of known
	// parts, not a guess.
	for _, id := range m.Spans {
		if c := s.cands[id]; c.start >= a && c.end <= b {
			nc.flags &^= Candidate
			break
		}
	}
	s.note(nc, "labelled by pattern %s", rule)
	s.cands = append(s.cands, nc)
	return nc
}

// positions maps terms [ts, te) to content positions [a, b).
func (s *state) positions(ts, te int) (int, int, bool) {
	n := s.tx.Len()
	a := sort.Search(n, func(p int) bool { return s.tx.TermIndex(p) >= ts })
	b := sort.Search(n, func(p int) bool { return s.tx.TermIndex(p) >= te })
	return a, b, a < b
}

// vetoed: a Blocked alias of typ overlaps positions [a, b) (the rule
// triggers follow, see blockedOverlap).
func (s *state) vetoed(typ string, a, b int) bool {
	for _, c := range s.blocked {
		if c.typ == typ && c.start < b && c.end > a {
			return true
		}
	}
	return false
}
