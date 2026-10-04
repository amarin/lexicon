package rules

// Label creates a span of Type over the terms captured by Role (a group or
// a non-type element): a part the dictionaries do not know, or a composite
// such as a person over its name parts. The parts of a composite appear in
// the output only for the pairs the host lists in ner.Config.Nesting.
type Label struct {
	Role   string  `yaml:"role"`
	Type   string  `yaml:"type"`
	Weight float32 `yaml:"weight,omitempty"` // default 1
}
