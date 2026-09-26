package ner

import (
	"testing"

	"github.com/amarin/lexicon/gazetteer"
)

func TestBlockedVeto(t *testing.T) {
	entries := append(testEntries(), gazetteer.Entry{Alias: "Петров", Type: "surname", Ref: "surname:9", Flags: gazetteer.Blocked})
	p, _ := newPipeline(t, entries, "")
	assertBrief(t, extract(t, p, Doc{Text: "Иван Петров"}).Spans, "given_name:Иван")
}

func TestCaseSensitive(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), "")
	res := extract(t, p, Doc{Text: "Прибыл из СПб"})
	assertBrief(t, res.Spans, "division:СПб")
	if sp := res.Spans[0]; sp.Normal[0] != "Санкт-Петербург" {
		t.Fatalf("span = %+v", sp)
	}
	assertBrief(t, extract(t, p, Doc{Text: "прибыл из спб"}).Spans)
}

func TestMinLemmaMatchRunes(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), "")
	assertBrief(t, extract(t, p, Doc{Text: "видел Уса"}).Spans)
	assertBrief(t, extract(t, p, Doc{Text: "видел Ус"}).Spans, "surname:Ус")
	p1, _ := newPipeline(t, testEntries(), "", func(c *Config) { c.MinLemmaMatchRunes = 1 })
	assertBrief(t, extract(t, p1, Doc{Text: "видел Уса"}).Spans, "surname:Уса")
}

func TestRequiresContextWithoutRules(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), "")
	assertBrief(t, extract(t, p, Doc{Text: "ударил мороз"}).Spans)
	assertBrief(t, extract(t, p, Doc{Text: "нет Мира"}).Spans)
}
