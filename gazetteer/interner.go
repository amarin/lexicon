package gazetteer

import "sync"

// interner maps strings (lemmas and normalized forms) to uint32 IDs. It is
// append-only and process-wide, so IDs stay valid across snapshots; ID 0 is
// reserved for "never seen" and never appears in a trie.
type interner struct {
	mu   sync.RWMutex
	ids  map[string]uint32
	strs []string
}

// global is shared by every Gazetteer in the process.
var global = newInterner()

func newInterner() *interner {
	return &interner{ids: map[string]uint32{}, strs: []string{""}}
}

// intern returns the ID of s, assigning a new one on first sight.
func (in *interner) intern(s string) uint32 {
	if id := in.lookup(s); id != 0 {
		return id
	}
	in.mu.Lock()
	defer in.mu.Unlock()
	if id, ok := in.ids[s]; ok {
		return id
	}
	id := uint32(len(in.strs))
	in.strs = append(in.strs, s)
	in.ids[s] = id
	return id
}

// lookup returns the ID of s or 0 when s was never interned.
func (in *interner) lookup(s string) uint32 {
	in.mu.RLock()
	id := in.ids[s]
	in.mu.RUnlock()
	return id
}

// text returns the string of id ("" for 0 or unknown IDs).
func (in *interner) text(id uint32) string {
	in.mu.RLock()
	defer in.mu.RUnlock()
	if int(id) >= len(in.strs) {
		return ""
	}
	return in.strs[id]
}
