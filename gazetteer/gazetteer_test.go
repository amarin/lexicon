package gazetteer

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/amarin/lexicon"
)

func matchCount(g *Gazetteer, an Analyzer, text string) int {
	tx := Prepare(an.Analyze(text, lexicon.Profile{}, lexicon.ModeFull))
	return len(g.Snapshot().Match(tx, nil))
}

func TestGazetteerRefresh(t *testing.T) {
	ctx := context.Background()
	an := &versionedStub{v: "a1"}
	names := &mutableSource{name: "names"}
	names.set("1", nil, Entry{Alias: "Иван/иван", Type: "given_name", Ref: "grp:1", Canonical: "Иван"})
	surnames := &mutableSource{name: "surnames"}
	surnames.set("1", nil, Entry{Alias: "Петров/петров", Type: "surname"})

	g, err := New(ctx, Config{Analyzer: an, Sources: []Source{names, surnames}})
	if err != nil {
		t.Fatal(err)
	}
	s0 := g.Snapshot()
	if reps := s0.Reports(); len(reps) != 2 || reps[0].Aliases != 1 || reps[1].Aliases != 1 {
		t.Fatalf("reports = %+v", reps)
	}
	if s0.Version() == "" || matchCount(g, an, "Иван/иван Петров/петров") != 4 {
		t.Fatal("initial build is not usable")
	}

	reps, err := g.Refresh(ctx)
	if err != nil || len(reps) != 0 || g.Snapshot() != s0 {
		t.Fatalf("unchanged refresh must be a no-op: %v %v", reps, err)
	}

	names.set("2", nil, Entry{Alias: "Иоанн/иоанн", Type: "given_name", Ref: "grp:1", Canonical: "Иван"})
	reps, err = g.Refresh(ctx)
	if err != nil || len(reps) != 1 || reps[0].Source != "names" || reps[0].Version != "2" {
		t.Fatalf("changed source refresh: %+v %v", reps, err)
	}
	s1 := g.Snapshot()
	if s1 == s0 || s1.Version() == s0.Version() || matchCount(g, an, "Иоанн/иоанн") != 2 || matchCount(g, an, "Иван/иван") != 0 {
		t.Fatal("new snapshot not published")
	}
	if s0.Version() == "" || len(s0.Match(Prepare(an.Analyze("Иван/иван", lexicon.Profile{}, lexicon.ModeFull)), nil)) != 2 {
		t.Fatal("old snapshot must stay intact for in-flight readers")
	}

	an.mu.Lock()
	an.v = "a2"
	an.mu.Unlock()
	if reps, _ := g.Refresh(ctx); len(reps) != 2 {
		t.Fatalf("analyzer version change must rebuild all sources: %+v", reps)
	}

	rep, err := g.RefreshSource(ctx, "surnames")
	if err != nil || rep.Source != "surnames" {
		t.Fatalf("forced refresh: %+v %v", rep, err)
	}
	if _, err := g.RefreshSource(ctx, "nope"); !errors.Is(err, ErrUnknownSource) {
		t.Fatalf("unknown source: %v", err)
	}

	boom := errors.New("database is locked")
	surnames.set("3", boom)
	reps, _ = g.Refresh(ctx)
	if len(reps) != 1 || !errors.Is(reps[0].Err, boom) || reps[0].Version != "1" {
		t.Fatalf("failed source report: %+v", reps)
	}
	if matchCount(g, an, "Петров/петров") != 2 {
		t.Fatal("failed source must keep matching with its previous data")
	}
	if g.Expand("иоанн")[0] != "иоанн" || g.Canonical("иоанн")[0] != "Иван" {
		t.Fatal("Gazetteer.Expand/Canonical must use the current snapshot")
	}
}

func TestGazetteerConfigErrors(t *testing.T) {
	ctx := context.Background()
	if _, err := New(ctx, Config{}); err == nil {
		t.Fatal("nil analyzer accepted")
	}
	a, b := &mutableSource{name: "x"}, &mutableSource{name: "x"}
	if _, err := New(ctx, Config{Analyzer: stubAnalyzer{}, Sources: []Source{a, b}}); err == nil {
		t.Fatal("duplicate source names accepted")
	}
}

func TestGazetteerConcurrentReadersAndRefresh(t *testing.T) {
	ctx := context.Background()
	an := stubAnalyzer{version: "s1"}
	src := &mutableSource{name: "names"}
	src.set("0", nil, Entry{Alias: "Иван/иван", Type: "given_name"})
	g, err := New(ctx, Config{Analyzer: an, Sources: []Source{src}})
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	stop := make(chan struct{})
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					if matchCount(g, an, "Иван/иван") != 2 {
						t.Error("reader saw a half-built snapshot")
						return
					}
				}
			}
		}()
	}
	for i := 1; i <= 50; i++ {
		src.set(string(rune('a'+i%26))+"v", nil)
		if _, err := g.Refresh(ctx); err != nil {
			t.Fatal(err)
		}
	}
	close(stop)
	wg.Wait()
}

// Snapshot.Version covers the profile configuration: aliases analyzed with
// other profiles compile to other keys.
func TestSnapshotVersionCoversProfiles(t *testing.T) {
	ctx := context.Background()
	src := NewSliceSource("s", "1", []Entry{{Alias: "Иван", Type: "given_name"}})
	build := func(tp map[string]lexicon.Profile, def lexicon.Profile) string {
		t.Helper()
		g, err := New(ctx, Config{Analyzer: stubAnalyzer{}, TypeProfiles: tp, DefaultProfile: def, Sources: []Source{src}})
		if err != nil {
			t.Fatal(err)
		}
		return g.Snapshot().Version()
	}
	names := lexicon.Profile{Name: "name", Kinds: []lexicon.Kind{lexicon.KindBase}}
	base := build(map[string]lexicon.Profile{"given_name": names}, lexicon.Profile{Name: "text"})
	if again := build(map[string]lexicon.Profile{"given_name": names}, lexicon.Profile{Name: "text"}); again != base {
		t.Fatal("identical configuration must give the same version")
	}
	names.Kinds = []lexicon.Kind{lexicon.KindBase, "given"}
	if build(map[string]lexicon.Profile{"given_name": names}, lexicon.Profile{Name: "text"}) == base {
		t.Fatal("changing TypeProfiles must change Snapshot.Version")
	}
	if build(map[string]lexicon.Profile{"given_name": {Name: "name", Kinds: []lexicon.Kind{lexicon.KindBase}}}, lexicon.Profile{Name: "other"}) == base {
		t.Fatal("changing DefaultProfile must change Snapshot.Version")
	}
}

// New copies the profiles: editing the host's map afterwards changes
// neither how Refresh compiles nor the version.
func TestNewCopiesTypeProfiles(t *testing.T) {
	ctx := context.Background()
	tp := map[string]lexicon.Profile{"given_name": {Name: "name", Kinds: []lexicon.Kind{lexicon.KindBase}}}
	g, err := New(ctx, Config{Analyzer: stubAnalyzer{}, TypeProfiles: tp, Sources: []Source{NewSliceSource("s", "1", []Entry{{Alias: "Иван", Type: "given_name"}})}})
	if err != nil {
		t.Fatal(err)
	}
	tp["given_name"].Kinds[0] = "other"
	tp["surname"] = lexicon.Profile{Name: "x"}
	if g.b.typeProfiles["given_name"].Kinds[0] != lexicon.KindBase || len(g.b.typeProfiles) != 1 {
		t.Fatalf("host edits leaked: %+v", g.b.typeProfiles)
	}
}
