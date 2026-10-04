package rules

// Relabel changes the type of the span captured by Role (a role bound to a
// type element). The reading it replaces stays in the span's alternatives.
type Relabel struct {
	Role   string  `yaml:"role"`
	Type   string  `yaml:"type"`
	Weight float32 `yaml:"weight,omitempty"` // default 1
}
