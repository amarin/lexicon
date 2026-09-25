package lexicon

// DefaultCacheSize is the lemma cache size used when AnalyzerOptions.CacheSize is 0.
const DefaultCacheSize = 50000

// AnalyzerOptions configure NewAnalyzer.
type AnalyzerOptions struct {
	// CacheSize bounds the per-analyzer lemma cache (entries keyed by profile
	// name and form); 0 = DefaultCacheSize, negative = no cache.
	CacheSize int
}
