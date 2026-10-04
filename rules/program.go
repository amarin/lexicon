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
