package lexicon

import (
	"errors"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"
)

// TestReloadPicksUpChanges: new files and externally changed state apply.
func TestReloadPicksUpChanges(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "surname.test.dat", datBytes(t, surnameForms))

	st := &MemState{}
	r := openTest(t, dir, st)
	v1 := r.Version()

	writeFile(t, dir, "toponym.x.tsv", []byte("москва\tмоскве\tNOUN,Geox\n"))

	if err := st.SetEnabled(t.Context(), "surname.test", false); err != nil { // e.g. a CLI changed it
		t.Fatal(err)
	}

	if err := r.Reload(t.Context()); err != nil {
		t.Fatal(err)
	}

	if got := names(r); got != "base.builtin:base|abbrev.test:abbrev|surname.test:surname|toponym.x:toponym" {
		t.Fatalf("List = %s", got)
	}

	if r.Version() == v1 || len(r.Parse("кузнецов", []Kind{"surname"})) != 0 || len(exact(r.Parse("москве", nil))) != 1 {
		t.Fatal("reload did not apply files and state")
	}
}

// TestReloadErrorKeepsSnapshot: a failed reload changes nothing.
func TestReloadErrorKeepsSnapshot(t *testing.T) {
	dir := t.TempDir()
	r := openTest(t, dir, nil)
	v := r.Version()

	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(dir, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := r.Reload(t.Context()); err == nil {
		t.Fatal("Reload without error")
	}

	if r.Version() != v || len(exact(r.Parse("кота", nil))) != 1 {
		t.Fatal("snapshot changed after a failed reload")
	}
}

// TestReloadRetiresAfterInFlight: dictionaries of a retired snapshot close
// only after the last in-flight user releases it.
func TestReloadRetiresAfterInFlight(t *testing.T) {
	var (
		mu     sync.Mutex
		closed []string
	)

	testHookDictClosed = func(name string) {
		mu.Lock()
		closed = append(closed, name)
		mu.Unlock()
	}
	t.Cleanup(func() { testHookDictClosed = nil })

	dir := t.TempDir()
	writeFile(t, dir, "surname.test.dat", datBytes(t, surnameForms))

	r := openTest(t, dir, nil)
	s := r.acquire() // an in-flight Parse

	if err := r.Reload(t.Context()); err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	n := len(closed)
	mu.Unlock()

	if n != 0 {
		t.Fatalf("closed while pinned: %v", closed)
	}

	if got := exact(s.parse("кузнецов", []Kind{"surname"})); len(got) != 1 {
		t.Fatalf("pinned snapshot parse = %+v", got)
	}

	if err := s.release(); err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	defer mu.Unlock()

	slices.Sort(closed)

	if got := strings.Join(closed, "|"); got != "abbrev.test|base.builtin|surname.test" {
		t.Fatalf("closed after release: %s", got)
	}
}

// TestCloseAndReloadReleaseEveryDictionary: a snapshot pinned across a Reload
// and released only after Close must not leak its generation's dictionaries,
// and Close must not double-release the generation Reload opened. Counting
// hook calls (not just names, which repeat across generations) catches both
// a leak (too few) and a double release (too many).
func TestCloseAndReloadReleaseEveryDictionary(t *testing.T) {
	var (
		mu     sync.Mutex
		closed []string
	)

	testHookDictClosed = func(name string) {
		mu.Lock()
		closed = append(closed, name)
		mu.Unlock()
	}
	t.Cleanup(func() { testHookDictClosed = nil })

	dir := t.TempDir()
	writeFile(t, dir, "surname.test.dat", datBytes(t, surnameForms))

	r := openTest(t, dir, nil)

	var names1 []string
	for _, e := range r.List() {
		names1 = append(names1, e.Name)
	}

	s := r.acquire() // pin generation 1, past the Reload below

	if err := r.Reload(t.Context()); err != nil { // opens generation 2
		t.Fatal(err)
	}

	if err := r.Close(); err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	afterClose := slices.Clone(closed)
	mu.Unlock()

	// Close releases the registry's own reference to generation 2, which
	// nothing else pins: generation 2 must be fully closed by now.
	// Generation 1 is still pinned by s: none of it may have closed yet.
	if len(afterClose) != len(names1) {
		t.Fatalf("closed after Close = %v, want exactly generation 2 (%v)", afterClose, names1)
	}

	if err := s.release(); err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(closed) != 2*len(names1) {
		t.Fatalf("closed = %v, want %d entries (both generations)", closed, 2*len(names1))
	}

	counts := map[string]int{}
	for _, name := range closed {
		counts[name]++
	}

	for _, name := range names1 {
		if counts[name] != 2 {
			t.Errorf("%s closed %d times, want 2 (one per generation)", name, counts[name])
		}
	}
}

// TestConcurrentParseReload: Parse never observes a closed dictionary while
// Reload and SetEnabled run (run with -race).
func TestConcurrentParseReload(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "surname.test.dat", datBytes(t, surnameForms))

	r := openTest(t, dir, nil)
	ctx := t.Context()
	stop := make(chan struct{})

	var wg sync.WaitGroup

	for range 4 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				select {
				case <-stop:
					return
				default:
				}

				for _, rd := range r.Parse("кузнецов", nil) {
					if rd.Normal == "" {
						t.Error("empty lemma")
					}
				}
			}
		}()
	}

	var err error

	for i := 0; i < 20 && err == nil; i++ {
		if err = r.Reload(ctx); err == nil {
			err = r.SetEnabled(ctx, "surname.test", i%2 == 0)
		}
	}

	close(stop)
	wg.Wait()

	if err != nil {
		t.Fatal(err)
	}
}

// TestClose: after Close, Parse is empty and state changes fail.
func TestClose(t *testing.T) {
	r := openTest(t, "", nil)

	if err := r.Close(); err != nil {
		t.Fatal(err)
	}

	if r.Parse("кота", nil) != nil || len(r.List()) != 0 {
		t.Fatal("closed registry parses or lists")
	}

	if err := r.SetEnabled(t.Context(), "abbrev.test", false); !errors.Is(err, ErrClosed) {
		t.Fatalf("SetEnabled after Close: %v", err)
	}

	if err := r.Reload(t.Context()); !errors.Is(err, ErrClosed) {
		t.Fatalf("Reload after Close: %v", err)
	}

	if err := r.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}
