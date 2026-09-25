package lexicon

import (
	"context"
	"errors"
	"fmt"
)

// SetEnabled enables or disables a dictionary: stores the state, then swaps in
// a new snapshot. Unknown name — ErrUnknownDictionary; a store error leaves
// the registry unchanged.
func (r *Registry) SetEnabled(ctx context.Context, name string, on bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return ErrClosed
	}

	i := indexByName(r.all, name)
	if i < 0 {
		return fmt.Errorf("%w: %s", ErrUnknownDictionary, name)
	}

	if err := r.opts.State.SetEnabled(ctx, name, on); err != nil {
		return fmt.Errorf("lexicon: store state of %s: %w", name, err)
	}

	r.all[i].entry.Enabled = on

	return r.swap(newSnapshot(r.all))
}

// indexByName returns the index in all of the dictionary named name: the
// loaded (Error == "") entry when one exists among same-named duplicates
// (e.g. a "custom.x.dat" and a "custom.x.tsv" collide on "custom.x"), else
// the first same-named entry. -1 when none matches.
func indexByName(all []*dict, name string) int {
	first := -1

	for i, x := range all {
		if x.entry.Name != name {
			continue
		}

		if x.entry.Error == "" {
			return i
		}

		if first < 0 {
			first = i
		}
	}

	return first
}

// Reload rescans Options.Dir, reopens every dictionary, re-reads the state
// store and swaps in the new snapshot. Dictionaries of the old one are closed
// after in-flight Parse calls finish. On error the current snapshot stays.
//
// Replace dictionary files atomically (write a temporary file, rename):
// overwriting a memory-mapped file in place can crash in-flight parses.
func (r *Registry) Reload(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return ErrClosed
	}

	all, err := load(ctx, r.opts)
	if err != nil {
		return err
	}

	old := r.all
	r.all = all
	errs := []error{r.swap(newSnapshot(all))}

	for _, x := range old {
		errs = append(errs, x.release())
	}

	return errors.Join(errs...)
}
