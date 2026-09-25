package lexicon

import (
	"context"
	"errors"
	"fmt"
)

// SetEnabled enables or disables a dictionary: stores the state, then swaps in
// a new snapshot. Unknown name — ErrUnknownDictionary; a name that exists
// only as broken entries — an error with the entry's error text (not
// ErrUnknownDictionary); a store error leaves the registry unchanged. When
// name collides between a loaded file and a duplicate-file error stub (e.g. a
// "custom.x.dat" and a "custom.x.tsv"), the state applies to every entry
// named name — same as load, so List() looks the same whether the state was
// set here or picked up by a Reload — but the snapshot only ever contains the
// loaded one, since it skips error entries.
//
// A non-nil error from this method after the state store call succeeded (the
// swap or an old dictionary's Close) does not mean the change was rolled
// back: the new state and the new snapshot are already in effect.
func (r *Registry) SetEnabled(ctx context.Context, name string, on bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return ErrClosed
	}

	if !hasLoaded(r.all, name) {
		if e, ok := brokenEntry(r.all, name); ok {
			return fmt.Errorf("lexicon: dictionary %q is broken: %s", name, e)
		}

		return fmt.Errorf("%w: %s", ErrUnknownDictionary, name)
	}

	if err := r.opts.State.SetEnabled(ctx, name, on); err != nil {
		return fmt.Errorf("lexicon: store state of %s: %w", name, err)
	}

	for _, x := range r.all {
		if x.entry.Name == name {
			x.entry.Enabled = on
		}
	}

	return r.swap(newSnapshot(r.all))
}

// hasLoaded reports whether a dictionary named name exists and loaded
// successfully (Error == ""). SetEnabled requires this: toggling a name that
// exists only as a duplicate-file error stub is not meaningful, since the
// snapshot never contains it.
func hasLoaded(all []*dict, name string) bool {
	for _, x := range all {
		if x.entry.Name == name && x.entry.Error == "" {
			return true
		}
	}

	return false
}

// brokenEntry returns the error text of the first entry named name that has
// one; false when there is none.
func brokenEntry(all []*dict, name string) (string, bool) {
	for _, x := range all {
		if x.entry.Name == name && x.entry.Error != "" {
			return x.entry.Error, true
		}
	}

	return "", false
}

// Reload rescans Options.Dir, reopens every dictionary, re-reads the state
// store and swaps in the new snapshot. Dictionaries of the old one are closed
// after in-flight Parse calls finish. On error from load, the current
// snapshot stays; a swap always succeeds once load has, since load only
// fails before opening any dictionary (see its doc comment).
//
// A non-nil error from this method after load succeeded (releasing an old
// generation's dictionaries) does not mean the reload was rolled back: the
// new snapshot is already in effect and the old generation's dictionaries
// are release()d regardless — the error only reports a Close failure on one
// of them.
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
