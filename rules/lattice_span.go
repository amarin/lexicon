package rules

// LatticeSpan is a candidate span starting at some term of a Lattice.
type LatticeSpan struct {
	ID   int    // caller's identifier, reported back in Capture.Span and PatternMatch.Spans
	Type string // span type
	End  int    // exclusive term index within the lattice
}
