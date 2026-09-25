package lexicon

// cacheGen is the lemma cache of one dictionary-set version.
type cacheGen struct {
	version string
	lru     *lru
}
