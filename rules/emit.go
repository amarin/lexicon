package rules

// Emit records a fact of Kind; Args maps fact roles to pattern roles, each
// of which must be a span: bound to a type element, or labelled or
// relabelled earlier in the same pattern.
type Emit struct {
	Kind string            `yaml:"kind"`
	Args map[string]string `yaml:"args"`
}
