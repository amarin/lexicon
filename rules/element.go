package rules

// Element is one step of a Pattern: either a Group, or conditions on what
// it consumes. Type consumes one candidate span of that type; without Type
// the element consumes one term. Lemma, Grammeme, Token and Shape must all
// hold for the term (next to Type: for every word of the span). Not is a
// condition that must not match at the same position. Selector strings
// accept alternatives: "a|b".
type Element struct {
	Type     string    `yaml:"type,omitempty"`     // candidate span type(s)
	Lemma    string    `yaml:"lemma,omitempty"`    // lemma or form(s) of a term
	Grammeme string    `yaml:"grammeme,omitempty"` // grammeme(s) in any lemma tag of a term
	Token    string    `yaml:"token,omitempty"`    // token kind(s): word, number, punct, symbol
	Shape    Shape     `yaml:"shape,omitempty"`    // letter case and script of a word
	Not      *Element  `yaml:"not,omitempty"`      // must not match here: type, lemma, grammeme, token, shape only
	Any      bool      `yaml:"any,omitempty"`      // any one term; excludes every other condition
	Group    []Element `yaml:"group,omitempty"`    // nested sequence
	Repeat   Repeat    `yaml:"repeat,omitempty"`   // "", "?", "*", "+" (greedy); quote in YAML
	Role     string    `yaml:"role,omitempty"`     // capture name for actions

	line int // source line, set by Load (decision P11)
}

// hasCondition reports whether the element tests what it consumes (as
// opposed to a pure group or an empty element).
func (e Element) hasCondition() bool {
	return e.Type != "" || e.Lemma != "" || e.Grammeme != "" || e.Token != "" ||
		e.Shape != (Shape{}) || e.Not != nil || e.Any
}
