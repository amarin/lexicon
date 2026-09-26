package gazetteer

import (
	"context"
	"fmt"
	"testing"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/internal/fakedict"
)

func describe(tx *Text, ms []Match) []string {
	var out []string
	for _, m := range ms {
		kind := "lemma"
		if m.Kind == BySurface {
			kind = "surface"
		}
		for _, a := range m.Aliases {
			out = append(out, fmt.Sprintf("%d-%d %s %s", m.Start, m.End, kind, a.Entry.Type))
		}
	}
	return out
}

func assertMatches(t *testing.T, got, want []string) {
	t.Helper()
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("matches:\n got %q\nwant %q", got, want)
	}
}

var stubEntries = []Entry{
	{Alias: "Иван/иван Петров/петров", Type: "person"},
	{Alias: "Петров/петров", Type: "surname"},
	{Alias: "С/село/сын Иваново", Type: "division"},
}

func TestMatchStub(t *testing.T) {
	an := stubAnalyzer{version: "s1"}
	snap := buildSnapshot(t, an, stubEntries...)
	match := func(text string) []string {
		tx := Prepare(an.Analyze(text, lexicon.Profile{}, lexicon.ModeFull))
		return describe(tx, snap.Match(tx, nil))
	}
	// Lemma match across inflection; surface only when forms are equal.
	assertMatches(t, match("у/у иван/иван петрова/петров"), []string{"1-3 lemma person", "2-3 lemma surname"})
	assertMatches(t, match("Иван/иван Петров/петров"), []string{
		"0-2 lemma person", "0-2 surface person", "1-2 lemma surname", "1-2 surface surname",
	})
	// Punctuation between words breaks a multi-word match.
	assertMatches(t, match("иван/иван . петров/петров"), []string{"1-2 lemma surname", "1-2 surface surname"})
	// Ambiguous text token branches over its lemmas.
	assertMatches(t, match("с/сын/село иваново"), []string{"0-2 lemma division", "0-2 lemma division", "0-2 surface division"})
	// Unknown lemmas never match.
	assertMatches(t, match("никто/никто"), nil)
}

func TestMatchWithAnalyzer(t *testing.T) {
	an := fakedict.NewAnalyzer()
	b := &builder{analyzer: an, typeProfiles: fakedict.TypeProfiles(), defaultProfile: fakedict.Profiles()["text"]}
	src := NewSliceSource("places", "1", []Entry{
		{Alias: "дер. Лягушкино", Type: "division", Ref: "division:1", Canonical: "Лягушкино"},
		{Alias: "Лягушкина", Type: "division", Ref: "division:1", Canonical: "Лягушкино"},
	})
	snap := &Snapshot{sources: []*compiledSource{b.build(context.Background(), src, "1", an.Version(), &compiledSource{name: "places"})}}
	match := func(text string) []string {
		tx := Prepare(an.Analyze(text, fakedict.Profiles()["text"], lexicon.ModeFull))
		return describe(tx, snap.Match(tx, nil))
	}
	assertMatches(t, match("в дер. Лягушкиной"), []string{"2-3 lemma division"})
	assertMatches(t, match("деревня Лягушкино"), []string{"0-2 lemma division"})
	assertMatches(t, match("в дер. Лягушкино"), []string{"1-3 lemma division", "1-3 surface division"})
	assertMatches(t, match("деревня. Лягушкино"), nil)
}

func TestMatchDoesNotAllocate(t *testing.T) {
	an := stubAnalyzer{version: "s1"}
	snap := buildSnapshot(t, an, stubEntries...)
	tx := Prepare(an.Analyze("у/у Иван/иван Петров/петров с/сын/село иваново", lexicon.Profile{}, lexicon.ModeFull))
	buf := make([]Match, 0, 32)
	allocs := testing.AllocsPerRun(100, func() { buf = snap.Match(tx, buf[:0]) })
	if allocs != 0 {
		t.Fatalf("Match allocated %.1f times per run", allocs)
	}
}
