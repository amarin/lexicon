package lexicon

import (
	"context"
	"maps"
	"sync"
)

// MemState is an in-memory StateStore; the zero value is ready to use.
type MemState struct {
	mu sync.Mutex
	m  map[string]bool
}

// Enabled returns a copy of the stored states.
func (s *MemState) Enabled(context.Context) (map[string]bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := maps.Clone(s.m)
	if out == nil {
		out = map[string]bool{}
	}

	return out, nil
}

// SetEnabled stores the state of a dictionary.
func (s *MemState) SetEnabled(_ context.Context, name string, on bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.m == nil {
		s.m = map[string]bool{}
	}

	s.m[name] = on

	return nil
}
