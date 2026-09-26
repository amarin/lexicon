package ner

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/amarin/lexicon/gazetteer"
	"github.com/amarin/lexicon/internal/fakedict"
)

// Extract pins one snapshot per call; refreshing the gazetteer concurrently
// must neither race nor change results for unchanged content.
func TestExtractConcurrentWithRefresh(t *testing.T) {
	p, gz := newPipeline(t, testEntries(), placesRules)
	const text = "крестьянин деревни Лягушкиной Иван Петров"
	want := brief(extract(t, p, Doc{Text: text}).Spans)
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				res, err := p.Extract(context.Background(), Doc{Text: text}, Explain())
				if err != nil {
					t.Error(err)
					return
				}
				if got := brief(res.Spans); fmt.Sprint(got) != fmt.Sprint(want) {
					t.Errorf("got %q, want %q", got, want)
					return
				}
			}
		}()
	}
	for i := 0; i < 20; i++ {
		if _, err := gz.RefreshSource(context.Background(), "test"); err != nil {
			t.Fatal(err)
		}
	}
	wg.Wait()
}

// A source whose content changes between refreshes: every Extract sees
// either the old or the new dictionary as a whole, spans and version
// together, never a mix.
func TestExtractConcurrentWithChangingSource(t *testing.T) {
	ctx := context.Background()
	b := slices.Clone(testEntries())
	b = slices.DeleteFunc(b, func(e gazetteer.Entry) bool { return e.Type == "surname" })
	b = append(b, gazetteer.Entry{Alias: "Петров", Type: "division", Ref: "division:9", Canonical: "Петров"})
	src := &toggleSource{a: testEntries(), x: b}
	an := fakedict.NewAnalyzer()
	gz, err := gazetteer.New(ctx, gazetteer.Config{
		Analyzer: an, TypeProfiles: fakedict.TypeProfiles(), DefaultProfile: fakedict.Profiles()["text"],
		Sources: []gazetteer.Source{src},
	})
	if err != nil {
		t.Fatal(err)
	}
	p, err := New(Config{Analyzer: an, Gazetteer: gz, Profiles: fakedict.Profiles(), DefaultProfile: "text"})
	if err != nil {
		t.Fatal(err)
	}
	const text = "крестьянин деревни Лягушкиной Иван Петров"
	render := func(r Result) string { return fmt.Sprint(r.Version, brief(r.Spans)) }
	wantA := render(extract(t, p, Doc{Text: text}))
	src.flip()
	if _, err := gz.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	wantB := render(extract(t, p, Doc{Text: text}))
	if wantA == wantB {
		t.Fatalf("the two dictionaries must give different outputs: %s", wantA)
	}
	var wg sync.WaitGroup
	stop := make(chan struct{})
	var seenA, seenB atomic.Int64
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				res, err := p.Extract(ctx, Doc{Text: text})
				if err != nil {
					t.Error(err)
					return
				}
				switch render(res) {
				case wantA:
					seenA.Add(1)
				case wantB:
					seenB.Add(1)
				default:
					t.Errorf("mixed output %s\nwant %s\n  or %s", render(res), wantA, wantB)
					return
				}
			}
		}()
	}
	for i := 0; i < 40; i++ {
		src.flip()
		if _, err := gz.Refresh(ctx); err != nil {
			t.Fatal(err)
		}
	}
	close(stop)
	wg.Wait()
	t.Logf("outputs seen: old %d, new %d", seenA.Load(), seenB.Load())
}
