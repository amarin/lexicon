package rules

// Shape restricts the words a trigger may cover.
type Shape struct {
	Case   string `yaml:"case,omitempty"`   // lower | title | upper
	Script string `yaml:"script,omitempty"` // cyrillic | latin
}
