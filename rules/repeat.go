package rules

import (
	"fmt"
	"slices"

	"go.yaml.in/yaml/v3"
)

// Repeat is how often an element may match; all repetitions are greedy.
// YAML: a quoted scalar "", "?", "*" or "+" (unquoted * and ? are YAML
// indicators).
type Repeat uint8

const (
	// RepeatOnce: exactly once ("" or omitted; zero value).
	RepeatOnce Repeat = iota
	// RepeatOptional: zero or one ("?").
	RepeatOptional
	// RepeatZeroOrMore: zero or more ("*").
	RepeatZeroOrMore
	// RepeatOneOrMore: one or more ("+").
	RepeatOneOrMore
)

// repeatMarks are the YAML markers, indexed by Repeat.
var repeatMarks = []string{"", "?", "*", "+"}

func (r Repeat) String() string {
	if int(r) < len(repeatMarks) {
		return repeatMarks[r]
	}
	return fmt.Sprintf("Repeat(%d)", uint8(r))
}

// MarshalYAML implements yaml.Marshaler.
func (r Repeat) MarshalYAML() (any, error) { return r.String(), nil }

// UnmarshalYAML implements yaml.Unmarshaler; anything but "", "?", "*" or
// "+" is an error naming the line.
func (r *Repeat) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("line %d: repeat: want \"?\", \"*\" or \"+\"", value.Line)
	}
	i := slices.Index(repeatMarks, value.Value)
	if i < 0 {
		return fmt.Errorf("line %d: unknown repeat %q (want \"?\", \"*\" or \"+\", quoted)", value.Line, value.Value)
	}
	*r = Repeat(i)
	return nil
}
