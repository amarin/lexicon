package lexicon

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

// Registry holds every dictionary (built-in and from Options.Dir), their
// state and the current snapshot of enabled ones. Parse reads the snapshot
// without locks; state changes and reloads build a new snapshot and swap it
// atomically. Safe for concurrent use.
type Registry struct {
	opts   Options
	mu     sync.Mutex // serialises List, Summary, SetEnabled, Reload, Close
	all    []*dict    // registry order; each holds one registry reference
	closed bool
	cur    atomic.Pointer[snapshot]
}

var _ Dictionaries = (*Registry)(nil)

// Open loads the dictionaries of o and applies the stored state. Broken files
// and files of undeclared kinds never fail Open: they are listed with
// Entry.Error. Errors come from the state store, from reading the directory,
// from an invalid Options.Kinds entry and from a built-in whose kind is not
// declared (a host programming error, owner decision 7).
func Open(ctx context.Context, o Options) (*Registry, error) {
	if o.State == nil {
		o.State = &MemState{}
	}

	kinds, err := newKindSet(o.Kinds)
	if err != nil {
		return nil, err
	}

	for _, b := range o.Builtin {
		if k, err := kindOf(b.Name); err == nil {
			if err := kinds.check(k); err != nil {
				return nil, fmt.Errorf("lexicon: built-in %q: %w", b.Name, err)
			}
		}
	}

	all, err := load(ctx, o)
	if err != nil {
		return nil, err
	}

	r := &Registry{opts: o, all: all}
	r.cur.Store(newSnapshot(all))

	return r, nil
}

// acquire pins the current snapshot.
func (r *Registry) acquire() *snapshot {
	for {
		if s := r.cur.Load(); s.tryAcquire() {
			return s
		}
	}
}

// swap installs s and releases the registry's reference to the old snapshot.
func (r *Registry) swap(s *snapshot) error {
	return r.cur.Swap(s).release()
}

// Parse returns the readings of word from enabled dictionaries of kinds
// (empty = all) in registry order: exact readings, or — when there are none —
// predictions of base dictionaries only.
func (r *Registry) Parse(word string, kinds []Kind) []Reading {
	s := r.acquire()
	defer s.release()

	return s.parse(word, kinds)
}

// Version identifies the set of enabled dictionaries and their contents.
func (r *Registry) Version() string { return r.cur.Load().version }

// List returns copies of all entries in registry order.
func (r *Registry) List() []Entry {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]Entry, 0, len(r.all))
	for _, x := range r.all {
		out = append(out, x.entry)
	}

	return out
}

// Summary is a one-line state for host logs.
func (r *Registry) Summary() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	on, broken, base := 0, 0, "no"

	for _, x := range r.all {
		if x.entry.Error != "" {
			broken++
		}

		if x.entry.Enabled && x.d != nil {
			on++

			if x.entry.Kind == KindBase {
				base = "yes"
			}
		}
	}

	return fmt.Sprintf("dictionaries: %d of %d enabled; base: %s; broken: %d", on, len(r.all), base, broken)
}

// Close releases the dictionaries once in-flight Parse calls finish. After
// Close, Parse returns nil, List is empty; Close is idempotent.
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true
	errs := []error{r.swap(newSnapshot(nil))}

	for _, x := range r.all {
		errs = append(errs, x.release())
	}

	r.all = nil

	return errors.Join(errs...)
}
