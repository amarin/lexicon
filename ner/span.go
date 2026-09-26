package ner

// Span is one resolved mention. Doc.Text[Start:End] == Surface.
type Span struct {
	Start, End         int // bytes in Doc.Text
	RuneStart, RuneEnd int // code points in Doc.Text
	Surface            string
	Type               string
	Normal             []string          // canonical forms; >1 when ambiguous
	Refs               []string          // opaque gazetteer refs; empty for candidates
	Attrs              map[string]string // entry attributes (conflicting keys dropped)
	Flags              SpanFlag
	Score              float32
	Evidence           []string      // only with Explain
	Alternatives       []Alternative // other readings of the same range, best first (D14)
}
