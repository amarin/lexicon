package lexicon

import "testing"

func TestLRUEvictsLeastRecentlyUsed(t *testing.T) {
	c := newLRU(2)
	c.put("a", wordResult{stop: true})
	c.put("b", wordResult{})

	if _, ok := c.get("a"); !ok { // a becomes most recent
		t.Fatal("a missing")
	}

	c.put("c", wordResult{})

	if _, ok := c.get("b"); ok {
		t.Fatal("b not evicted")
	}

	if r, ok := c.get("a"); !ok || !r.stop {
		t.Fatal("a evicted")
	}
}
