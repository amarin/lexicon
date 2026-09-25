package lexicon

import "slices"

// wordResult is the analysis of one form under one profile.
type wordResult struct {
	lemmas []Lemma
	stop   bool // every exact reading is a service word
}

// clone copies the lemma slice so callers may modify it.
func (r wordResult) clone() wordResult {
	return wordResult{lemmas: slices.Clone(r.lemmas), stop: r.stop}
}
