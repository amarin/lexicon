package ner

// rangeKey groups gazetteer matches into candidates.
type rangeKey struct {
	start, end int
	typ        string
}
