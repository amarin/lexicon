package textnorm

// runePos is a code point of the input and its byte offset; its index in the
// tokenizer slice is its code-point offset.
type runePos struct {
	off int
	r   rune
	cyr bool // letter of the Cyrillic category (tokenizer.classify)
}
