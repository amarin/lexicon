package ner

import (
	"slices"
	"testing"

	"github.com/amarin/lexicon/gazetteer"
)

func TestHintsAbsorbAndBoost(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), placesRules)
	res := extract(t, p, Doc{Text: "крестьянин деревни Лягушкиной Иван Петров"}, Explain())
	assertBrief(t, res.Spans, "estate:крестьянин", "division:деревни Лягушкиной", "given_name:Иван", "surname:Петров")
	div := spanOf(t, res.Spans, "division", "деревни Лягушкиной")
	if !slices.Contains(div.Evidence, "hint «деревни» → division +2") {
		t.Fatalf("evidence = %q", div.Evidence)
	}

	res = extract(t, p, Doc{Text: "Жил в дер. Лягушкиной."})
	d := spanOf(t, res.Spans, "division", "дер. Лягушкиной")
	if !d.Flags.Has(Abbrev) || d.Refs[0] != "division:1" {
		t.Fatalf("abbreviated division = %+v", d)
	}

	res = extract(t, p, Doc{Text: "жил на ул. Мира"})
	assertBrief(t, res.Spans, "street:ул. Мира")

	assertBrief(t, extract(t, p, Doc{Text: "Лягушкино Боровского уезда"}).Spans,
		"division:Лягушкино", "division:Боровского уезда")
}

func TestHintSetsGatedByTags(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), placesRules)
	assertBrief(t, extract(t, p, Doc{Text: "крестьянин Мороз"}).Spans, "estate:крестьянин")
	assertBrief(t, extract(t, p, Doc{Text: "крестьянин Мороз", Tags: []string{"period:pre1917"}}).Spans,
		"estate:крестьянин", "surname:Мороз")
}

func TestNestingEndToEnd(t *testing.T) {
	entries := append(testEntries(), gazetteer.Entry{
		Alias: "Лягушкино Боровского уезда", Type: "division", Ref: "division:1", Canonical: "Лягушкино",
	})
	flat, _ := newPipeline(t, entries, placesRules)
	assertBrief(t, extract(t, flat, Doc{Text: "Лягушкино Боровского уезда"}).Spans,
		"division:Лягушкино Боровского уезда")

	nest, _ := newPipeline(t, entries, placesRules, func(c *Config) {
		c.Nesting = map[string][]string{"division": {"division"}}
	})
	res := extract(t, nest, Doc{Text: "Лягушкино Боровского уезда"})
	assertBrief(t, res.Spans, "division:Лягушкино Боровского уезда", "division:Лягушкино", "division:Боровского уезда")
	if res.Spans[0].Flags.Has(Nested) || !res.Spans[1].Flags.Has(Nested) || !res.Spans[2].Flags.Has(Nested) {
		t.Fatalf("nested flags: %v %v %v", res.Spans[0].Flags, res.Spans[1].Flags, res.Spans[2].Flags)
	}
}
