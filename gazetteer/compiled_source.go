package gazetteer

// compiledSource is the immutable compiled form of one source. A source that
// never compiled successfully has built == false and nil tries.
type compiledSource struct {
	name            string
	version         string
	analyzerVersion string
	built           bool
	lemma, surface  *trie
	groups          *groups
	aliases         []*Alias
	report          SourceReport
}
