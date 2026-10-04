package rules

import "github.com/amarin/lexicon"

// Lattice is what patterns run over: the terms of one sentence and the
// candidate spans starting at each term (longest first, End relative to
// the lattice). ner implements it.
type Lattice interface {
	Len() int
	Term(i int) *lexicon.Term
	SpansAt(i int) []LatticeSpan
}
