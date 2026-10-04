package rules

import "slices"

// Program is a compiled Pattern.
type Program struct {
	name    string
	set     string
	insts   []inst
	roles   []string
	actions []Action
}

// Name returns the pattern name.
func (p *Program) Name() string { return p.name }

// Set returns the name of the rule set the pattern belongs to.
func (p *Program) Set() string { return p.set }

// Roles returns the capture roles in order of first appearance.
func (p *Program) Roles() []string { return slices.Clone(p.roles) }

// Actions returns the validated actions with defaults applied (read-only).
func (p *Program) Actions() []Action { return p.actions }

// FindAll returns the non-overlapping, non-empty matches of the program in
// l, scanning start positions left to right.
func (p *Program) FindAll(l Lattice) []PatternMatch {
	n := l.Len()
	m := &matcher{
		insts:   p.insts,
		l:       l,
		n:       n,
		visited: make([]uint32, len(p.insts)*(n+1)),
		caps:    make([]int, 2*len(p.roles)),
		spans:   make([]int, len(p.roles)),
	}
	var out []PatternMatch
	for start := 0; start < n; {
		m.reset()
		end, ok := m.run(0, start)
		if !ok || end == start {
			start++
			continue
		}
		out = append(out, p.match(m, start, end))
		start = end
	}
	return out
}

func (p *Program) match(m *matcher, start, end int) PatternMatch {
	pm := PatternMatch{Start: start, End: end, Roles: map[string]Capture{}, Spans: slices.Clone(m.used)}
	for i, role := range p.roles {
		s, e := m.caps[2*i], m.caps[2*i+1]
		if s < 0 || e < 0 {
			continue
		}
		pm.Roles[role] = Capture{Start: s, End: e, Span: m.spans[i]}
	}
	return pm
}
