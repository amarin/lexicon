package rules

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

// compiler turns a Pattern into a Program (Thompson construction).
type compiler struct {
	insts    []inst
	roles    []string
	spanRole map[string]bool // role → bound only to type elements
}

// compilePattern compiles pt. Messages do not repeat the pattern name
// (compileSet prefixes "<file>:<line>: <set>/pattern \"<name>\""); element
// errors are *lineError with the element's source line.
func compilePattern(set string, pt Pattern) (*Program, error) {
	if pt.Name == "" {
		return nil, errors.New("pattern without name")
	}
	if len(pt.Elements) == 0 {
		return nil, errors.New("no elements")
	}
	c := &compiler{spanRole: map[string]bool{}}
	if err := c.seq(pt.Elements); err != nil {
		return nil, err
	}
	c.emit(inst{op: opMatch})
	actions, err := c.actions(pt.Actions)
	if err != nil {
		return nil, err
	}
	return &Program{name: pt.Name, set: set, insts: c.insts, roles: c.roles, actions: actions}, nil
}

func (c *compiler) emit(in inst) int {
	c.insts = append(c.insts, in)
	return len(c.insts) - 1
}

func (c *compiler) slot(role string) int {
	i := slices.Index(c.roles, role)
	if i < 0 {
		c.roles = append(c.roles, role)
		i = len(c.roles) - 1
	}
	return i
}

func (c *compiler) markRole(role string, span bool) {
	if prev, ok := c.spanRole[role]; ok {
		c.spanRole[role] = prev && span
		return
	}
	c.spanRole[role] = span
}

func (c *compiler) seq(es []Element) error {
	for _, e := range es {
		if err := c.elem(e); err != nil {
			return err
		}
	}
	return nil
}

// elem compiles repetition around once(e); greedy: split prefers the body.
func (c *compiler) elem(e Element) error {
	switch e.Repeat {
	case RepeatOnce:
		return c.once(e)
	case RepeatOptional:
		split := c.emit(inst{op: opSplit})
		if err := c.once(e); err != nil {
			return err
		}
		c.insts[split].x, c.insts[split].y = split+1, len(c.insts)
	case RepeatZeroOrMore:
		split := c.emit(inst{op: opSplit})
		if err := c.once(e); err != nil {
			return err
		}
		c.emit(inst{op: opJmp, x: split})
		c.insts[split].x, c.insts[split].y = split+1, len(c.insts)
	case RepeatOneOrMore:
		start := len(c.insts)
		if err := c.once(e); err != nil {
			return err
		}
		split := c.emit(inst{op: opSplit})
		c.insts[split].x, c.insts[split].y = start, split+1
	default: // only a Repeat built in Go; YAML input fails at load
		return atLine(e.line, fmt.Errorf("unknown repeat %s", e.Repeat))
	}
	return nil
}

func (c *compiler) once(e Element) error {
	slot := -1
	if e.Role != "" {
		slot = c.slot(e.Role)
		c.emit(inst{op: opSave, slot: 2 * slot})
	}
	if len(e.Group) > 0 {
		if e.hasCondition() {
			return atLine(e.line, errors.New("an element with a group cannot also have conditions"))
		}
		if e.Role != "" {
			c.markRole(e.Role, false)
		}
		if err := c.seq(e.Group); err != nil {
			return err
		}
	} else {
		sel, err := newSelector(e)
		if err != nil {
			return atLine(e.line, err)
		}
		if e.Role != "" {
			c.markRole(e.Role, sel.spans())
		}
		c.emit(inst{op: opElem, sel: sel, slot: slot})
	}
	if e.Role != "" {
		c.emit(inst{op: opSave, slot: 2*slot + 1})
	}
	return nil
}

// actions validates actions (decisions P9, P12) and applies default weights.
func (c *compiler) actions(as []Action) ([]Action, error) {
	if len(as) == 0 {
		return nil, errors.New("no actions")
	}
	made := map[string]bool{} // roles labelled earlier in this pattern
	known := func(i int, role string) error {
		if _, ok := c.spanRole[role]; !ok {
			return fmt.Errorf("action %d: unknown role %q", i, role)
		}
		return nil
	}
	isSpan := func(role string) bool { return c.spanRole[role] || made[role] }
	out := make([]Action, 0, len(as))
	for i, a := range as {
		n := 0
		for _, set := range []bool{a.Relabel != nil, a.Boost != nil, a.Label != nil, a.Emit != nil} {
			if set {
				n++
			}
		}
		if n != 1 {
			return nil, fmt.Errorf("action %d: want exactly one of relabel, boost, label, emit", i)
		}
		switch {
		case a.Relabel != nil:
			r := *a.Relabel
			if err := known(i, r.Role); err != nil {
				return nil, err
			}
			if !c.spanRole[r.Role] {
				return nil, fmt.Errorf("action %d: relabel role %q is not bound to a type element (use label)", i, r.Role)
			}
			if r.Type == "" {
				return nil, fmt.Errorf("action %d: relabel without type", i)
			}
			if r.Weight == 0 {
				r.Weight = 1
			}
			out = append(out, Action{Relabel: &r})
		case a.Boost != nil:
			b := *a.Boost
			if err := known(i, b.Role); err != nil {
				return nil, err
			}
			if !isSpan(b.Role) {
				return nil, fmt.Errorf("action %d: boost role %q is not a span", i, b.Role)
			}
			if b.Weight == 0 {
				b.Weight = 1
			}
			out = append(out, Action{Boost: &b})
		case a.Label != nil:
			l := *a.Label
			if err := known(i, l.Role); err != nil {
				return nil, err
			}
			if c.spanRole[l.Role] {
				return nil, fmt.Errorf("action %d: label role %q is bound to a type element (use relabel)", i, l.Role)
			}
			if l.Type == "" {
				return nil, fmt.Errorf("action %d: label without type", i)
			}
			if l.Weight == 0 {
				l.Weight = 1
			}
			made[l.Role] = true
			out = append(out, Action{Label: &l})
		default:
			e := *a.Emit
			if e.Kind == "" {
				return nil, fmt.Errorf("action %d: emit without kind", i)
			}
			if len(e.Args) == 0 {
				return nil, fmt.Errorf("action %d: emit without args", i)
			}
			factRoles := make([]string, 0, len(e.Args))
			for k := range e.Args {
				factRoles = append(factRoles, k)
			}
			slices.Sort(factRoles)
			for _, factRole := range factRoles {
				role := e.Args[factRole]
				if err := known(i, role); err != nil {
					return nil, fmt.Errorf("action %d: emit arg %q: unknown role %q", i, factRole, role)
				}
				if !isSpan(role) {
					return nil, fmt.Errorf("action %d: emit arg %q: role %q is not a span (bind it to a type element or label it first)", i, factRole, role)
				}
			}
			e.Args = maps.Clone(e.Args)
			out = append(out, Action{Emit: &e})
		}
	}
	return out, nil
}
