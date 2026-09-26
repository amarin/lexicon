package ner

// SpanFlag qualifies a span.
type SpanFlag uint8

const (
	// Ambiguous: several refs/normal forms, a covered ambiguous abbreviation
	// (not absorbed by a rule), or a tie between types. Homonymy of an
	// ordinary covered word does not set it.
	Ambiguous SpanFlag = 1 << iota
	// Predicted: a covered word was known only by predicted lemmas.
	Predicted
	// Abbrev: a covered word is an abbreviation.
	Abbrev
	// Candidate: proposed by a trigger or pattern, not a dictionary record.
	Candidate
	// Nested: inside another output span (Config.Nesting).
	Nested
)

var spanFlagNames = [...]string{"ambiguous", "predicted", "abbrev", "candidate", "nested"}

// Has reports whether every flag in x is set.
func (f SpanFlag) Has(x SpanFlag) bool { return f&x == x }

// Names returns the names of the set flags in declaration order.
func (f SpanFlag) Names() []string {
	var out []string
	for i, n := range spanFlagNames {
		if f&(1<<i) != 0 {
			out = append(out, n)
		}
	}
	return out
}
