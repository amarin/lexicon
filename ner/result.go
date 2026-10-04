package ner

// Result is the output of one Extract call.
type Result struct {
	// Spans are ordered by Start, then longer first; on one range a
	// labelled span comes before its same-range part.
	Spans []Span
	// Facts are what patterns emitted, in match order. The values of
	// Fact.Args index Spans; a fact with an argument that is not in Spans
	// (it lost resolution or the Types filter dropped it) is omitted.
	Facts []Fact
	// Version identifies analyzer, dictionaries, rules and configuration;
	// hosts store it with suggestions to know when to recompute.
	Version string
}
