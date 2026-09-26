package gazetteer

// MatchKind tells which key produced a match.
type MatchKind uint8

const (
	// ByLemma: the lemma sequence of the text matched a lemma key.
	ByLemma MatchKind = iota + 1
	// BySurface: the normalized forms of the text matched a surface key.
	BySurface
)
