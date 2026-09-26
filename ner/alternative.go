package ner

// Alternative is a reading of a span's range that did not win: another type
// that lost resolution on the same range, or the type a pattern relabelled
// away from (v0.3). Alternatives never make a span Ambiguous by themselves;
// only a tie does (D7).
type Alternative struct {
	Type     string
	Refs     []string          // opaque gazetteer refs; empty for candidates
	Normal   []string          // canonical forms
	Attrs    map[string]string // entry attributes (conflicting keys dropped)
	Score    float32
	Evidence []string // only with Explain
}
