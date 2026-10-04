package ner

import (
	"testing"
	"unicode/utf8"

	"github.com/amarin/lexicon/gazetteer"
)

// A span whose last word is a dotted abbreviation ends after the dot
// (decision P16).
func TestSpanEndsAfterAbbreviationDot(t *testing.T) {
	entries := append(testEntries(), gazetteer.Entry{Alias: "губ.", Type: "division", Ref: "division:10", Canonical: "губерния"})
	p, _ := newPipeline(t, entries, placesRules)
	for _, tc := range []struct {
		name, text string
		want       []string
	}{
		{"abbreviation lemma", "крест. Иван", []string{"estate:крест.", "given_name:Иван"}},
		{"alias written with a dot", "Калужской губ. Иван", []string{"division:губ.", "given_name:Иван"}},
		{"abbreviation at the end of the text", "жил крест.", []string{"estate:крест."}},
		{"sentence dot after an ordinary word", "жил в Лягушкиной.", []string{"division:Лягушкиной"}},
		{"dot inside the span", "жил в дер. Лягушкиной", []string{"division:дер. Лягушкиной"}},
	} {
		res := extract(t, p, Doc{Text: tc.text})
		assertBrief(t, res.Spans, tc.want...)
		for _, s := range res.Spans {
			if tc.text[s.Start:s.End] != s.Surface ||
				s.RuneEnd != utf8.RuneCountInString(tc.text[:s.End]) {
				t.Fatalf("%s: offsets of %+v", tc.name, s)
			}
		}
	}
}

func TestExtractorVersion(t *testing.T) {
	if extractorVersion != "ner-3" {
		t.Fatalf("extractorVersion = %q: P16 changes span ends, hosts must re-extract", extractorVersion)
	}
}
