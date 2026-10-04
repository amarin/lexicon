package rules

import (
	"errors"
	"fmt"
	"slices"
)

// compiledSet is a validated RuleSet.
type compiledSet struct {
	name     string
	when     []string
	hints    []*HintRule
	triggers []*TriggerRule
	patterns []*Program
}

// compileSet validates rs from file (File.name); errors are *ruleError
// located at the offending rule.
func compileSet(file string, rs *RuleSet) (*compiledSet, error) {
	cs := &compiledSet{name: rs.Name, when: slices.Clone(rs.When)}
	fail := func(line int, rule string, err error) error {
		return &ruleError{file: file, line: line, set: rs.Name, rule: rule, msg: err.Error()}
	}
	for i, h := range rs.Hints {
		hr, err := compileHint(rs.Name, h)
		if err != nil {
			return nil, fail(h.line, fmt.Sprintf("hint %d", i), err)
		}
		cs.hints = append(cs.hints, hr)
	}
	for i, t := range rs.Triggers {
		tr, err := compileTrigger(rs.Name, t)
		if err != nil {
			return nil, fail(t.line, fmt.Sprintf("trigger %d", i), err)
		}
		cs.triggers = append(cs.triggers, tr)
	}
	for i, pt := range rs.Patterns {
		prog, err := compilePattern(rs.Name, pt)
		if err != nil {
			line := pt.line
			var le *lineError
			if errors.As(err, &le) && le.line > 0 {
				line = le.line
			}
			rule := fmt.Sprintf("pattern %q", pt.Name)
			if pt.Name == "" {
				rule = fmt.Sprintf("pattern %d", i)
			}
			return nil, fail(line, rule, err)
		}
		cs.patterns = append(cs.patterns, prog)
	}
	return cs, nil
}

func (cs *compiledSet) activeFor(tags []string) bool {
	for _, w := range cs.when {
		if !slices.Contains(tags, w) {
			return false
		}
	}
	return true
}
