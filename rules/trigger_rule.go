package rules

import (
	"errors"
	"fmt"

	"github.com/amarin/lexicon"
)

// TriggerRule is a validated Trigger with defaults applied.
type TriggerRule struct {
	Trigger
	Set   string
	key   keyword
	shape shapeCheck
	stops stopSet
}

func compileTrigger(set string, t Trigger) (*TriggerRule, error) {
	if t.Type == "" {
		return nil, errors.New("empty type")
	}
	if t.Dir == Both {
		return nil, errors.New("triggers need dir left or right")
	}
	key, err := newKeyword(t.Lemma, t.Dotted)
	if err != nil {
		return nil, err
	}
	if t.Window == (Window{}) {
		t.Window = Window{1, 1}
	}
	if t.Window.Min < 1 || t.Window.Min > t.Window.Max || t.Window.Max > MaxWindow {
		return nil, fmt.Errorf("window %d..%d must satisfy 1 <= min <= max <= %d", t.Window.Min, t.Window.Max, MaxWindow)
	}
	if t.Weight == 0 {
		t.Weight = 1
	}
	shape, err := compileShape(t.Shape)
	if err != nil {
		return nil, err
	}
	stops, err := compileStops(t.StopAt)
	if err != nil {
		return nil, err
	}
	return &TriggerRule{Trigger: t, Set: set, key: key, shape: shape, stops: stops}, nil
}

// Keyword reports whether t is this trigger's keyword.
func (r *TriggerRule) Keyword(t *lexicon.Term) bool { return r.key.match(t) }

// Accepts reports whether t may be covered by the trigger's candidate.
func (r *TriggerRule) Accepts(t *lexicon.Term) bool { return r.shape.ok(t) && !r.stops.stops(t) }
