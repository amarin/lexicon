package textnorm

// TokenKind classifies a token.
type TokenKind uint8

const (
	TokenWord   TokenKind = iota + 1 // letters (one script category), inner hyphens
	TokenNumber                      // ASCII digits
	TokenPunct                       // one punctuation cluster
	TokenSymbol                      // any other visible cluster (№, emoji, …)
)

var tokenKindNames = [...]string{"", "word", "number", "punct", "symbol"}

// String returns "word", "number", "punct" or "symbol".
func (k TokenKind) String() string {
	if int(k) < len(tokenKindNames) {
		return tokenKindNames[k]
	}

	return "unknown"
}
