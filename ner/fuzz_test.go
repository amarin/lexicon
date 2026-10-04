package ner

import (
	"context"
	"testing"
	"unicode/utf8"
)

// FuzzExtractOffsets checks the span offset contract on arbitrary input.
func FuzzExtractOffsets(f *testing.F) {
	for _, seed := range []string{
		"крестьянин деревни Лягушкиной Иван Петров",
		"Жил в дер. Лягушкиной.",
		"на ул. Мира, д. 5",
		"Иоа́нн Петровъ",
		"И. Петров; СПб",
		"Иван Петров Сидоров Анна Михаилова Кузнецова",
		"Анна Петрова деревни Головина, 25 лет.",
		"отвѣтчикъ Сидоровъ; крест. Иванъ Лаптевъ.",
		"Мария, дочь Ивана Петрова, 25 лет",
	} {
		f.Add(seed)
	}
	places, _ := newPipeline(f, testEntries(), placesRules)
	facts, _ := newPipeline(f, testEntries(), factRules, personNesting)
	pipelines := []*Pipeline{places, personsPipeline(f, true), facts}
	f.Fuzz(func(t *testing.T, text string) {
		if !utf8.ValidString(text) {
			t.Skip()
		}
		for _, p := range pipelines {
			checkOffsets(t, p, text)
		}
	})
}

// checkOffsets checks the span offset contract of one extraction, labelled
// spans and abbreviation dots included, and that facts point at spans.
func checkOffsets(t *testing.T, p *Pipeline, text string) {
	t.Helper()
	res, err := p.Extract(context.Background(), Doc{Text: text})
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range res.Spans {
		if s.Start < 0 || s.Start > s.End || s.End > len(text) || text[s.Start:s.End] != s.Surface {
			t.Fatalf("bad byte offsets %+v in %q", s, text)
		}
		if s.RuneStart != utf8.RuneCountInString(text[:s.Start]) || s.RuneEnd != utf8.RuneCountInString(text[:s.End]) {
			t.Fatalf("bad rune offsets %+v in %q", s, text)
		}
	}
	for _, f := range res.Facts {
		for role, i := range f.Args {
			if i < 0 || i >= len(res.Spans) {
				t.Fatalf("fact %s arg %s points at span %d of %d in %q", f.Kind, role, i, len(res.Spans), text)
			}
		}
	}
}
