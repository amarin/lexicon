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

// assertNestedInside: every span flagged Nested lies inside another span.
func assertNestedInside(t *testing.T, spans []Span) {
	t.Helper()
	for i, s := range spans {
		if !s.Flags.Has(Nested) {
			continue
		}
		ok := false
		for j, o := range spans {
			if i != j && s.Start >= o.Start && s.End <= o.End {
				ok = true
			}
		}
		if !ok {
			t.Fatalf("nested span %+v lies outside every other span %+v", s, spans)
		}
	}
}

const recordOverDivision = `
sets:
  - name: records
    patterns:
      - name: record
        elements:
          - group:
              - {type: given_name}
              - {type: division}
            role: rec
        actions:
          - label: {role: rec, type: record}
`

const placeAfterWord = `
sets:
  - name: places
    patterns:
      - name: place
        elements:
          - {lemma: жил}
          - group:
              - {type: division}
            role: pl
        actions:
          - label: {role: pl, type: place}
`

// A labelled span takes the dot of its last nested part (a part matched by
// an alias written with a dot), so a nested span never ends past its outer.
func TestLabelledSpanTakesPartAbbreviationDot(t *testing.T) {
	entries := append(testEntries(), gazetteer.Entry{Alias: "губ.", Type: "division", Ref: "division:10", Canonical: "губерния"})
	for _, tc := range []struct {
		name, rules, text string
		nesting           map[string][]string
		want              []string
	}{
		{"last part", recordOverDivision, "Иван губ. жил", map[string][]string{"record": {"given_name", "division"}},
			[]string{"record:Иван губ.", "given_name:Иван", "division:губ."}},
		{"same range", placeAfterWord, "жил губ.", map[string][]string{"place": {"division"}},
			[]string{"place:губ.", "division:губ."}},
	} {
		p, _ := newPipeline(t, entries, tc.rules, func(c *Config) { c.Nesting = tc.nesting })
		res := extract(t, p, Doc{Text: tc.text})
		assertBrief(t, res.Spans, tc.want...)
		assertNestedInside(t, res.Spans)
		outer, inner := res.Spans[0], res.Spans[len(res.Spans)-1]
		if outer.End != inner.End {
			t.Fatalf("%s: ends differ: %+v %+v", tc.name, outer, inner)
		}
		for _, s := range res.Spans {
			if tc.text[s.Start:s.End] != s.Surface {
				t.Fatalf("%s: surface of %+v", tc.name, s)
			}
		}
	}
}
