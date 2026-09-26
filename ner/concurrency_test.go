package ner

import (
	"context"
	"fmt"
	"sync"
	"testing"
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
