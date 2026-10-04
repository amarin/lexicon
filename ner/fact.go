package ner

// Fact is plain data emitted by a pattern. Args maps a fact role to an
// index in Result.Spans; building domain facts from it is host logic.
type Fact struct {
	Kind string
	Args map[string]int
	Rule string // name of the pattern that emitted it
}
