package gazetteer

// LineError describes one rejected line of a TSV source.
type LineError struct {
	Line int
	Msg  string
}
