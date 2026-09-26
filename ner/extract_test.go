package ner

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestExtractGazetteerSpans(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), "")
	res := extract(t, p, Doc{Text: "крестьянин Лягушкиной Иоанн Петров"})
	assertBrief(t, res.Spans, "estate:крестьянин", "division:Лягушкиной", "given_name:Иоанн", "surname:Петров")

	div := spanOf(t, res.Spans, "division", "Лягушкиной")
	if !reflect.DeepEqual(div.Normal, []string{"Лягушкино"}) || !reflect.DeepEqual(div.Refs, []string{"division:1"}) ||
		div.Attrs["division_type"] != "деревня" || div.Flags != 0 {
		t.Fatalf("division span = %+v", div)
	}
	if g := spanOf(t, res.Spans, "given_name", "Иоанн"); !reflect.DeepEqual(g.Normal, []string{"Иван"}) {
		t.Fatalf("variant must normalize to the record's canonical form: %+v", g)
	}
	if res.Spans[0].Evidence != nil {
		t.Fatal("evidence without Explain")
	}
	res = extract(t, p, Doc{Text: "крестьянин Лягушкиной"}, Explain())
	if ev := spanOf(t, res.Spans, "division", "Лягушкиной").Evidence; len(ev) != 1 || ev[0] != "lemma match «Лягушкина» (test)" {
		t.Fatalf("evidence = %q", ev)
	}
}

func TestExtractOffsets(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), "")
	const text = "Жил в дер. Лягушкиной."
	res := extract(t, p, Doc{Text: text})
	sp := spanOf(t, res.Spans, "division", "Лягушкиной")
	start := strings.Index(text, "Лягушкиной")
	if sp.Start != start || sp.End != start+len("Лягушкиной") || text[sp.Start:sp.End] != sp.Surface {
		t.Fatalf("byte offsets: %+v", sp)
	}
	if sp.RuneStart != utf8.RuneCountInString(text[:sp.Start]) || sp.RuneEnd != utf8.RuneCountInString(text[:sp.End]) {
		t.Fatalf("rune offsets: %+v", sp)
	}
}

func TestExtractTypesFilter(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), "")
	res := extract(t, p, Doc{Text: "крестьянин Лягушкиной Иван Петров", Types: []string{"given_name", "surname"}})
	assertBrief(t, res.Spans, "given_name:Иван", "surname:Петров")
}
