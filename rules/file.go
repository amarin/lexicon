package rules

// File is one rule file: a provenance header (top-level meta mapping) and
// rule sets.
type File struct {
	Meta map[string]string `yaml:"meta,omitempty"` // source, license, version, url, generated_from
	Sets []RuleSet         `yaml:"sets"`

	name string // file name given to LoadNamed/LoadFile; "" = "<input>"
}
