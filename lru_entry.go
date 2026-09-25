package lexicon

// lruEntry is a cached value with its key (list element payload).
type lruEntry struct {
	key string
	val wordResult
}
