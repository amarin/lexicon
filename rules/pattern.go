package rules

// Pattern is a sequence over the terms and candidate spans of one
// sentence; when it matches, its actions run in order.
type Pattern struct {
	Name     string    `yaml:"name"`
	Elements []Element `yaml:"elements"`
	Actions  []Action  `yaml:"actions"`

	line int // source line, set by Load (decision P11)
}
