package gazetteer

import (
	"context"
	"testing"
)

// buildSnapshot compiles entries as one source named "test" and wraps it in
// a snapshot, without a Gazetteer.
func buildSnapshot(t testing.TB, an Analyzer, entries ...Entry) *Snapshot {
	t.Helper()
	b := &builder{analyzer: an}
	cs := b.build(context.Background(), NewSliceSource("test", "1", entries), "1", an.Version(), &compiledSource{name: "test"})
	if cs.report.Err != nil {
		t.Fatal(cs.report.Err)
	}
	return &Snapshot{sources: []*compiledSource{cs}}
}

func aliasByText(cs *compiledSource, alias string) *Alias {
	for _, a := range cs.aliases {
		if a.Entry.Alias == alias {
			return a
		}
	}
	return nil
}
