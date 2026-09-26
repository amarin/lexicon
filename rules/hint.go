package rules

// Hint is a keyword («ул.», «деревня», «уезда») that boosts spans of Type
// within Window content words in Dir, satisfies their RequiresContext and,
// with Absorb, extends an adjacent span over the keyword.
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
