package gazetteer

// builderNode is a mutable trie node used while compiling a source.
type builderNode struct {
	children map[uint32]uint32
	aliases  []*Alias
}
