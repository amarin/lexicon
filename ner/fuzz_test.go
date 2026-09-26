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
	} {
		f.Add(seed)
	}
	p, _ := newPipeline(f, testEntries(), placesRules)
	f.Fuzz(func(t *testing.T, text string) {
		if !utf8.ValidString(text) {
			t.Skip()
		}
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
	})
}
