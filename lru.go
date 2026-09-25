package lexicon

import (
	"container/list"
	"sync"
)

// lru is a bounded least-recently-used map from keys to word results; safe for
// concurrent use.
type lru struct {
	mu  sync.Mutex
	cap int
	ll  *list.List
	m   map[string]*list.Element
}

func newLRU(capacity int) *lru {
	return &lru{cap: capacity, ll: list.New(), m: make(map[string]*list.Element)}
}

func (c *lru) get(key string) (wordResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.m[key]
	if !ok {
		return wordResult{}, false
	}

	c.ll.MoveToFront(e)

	return e.Value.(*lruEntry).val, true
}

func (c *lru) put(key string, v wordResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e, ok := c.m[key]; ok {
		e.Value.(*lruEntry).val = v
		c.ll.MoveToFront(e)

		return
	}

	c.m[key] = c.ll.PushFront(&lruEntry{key: key, val: v})

	if c.ll.Len() > c.cap {
		last := c.ll.Back()
		c.ll.Remove(last)
		delete(c.m, last.Value.(*lruEntry).key)
	}
}
