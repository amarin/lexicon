package ner

// selection is a set of non-conflicting candidates and its total value.
type selection struct {
	value float64
	items []*candidate
}
