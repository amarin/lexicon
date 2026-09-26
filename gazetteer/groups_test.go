package gazetteer

import (
	"context"
	"reflect"
	"testing"
)

func TestVariantGroups(t *testing.T) {
	an := stubAnalyzer{version: "s1"}
	b := &builder{analyzer: an}
	build := func(name string, entries ...Entry) *compiledSource {
		return b.build(context.Background(), NewSliceSource(name, "1", entries), "1", "s1", &compiledSource{name: name})
	}
	names := build("names",
		Entry{Alias: "Иван/иван", Type: "given_name", Ref: "grp:1", Canonical: "Иван"},
		Entry{Alias: "Иоанн/иоанн", Type: "given_name", Ref: "grp:1", Canonical: "Иван"},
		Entry{Alias: "Евдокия/евдокия", Type: "given_name", Ref: "grp:2", Canonical: "Евдокия"},
		Entry{Alias: "Авдотья/авдотья", Type: "given_name", Ref: "grp:2", Canonical: "Евдокия"},
		Entry{Alias: "Петров/петров", Type: "surname", Canonical: "Петров"},
	)
	other := build("other", Entry{Alias: "Ваня/ваня", Type: "given_name", Ref: "grp:1", Canonical: "Ваня"})
	snap := &Snapshot{sources: []*compiledSource{names, other}}

	check := func(name string, got, want []string) {
		t.Helper()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
	check(`Canonical("иоанн")`, snap.Canonical("иоанн"), []string{"Иван"})
	check(`Canonical("петров")`, snap.Canonical("петров"), []string{"Петров"})
	check(`Canonical("нет")`, snap.Canonical("нет"), nil)
	check(`Expand("иван")`, snap.Expand("иван"), []string{"иван", "иоанн"})
	check(`Expand("авдотья")`, snap.Expand("авдотья"), []string{"авдотья", "евдокия"})
	check(`Expand("петров")`, snap.Expand("петров"), nil)
	check(`Expand("ваня")`, snap.Expand("ваня"), []string{"ваня"}) // same Ref, other source: separate group
}
