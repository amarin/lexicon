package lexicon

// Query is a parsed search query.
type Query struct {
	Terms []Term // complete words (stop words dropped), as ModeIndex
	// Partial is the last word when nothing follows it: searched by prefix
	// over indexed forms and by its lemmas (union); a service word has no
	// lemmas and is prefix-only. nil when the last word is complete.
	Partial *Term
}
