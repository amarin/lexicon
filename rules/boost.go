package rules

// Boost adds Weight to the span captured by Role and gives it context.
type Boost struct {
	Role   string  `yaml:"role"`
	Weight float32 `yaml:"weight,omitempty"` // default 1
}
