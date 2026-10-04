package rules

// Action is exactly one of Relabel, Boost, Label or Emit.
type Action struct {
	Relabel *Relabel `yaml:"relabel,omitempty"`
	Boost   *Boost   `yaml:"boost,omitempty"`
	Label   *Label   `yaml:"label,omitempty"`
	Emit    *Emit    `yaml:"emit,omitempty"`
}
