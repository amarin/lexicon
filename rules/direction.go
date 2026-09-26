package rules

import (
	"fmt"

	"go.yaml.in/yaml/v3"
)

// Direction says where the affected words are relative to the keyword.
type Direction int

const (
	// Right: words after the keyword (default).
	Right Direction = iota
	// Left: words before the keyword.
	Left
	// Both: either side (hints only).
	Both
)

func (d Direction) String() string {
	switch d {
	case Left:
		return "left"
	case Both:
		return "both"
	default:
		return "right"
	}
}

// MarshalYAML implements yaml.Marshaler.
func (d Direction) MarshalYAML() (any, error) { return d.String(), nil }

// UnmarshalYAML implements yaml.Unmarshaler: a scalar right, left or both
// (empty = right).
func (d *Direction) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("line %d: direction: want left, right or both", value.Line)
	}
	switch value.Value {
	case "", "right":
		*d = Right
	case "left":
		*d = Left
	case "both":
		*d = Both
	default:
		return fmt.Errorf("line %d: unknown direction %q (want left, right or both)", value.Line, value.Value)
	}
	return nil
}
