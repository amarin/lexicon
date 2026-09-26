package gazetteer

// Match is one key match: content positions [Start, End) of a Text and the
// aliases whose key ended there. Aliases is shared and read-only.
type Match struct {
	Start, End int
	Kind       MatchKind
	Aliases    []*Alias
}
