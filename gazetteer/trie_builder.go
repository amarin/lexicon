package gazetteer

import "slices"

// trieBuilder accumulates keys; freeze flattens it into a trie.
type trieBuilder struct {
	nodes []builderNode
	depth int
}

func newTrieBuilder() *trieBuilder {
	return &trieBuilder{nodes: []builderNode{{}}}
}

// insert adds key → a; inserting the same alias twice under a key is a no-op.
func (b *trieBuilder) insert(key []uint32, a *Alias) {
	n := uint32(0)
	for _, id := range key {
		if b.nodes[n].children == nil {
			b.nodes[n].children = map[uint32]uint32{}
		}
		c, ok := b.nodes[n].children[id]
		if !ok {
			c = uint32(len(b.nodes))
			b.nodes = append(b.nodes, builderNode{})
			b.nodes[n].children[id] = c
		}
		n = c
	}
	if slices.Contains(b.nodes[n].aliases, a) {
		return
	}
	b.nodes[n].aliases = append(b.nodes[n].aliases, a)
	b.depth = max(b.depth, len(key))
}

// freeze returns the flattened, immutable trie.
func (b *trieBuilder) freeze() *trie {
	t := &trie{
		childStart: make([]uint32, len(b.nodes)+1),
		aliasStart: make([]uint32, len(b.nodes)+1),
		depth:      b.depth,
	}
	for i, n := range b.nodes {
		keys := make([]uint32, 0, len(n.children))
		for k := range n.children {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		t.childStart[i] = uint32(len(t.childKeys))
		for _, k := range keys {
			t.childKeys = append(t.childKeys, k)
			t.childNodes = append(t.childNodes, n.children[k])
		}
		t.aliasStart[i] = uint32(len(t.aliases))
		t.aliases = append(t.aliases, n.aliases...)
	}
	t.childStart[len(b.nodes)] = uint32(len(t.childKeys))
	t.aliasStart[len(b.nodes)] = uint32(len(t.aliases))
	return t
}
