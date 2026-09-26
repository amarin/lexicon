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

// TestFilterContextRecomputesOrigin: a candidate can carry a surface hit
// from a RequiresContext alias alongside a lemma-only hit from a plain
// alias. Dropping the first must not leave the candidate scored as if it
// were still a surface match.
func TestFilterContextRecomputesOrigin(t *testing.T) {
	entries := append(testEntries(),
		gazetteer.Entry{Alias: "Петрова", Type: "surname", Ref: "surname:900", Canonical: "Петров", Flags: gazetteer.RequiresContext})
	p, _ := newPipeline(t, entries, "")
	res := extract(t, p, Doc{Text: "видел Петрова"})
	sp := spanOf(t, res.Spans, "surname", "Петрова")
	if sp.Score != 2 {
		t.Fatalf("mixed RC-surface + lemma candidate without context must score like a lemma-only match: %+v", sp)
	}
}

// TestFilterEarlyRecomputesOrigin: same scenario, but the surface hit is
// dropped by the case-sensitive check instead of RequiresContext.
func TestFilterEarlyRecomputesOrigin(t *testing.T) {
	entries := append(testEntries(),
		gazetteer.Entry{Alias: "петрова", Type: "surname", Ref: "surname:901", Canonical: "Петров", Flags: gazetteer.CaseSensitive})
	p, _ := newPipeline(t, entries, "")
	res := extract(t, p, Doc{Text: "видел Петрова"})
	sp := spanOf(t, res.Spans, "surname", "Петрова")
	if sp.Score != 2 {
		t.Fatalf("mixed case-mismatch-surface + lemma candidate must score like a lemma-only match: %+v", sp)
	}
}

// TestBlockedVetoesLemmaOnlyMatch: D3 removes the whole candidate of a
// blocked entry's type on its exact range, even when the blocked entry
// itself was matched only by lemma, not by surface.
func TestBlockedVetoesLemmaOnlyMatch(t *testing.T) {
	entries := append(testEntries(),
		gazetteer.Entry{Alias: "Петров", Type: "surname", Ref: "surname:902", Flags: gazetteer.Blocked})
	p, _ := newPipeline(t, entries, "")
	assertBrief(t, extract(t, p, Doc{Text: "видел Петрова"}).Spans)
}
