package lexicon

import (
	"sync"
	"sync/atomic"
)

// countingDicts counts Parse calls and has a settable version.
type countingDicts struct {
	fakeDicts
	calls   atomic.Int64
	mu      sync.Mutex
	version string
}

func (c *countingDicts) Parse(word string, kinds []Kind) []Reading {
	c.calls.Add(1)

	return c.fakeDicts.Parse(word, kinds)
}

func (c *countingDicts) Version() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.version
}

func (c *countingDicts) setVersion(v string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.version = v
}
