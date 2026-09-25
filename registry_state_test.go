package lexicon

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// failingState is a StateStore whose calls fail with the given errors.
type failingState struct {
	enabledErr, setErr error
}

func (f failingState) Enabled(context.Context) (map[string]bool, error) {
	return map[string]bool{}, f.enabledErr
}

func (f failingState) SetEnabled(context.Context, string, bool) error { return f.setErr }

// TestSetEnabled (genodex P1): disabling persists, removes the dictionary from
// parsing, changes the version and survives reopening.
func TestSetEnabled(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "surname.test.dat", datBytes(t, surnameForms))

	st := &MemState{}
	r := openTest(t, dir, st)
	v1 := r.Version()

	if err := r.SetEnabled(t.Context(), "surname.test", false); err != nil {
		t.Fatal(err)
	}

	if r.Version() == v1 {
		t.Fatal("version unchanged")
	}

	if got := r.Parse("кузнецов", []Kind{"surname"}); len(got) != 0 {
		t.Fatalf("disabled dictionary parses: %+v", got)
	}

	for _, e := range openTest(t, dir, st).List() {
		if e.Name == "surname.test" && e.Enabled {
			t.Fatal("state lost on reopen")
		}
	}
}

// TestSetEnabledUnknown: unknown name — ErrUnknownDictionary, nothing stored.
func TestSetEnabledUnknown(t *testing.T) {
	st := &MemState{}
	r := openTest(t, "", st)

	if err := r.SetEnabled(t.Context(), "no.such", false); !errors.Is(err, ErrUnknownDictionary) {
		t.Fatalf("err = %v, want ErrUnknownDictionary", err)
	}

	if m, _ := st.Enabled(t.Context()); len(m) != 0 {
		t.Fatalf("state written: %v", m)
	}
}

// TestSetEnabledStateError: a failing store leaves the registry unchanged.
func TestSetEnabledStateError(t *testing.T) {
	r := openTest(t, "", failingState{setErr: errors.New("disk full")})
	v := r.Version()

	if err := r.SetEnabled(t.Context(), "abbrev.test", false); err == nil {
		t.Fatal("no error")
	}

	var enabled bool

	for _, e := range r.List() {
		if e.Name == "abbrev.test" {
			enabled = e.Enabled
		}
	}

	if r.Version() != v || !enabled {
		t.Fatal("registry changed despite the store error")
	}
}

// TestOpenStateError: a failing store fails Open.
func TestOpenStateError(t *testing.T) {
	if _, err := Open(t.Context(), Options{State: failingState{enabledErr: errors.New("db down")}}); err == nil {
		t.Fatal("Open without error")
	}
}

// TestSetEnabledDuplicateName: when a name is shared by a loaded file and a
// duplicate-file error stub (both a "custom.x.dat" and a "custom.x.tsv"),
// SetEnabled requires a loaded entry to exist (ErrUnknownDictionary
// otherwise) but then applies the state to every entry of that name — same
// as load, so List() looks the same whether the state came from SetEnabled
// or from a Reload picking up an externally-changed store (carried from the
// Task 13 review, refined in the Task 14 review).
func TestSetEnabledDuplicateName(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "custom.x.dat", datBytes(t, surnameForms))
	writeFile(t, dir, "custom.x.tsv", []byte("село\tсельцо\tNOUN\n"))

	r := openTest(t, dir, nil)

	if err := r.SetEnabled(t.Context(), "custom.x", false); err != nil {
		t.Fatal(err)
	}

	n := 0

	for _, e := range r.List() {
		if e.Name != "custom.x" {
			continue
		}

		n++

		if e.Enabled {
			t.Fatalf("%+v still enabled", e)
		}
	}

	if n != 2 {
		t.Fatalf("custom.x entries = %d, want 2: %+v", n, r.List())
	}
}

// TestSetEnabledBroken: a name that exists only as a broken entry gets its
// own error, not ErrUnknownDictionary; nothing is stored.
func TestSetEnabledBroken(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "custom.bad.dat", []byte("garbage"))

	st := &MemState{}
	r := openTest(t, dir, st)

	err := r.SetEnabled(t.Context(), "custom.bad", false)
	if err == nil || errors.Is(err, ErrUnknownDictionary) || !strings.Contains(err.Error(), `dictionary "custom.bad" is broken`) {
		t.Fatalf("err = %v", err)
	}

	if m, _ := st.Enabled(t.Context()); len(m) != 0 {
		t.Fatalf("state written: %v", m)
	}
}
