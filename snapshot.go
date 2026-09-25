package lexicon

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync/atomic"
)

// snapshot is an immutable set of enabled loaded dictionaries and its version.
// refs counts the registry's "current" reference plus in-flight Parse calls;
// on zero the snapshot releases its dictionaries.
type snapshot struct {
	dicts   []*dict
	version string
	refs    atomic.Int64
}

// newSnapshot takes the enabled loaded dictionaries of all (retaining each);
// the version is sha256 over (name, kind, hash) in order.
func newSnapshot(all []*dict) *snapshot {
	s := &snapshot{}
	h := sha256.New()

	for _, x := range all {
		if !x.entry.Enabled || x.d == nil {
			continue
		}

		x.retain()
		s.dicts = append(s.dicts, x)
		fmt.Fprintf(h, "%s\t%s\t%s\n", x.entry.Name, x.entry.Kind, x.entry.Hash)
	}

	s.version = hex.EncodeToString(h.Sum(nil))
	s.refs.Store(1)

	return s
}

// tryAcquire pins a live snapshot; false when it is already retired.
func (s *snapshot) tryAcquire() bool {
	for {
		n := s.refs.Load()
		if n == 0 {
			return false
		}

		if s.refs.CompareAndSwap(n, n+1) {
			return true
		}
	}
}

// release unpins; the last release releases the dictionaries.
func (s *snapshot) release() error {
	if s.refs.Add(-1) != 0 {
		return nil
	}

	var errs []error
	for _, x := range s.dicts {
		errs = append(errs, x.release())
	}

	return errors.Join(errs...)
}

// parse returns exact readings of the dictionaries of kinds (empty = all), or,
// when there are none, predictions of base dictionaries only. Non-base
// dictionaries are consulted only for known words (gomorphy IsKnown), so their
// predictions are never computed. Strings are copied: cached lemmas outlive
// unmapped dictionaries (D22).
func (s *snapshot) parse(word string, kinds []Kind) []Reading {
	var exact, predicted []Reading

	for _, x := range s.dicts {
		k := x.entry.Kind
		if len(kinds) > 0 && !slices.Contains(kinds, k) {
			continue
		}

		if k != KindBase && !x.d.IsKnown(word) {
			continue
		}

		for _, rd := range x.d.Parse(word) {
			out := Reading{
				Normal: strings.Clone(rd.Normal), Tag: strings.Clone(rd.Tag),
				Kind: k, Dict: x.entry.Name, Predicted: rd.Predicted,
			}

			switch {
			case !rd.Predicted:
				exact = append(exact, out)
			case k == KindBase:
				predicted = append(predicted, out)
			}
		}
	}

	if len(exact) > 0 {
		return exact
	}

	return predicted
}
