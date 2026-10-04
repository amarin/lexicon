package rules

// Capture is what a role consumed: terms [Start, End) of the lattice and,
// for a type element, the LatticeSpan.ID (otherwise -1).
type Capture struct {
	Start, End int
	Span       int
}
