package lexicon

import "errors"

var (
	// ErrUnknownDictionary: no dictionary with that name.
	ErrUnknownDictionary = errors.New("lexicon: unknown dictionary")
	// ErrClosed: the registry is closed.
	ErrClosed = errors.New("lexicon: registry is closed")
)
