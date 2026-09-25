package lexicon

import "sync/atomic"

// lemmaCache keeps one LRU per dictionary-set version: a new version replaces
// the generation, so results computed with old dictionaries are never served
// after a change (late writes go to the dropped generation).
type lemmaCache struct {
	size int
	cur  atomic.Pointer[cacheGen]
}

func newLemmaCache(size int) *lemmaCache { return &lemmaCache{size: size} }

// at returns the LRU of version, starting a new generation on change.
func (c *lemmaCache) at(version string) *lru {
	for {
		g := c.cur.Load()
		if g != nil && g.version == version {
			return g.lru
		}

		ng := &cacheGen{version: version, lru: newLRU(c.size)}
		if c.cur.CompareAndSwap(g, ng) {
			return ng.lru
		}
	}
}
