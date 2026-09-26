package ner

import "github.com/amarin/lexicon/gazetteer"

// hit is one alias supporting a candidate.
type hit struct {
	alias   *gazetteer.Alias
	surface bool // matched by its surface key
}
