package ner

// pendingFact is a fact recorded by a pattern before resolution.
type pendingFact struct {
	kind, rule string
	args       map[string]*candidate
}
