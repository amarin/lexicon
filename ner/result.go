package ner

// Result is the output of one Extract call.
type Result struct {
	Spans []Span // ordered by Start, then longer first, then Type
	// Facts are what patterns emitted, for the facts whose argument spans
	// are all in Spans.
	Facts []Fact
	// Version identifies analyzer, dictionaries, rules and configuration;
	// hosts store it with suggestions to know when to recompute.
	Version string
}
