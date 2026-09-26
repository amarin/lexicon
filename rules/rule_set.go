package rules

// RuleSet groups rules that apply together. It is active for a document
// when every tag in When is among the document tags (empty When: always).
type RuleSet struct {
	Name     string    `yaml:"name"`
	When     []string  `yaml:"when,omitempty"`
	Hints    []Hint    `yaml:"hints,omitempty"`
	Triggers []Trigger `yaml:"triggers,omitempty"`

	line int // source line of the set, set by Load (decision D15)
}
