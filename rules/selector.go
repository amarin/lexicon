package rules

import (
	"errors"
	"fmt"
	"slices"
)

// selector is the compiled test of one non-group element (decision P14).
// With types it consumes a candidate span whose words satisfy cond;
// without, one term that satisfies cond. not must not match at the same
// position.
type selector struct {
	types []string
	cond  termCond
	not   *selector
}

func newSelector(e Element) (*selector, error) {
	if e.Any {
		if e.Type != "" || e.Lemma != "" || e.Grammeme != "" || e.Token != "" || e.Shape != (Shape{}) || e.Not != nil {
			return nil, errors.New("any cannot be combined with other conditions")
		}
		return &selector{}, nil
	}
	if !e.hasCondition() {
		return nil, errors.New("an element needs type, lemma, grammeme, token, shape, not, any or group")
	}
	s := &selector{}
	if e.Type != "" {
		if s.types = splitAlt(e.Type); len(s.types) == 0 {
			return nil, fmt.Errorf("empty type %q", e.Type)
		}
	}
	cond, err := newTermCond(e)
	if err != nil {
		return nil, err
	}
	s.cond = cond
	if e.Not != nil {
		n := *e.Not
		if n.Role != "" || n.Repeat != RepeatOnce || len(n.Group) > 0 || n.Any || n.Not != nil {
			return nil, atLine(n.line, errors.New("not takes only type, lemma, grammeme, token and shape"))
		}
		if !n.hasCondition() {
			return nil, atLine(n.line, errors.New("empty not"))
		}
		ns, err := newSelector(n)
		if err != nil {
			return nil, atLine(n.line, err)
		}
		s.not = ns
	}
	return s, nil
}

// spans reports whether the selector consumes candidate spans.
func (s *selector) spans() bool { return s.types != nil }

func (s *selector) matchSpan(sp LatticeSpan) bool { return slices.Contains(s.types, sp.Type) }
