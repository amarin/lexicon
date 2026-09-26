package ner

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/amarin/lexicon/gazetteer"
	"github.com/amarin/lexicon/internal/fakedict"
	"github.com/amarin/lexicon/rules"
)

func testEntries() []gazetteer.Entry {
	village := map[string]string{"division_type": "деревня"}
	return []gazetteer.Entry{
		{Alias: "крестьянин", Type: "estate", Ref: "estate:1", Canonical: "крестьянин"},
		{Alias: "крест.", Type: "estate", Ref: "estate:1", Canonical: "крестьянин"},
		{Alias: "Лягушкино", Type: "division", Ref: "division:1", Canonical: "Лягушкино", Attrs: village},
		{Alias: "Лягушкина", Type: "division", Ref: "division:1", Canonical: "Лягушкино", Attrs: village},
		{Alias: "Боровский", Type: "division", Ref: "division:2", Canonical: "Боровский", Attrs: map[string]string{"division_type": "уезд"}},
		{Alias: "СПб", Type: "division", Ref: "division:3", Canonical: "Санкт-Петербург", Flags: gazetteer.CaseSensitive | gazetteer.SurfaceOnly},
		{Alias: "Иван", Type: "given_name", Ref: "given_name:1", Canonical: "Иван"},
		{Alias: "Иоанн", Type: "given_name", Ref: "given_name:1", Canonical: "Иван"},
		{Alias: "Мария", Type: "given_name", Ref: "given_name:2", Canonical: "Мария"},
		{Alias: "Вера", Type: "given_name", Ref: "given_name:3", Canonical: "Вера"},
		{Alias: "Петров", Type: "surname", Ref: "surname:1", Canonical: "Петров"},
		{Alias: "Мороз", Type: "surname", Ref: "surname:2", Canonical: "Мороз", Flags: gazetteer.RequiresContext},
		{Alias: "Ус", Type: "surname", Ref: "surname:3", Canonical: "Ус"},
		{Alias: "Мира", Type: "street", Ref: "street:1", Canonical: "Мира", Flags: gazetteer.SurfaceOnly | gazetteer.RequiresContext},
	}
}

// placesRules: place hints/trigger always on; a surname hint gated by period:pre1917.
const placesRules = `meta:
  source: lexicon tests
sets:
  - name: places
    hints:
      # «деревня Лягушкино», «дер. Лягушкиной»: boost and absorb the keyword.
      - {lemma: деревня, type: division, dir: right, window: 1, weight: 2, absorb: true}
      # «Боровского уезда»: the keyword follows the name.
      - {lemma: уезд, type: division, dir: left, window: 1, weight: 2, absorb: true}
      # «ул. Мира»: only the dotted abbreviation gives street context.
      - {lemma: ул, dotted: true, type: street, dir: right, window: 1, weight: 2, absorb: true}
    triggers:
      # An unknown title-case word after «деревня»/«село» is a division candidate.
      - lemma: деревня|село
        type: division
        dir: right
        window: 1..2
        shape: {case: title, script: cyrillic}
        stop_at: [punct, stop]
        weight: 1
        absorb: true
  - name: pre1917
    when: ["period:pre1917"]
    hints:
      - {lemma: крестьянин, type: surname, dir: right, window: 2, weight: 1}
`

func newPipeline(t testing.TB, entries []gazetteer.Entry, ruleYAML string, edit ...func(*Config)) (*Pipeline, *gazetteer.Gazetteer) {
	t.Helper()
	an := fakedict.NewAnalyzer()
	gz, err := gazetteer.New(context.Background(), gazetteer.Config{
		Analyzer:       an,
		TypeProfiles:   fakedict.TypeProfiles(),
		DefaultProfile: fakedict.Profiles()["text"],
		Sources:        []gazetteer.Source{gazetteer.NewSliceSource("test", "1", entries)},
	})
	if err != nil {
		t.Fatal(err)
	}
	var book *rules.Book
	if ruleYAML != "" {
		f, err := rules.Load(strings.NewReader(ruleYAML))
		if err != nil {
			t.Fatal(err)
		}
		if book, err = rules.Compile(f); err != nil {
			t.Fatal(err)
		}
	}
	cfg := Config{Analyzer: an, Gazetteer: gz, Rules: book, Profiles: fakedict.Profiles(), DefaultProfile: "text"}
	for _, e := range edit {
		e(&cfg)
	}
	p, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return p, gz
}

func extract(t testing.TB, p *Pipeline, d Doc, o ...Option) Result {
	t.Helper()
	res, err := p.Extract(context.Background(), d, o...)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

// brief renders spans as "type:surface" in output order.
func brief(spans []Span) []string {
	out := make([]string, len(spans))
	for i, s := range spans {
		out[i] = fmt.Sprintf("%s:%s", s.Type, s.Surface)
	}
	return out
}

func assertBrief(t *testing.T, got []Span, want ...string) {
	t.Helper()
	if fmt.Sprint(brief(got)) != fmt.Sprint(want) {
		t.Fatalf("spans:\n got %q\nwant %q", brief(got), want)
	}
}

func spanOf(t *testing.T, spans []Span, typ, surface string) Span {
	t.Helper()
	for _, s := range spans {
		if s.Type == typ && s.Surface == surface {
			return s
		}
	}
	t.Fatalf("no span %s:%s in %q", typ, surface, brief(spans))
	return Span{}
}
