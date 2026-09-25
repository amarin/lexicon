package lexicon

import (
	"sync/atomic"

	"github.com/amarin/gomorphy/pkg/morphology"
)

// testHookDictClosed, when set by tests, is called when the last reference of
// a dictionary is released (just before it is closed).
var testHookDictClosed func(name string)

// dict is a registry dictionary: its entry and the opened gomorphy dictionary
// (nil when loading failed). refs counts the registry's own reference plus one
// per snapshot containing it; the dictionary is closed when refs reaches zero,
// i.e. after the registry dropped it and no in-flight Parse uses it.
type dict struct {
	entry Entry
	d     *morphology.Dictionary
	data  []byte // buffer of a built-in opened with OpenBytes (gomorphy does not copy it)
	refs  atomic.Int64
}

func newDict(e Entry) *dict {
	x := &dict{entry: e}
	x.refs.Store(1)

	return x
}

func (x *dict) retain() { x.refs.Add(1) }

// release drops a reference and closes the dictionary on the last one.
func (x *dict) release() error {
	if x.refs.Add(-1) != 0 {
		return nil
	}

	if testHookDictClosed != nil {
		testHookDictClosed(x.entry.Name)
	}

	if x.d == nil {
		return nil
	}

	return x.d.Close()
}
