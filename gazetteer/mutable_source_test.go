package gazetteer

import (
	"context"
	"sync"
)

// mutableSource lets tests change entries, version and failures.
type mutableSource struct {
	name    string
	mu      sync.Mutex
	version string
	entries []Entry
	fail    error
}

func (m *mutableSource) Name() string { return m.name }

func (m *mutableSource) Version(context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.version, nil
}

func (m *mutableSource) Entries(ctx context.Context, yield func(Entry) error) error {
	m.mu.Lock()
	entries, fail := m.entries, m.fail
	m.mu.Unlock()
	if fail != nil {
		return fail
	}
	return NewSliceSource(m.name, "", entries).Entries(ctx, yield)
}

func (m *mutableSource) set(version string, fail error, entries ...Entry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.version, m.fail = version, fail
	if entries != nil {
		m.entries = entries
	}
}
