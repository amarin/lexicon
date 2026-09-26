package rules

// Trigger is a keyword that proposes a candidate span of Type over the
// following (Dir right) or preceding (Dir left) words that fit Shape, when
// no gazetteer span of that type overlaps them; otherwise it boosts the
// overlapping spans. Negative triggers subtract Weight instead.
type Trigger struct {
	Lemma    string    `yaml:"lemma"`
	Dotted   bool      `yaml:"dotted,omitempty"`
	Type     string    `yaml:"type"`
	Dir      Direction `yaml:"dir,omitempty"`
	Window   Window    `yaml:"window,omitempty"` // default 1..1
	Shape    Shape     `yaml:"shape,omitempty"`
	StopAt   []string  `yaml:"stop_at,omitempty"` // punct (implied), stop, number, latin
	Negative bool      `yaml:"negative,omitempty"`
	Weight   float32   `yaml:"weight,omitempty"` // default 1
	Absorb   bool      `yaml:"absorb,omitempty"`

	line int // source line, set by Load (decision D15)
}
