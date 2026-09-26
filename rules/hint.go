package rules

// Hint is a keyword («ул.», «деревня», «уезда») that boosts spans of Type
// within Window content words in Dir, satisfies their RequiresContext and,
// with Absorb, extends an adjacent span over the keyword. The window is
// measured from the keyword to the span's current edge, so after one hint
// absorbed a keyword, a later hint measures from the absorbed edge: results
// may depend on the order of hints. Hints have no shape.
type Hint struct {
	Lemma  string    `yaml:"lemma"`            // lemma or form; alternatives with "|"
	Dotted bool      `yaml:"dotted,omitempty"` // keyword must be followed by '.'
	Type   string    `yaml:"type"`
	Dir    Direction `yaml:"dir,omitempty"`
	Window int       `yaml:"window,omitempty"` // default 1
	Weight float32   `yaml:"weight,omitempty"` // default 1
	Absorb bool      `yaml:"absorb,omitempty"`

	line int // source line, set by Load (decision D15); 0 when built in Go
}
