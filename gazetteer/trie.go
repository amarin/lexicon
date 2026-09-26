package gazetteer

// trie is an immutable token trie: the children of node n are
// childKeys[childStart[n]:childStart[n+1]] (sorted) with the matching
// childNodes; the aliases ending at n are aliases[aliasStart[n]:aliasStart[n+1]].
type trie struct {
	childStart []uint32
	childKeys  []uint32
	childNodes []uint32
	aliasStart []uint32
	aliases    []*Alias
	depth      int
}

// child returns the child of node n labelled id (binary search, no allocation).
func (t *trie) child(n, id uint32) (uint32, bool) {
	lo, end := t.childStart[n], t.childStart[n+1]
	hi := end
	for lo < hi {
		mid := (lo + hi) / 2
		if t.childKeys[mid] < id {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if lo < end && t.childKeys[lo] == id {
		return t.childNodes[lo], true
	}
	return 0, false
}

// terminal returns the aliases whose key ends at node n (capacity-limited,
// so appending to it cannot corrupt the trie).
func (t *trie) terminal(n uint32) []*Alias {
	lo, hi := t.aliasStart[n], t.aliasStart[n+1]
	return t.aliases[lo:hi:hi]
}
