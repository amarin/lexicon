package ner

// Result is the output of one Extract call.
type Result struct {
	Spans []Span // ordered by Start, then longer first, then Type
	// Version identifies analyzer, dictionaries, rules and configuration;
	// hosts store it with suggestions to know when to recompute.
	Version string
}
