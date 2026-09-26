package rules

import (
	"fmt"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Window is an inclusive range of content words. YAML: a scalar — a number
// N (1..N), "min..max" or a quoted "N".
type Window struct{ Min, Max int }

// MarshalYAML implements yaml.Marshaler.
func (w Window) MarshalYAML() (any, error) { return fmt.Sprintf("%d..%d", w.Min, w.Max), nil }

// UnmarshalYAML implements yaml.Unmarshaler.
func (w *Window) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("line %d: window: want a number or \"min..max\"", value.Line)
	}
	s := strings.TrimSpace(value.Value)
	lo, hi, ok := strings.Cut(s, "..")
	if !ok {
		n, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("line %d: window: bad value %q", value.Line, s)
		}
		*w = Window{1, n}
		return nil
	}
	a, err1 := strconv.Atoi(strings.TrimSpace(lo))
	c, err2 := strconv.Atoi(strings.TrimSpace(hi))
	if err1 != nil || err2 != nil {
		return fmt.Errorf("line %d: window: bad range %q", value.Line, s)
	}
	*w = Window{a, c}
	return nil
}
