package ner

import (
	"context"
	"slices"
	"testing"
	"unicode/utf8"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/gazetteer"
	"github.com/amarin/lexicon/internal/fakedict"
	"github.com/amarin/lexicon/textnorm"
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

// TestHintWindowStopsAtPunctuation: a comma between the keyword and the
// candidate breaks the join, so the hint neither boosts nor absorbs it.
func TestHintWindowStopsAtPunctuation(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), placesRules)
	res := extract(t, p, Doc{Text: "деревня, Лягушкино"}, Explain())
	assertBrief(t, res.Spans, "division:Лягушкино")
	sp := res.Spans[0]
	if slices.Contains(sp.Evidence, "hint «деревня» → division +2") {
		t.Fatalf("expected no hint boost across punctuation, evidence = %q", sp.Evidence)
	}
}

// TestHintWindowAtDistance: a window of 2 reaches a candidate two content
// words away from the keyword.
func TestHintWindowAtDistance(t *testing.T) {
	const wide = `sets:
  - name: wide
    hints:
      - {lemma: около, type: division, dir: right, window: 2, weight: 5}
`
	p, _ := newPipeline(t, testEntries(), wide)
	res := extract(t, p, Doc{Text: "около большое Лягушкино"}, Explain())
	assertBrief(t, res.Spans, "division:Лягушкино")
	sp := res.Spans[0]
	if !slices.Contains(sp.Evidence, "hint «около» → division +5") {
		t.Fatalf("evidence = %q", sp.Evidence)
	}
}

// TestAbsorbedSpanOffsetsMatchSurface: an absorbed span (the hint keyword
// merged into the candidate) still satisfies the Text[Start:End] == Surface
// invariant, in bytes and in runes.
func TestAbsorbedSpanOffsetsMatchSurface(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), placesRules)
	doc := Doc{Text: "Жил в дер. Лягушкиной."}
	res := extract(t, p, doc)
	d := spanOf(t, res.Spans, "division", "дер. Лягушкиной")
	if doc.Text[d.Start:d.End] != d.Surface {
		t.Fatalf("Text[%d:%d] = %q, want Surface %q", d.Start, d.End, doc.Text[d.Start:d.End], d.Surface)
	}
	if utf8.RuneCountInString(doc.Text[:d.Start]) != d.RuneStart || utf8.RuneCountInString(doc.Text[:d.End]) != d.RuneEnd {
		t.Fatalf("rune offsets mismatch: RuneStart=%d RuneEnd=%d for %q", d.RuneStart, d.RuneEnd, d.Surface)
	}
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

// A keyword absorbed by a rule was disambiguated by that rule: it keeps
// contributing Abbrev but no longer Ambiguous («с.» is село or сын).
func TestAbsorbedKeywordIsNotAmbiguous(t *testing.T) {
	const yaml = `sets:
  - name: s
    hints:
      - {lemma: село, dotted: true, type: division, dir: right, window: 1, weight: 2, absorb: true}
`
	p, _ := newPipeline(t, testEntries(), yaml)
	d := spanOf(t, extract(t, p, Doc{Text: "с. Лягушкино"}).Spans, "division", "с. Лягушкино")
	if !d.Flags.Has(Abbrev) || d.Flags.Has(Ambiguous) {
		t.Fatalf("hint-absorbed «с.»: flags = %v", d.Flags)
	}
	p, _ = newPipeline(t, testEntries(), placesRules)
	d = spanOf(t, extract(t, p, Doc{Text: "с. Новое"}).Spans, "division", "с. Новое")
	if !d.Flags.Has(Abbrev) || d.Flags.Has(Ambiguous) || !d.Flags.Has(Candidate) {
		t.Fatalf("trigger-absorbed «с.»: flags = %v", d.Flags)
	}
}

// Of the covered words only an ambiguous abbreviation sets Ambiguous:
// homonymy of an ordinary word («стали»: сталь or стать) is resolved by the
// alias match, one ref and one normal form stay unambiguous.
func TestCoveredWordAmbiguity(t *testing.T) {
	d := fakedict.Genealogy()
	d.Add(lexicon.KindBase, "base.fake", "сталь", "NOUN,inan,femn sing,nomn", "сталь", "стали")
	d.Add(lexicon.KindBase, "base.fake", "стать", "VERB,perf,intr", "стали")
	an := lexicon.NewAnalyzer(d, textnorm.Modern, lexicon.AnalyzerOptions{})
	entries := []gazetteer.Entry{
		{Alias: "стали", Type: "material", Ref: "material:1", Canonical: "сталь"},
		{Alias: "с. Новое", Type: "division", Ref: "division:9", Canonical: "Новое"},
	}
	gz, err := gazetteer.New(context.Background(), gazetteer.Config{
		Analyzer: an, DefaultProfile: fakedict.Profiles()["text"],
		Sources: []gazetteer.Source{gazetteer.NewSliceSource("test", "1", entries)},
	})
	if err != nil {
		t.Fatal(err)
	}
	p, err := New(Config{Analyzer: an, Gazetteer: gz, Profiles: fakedict.Profiles(), DefaultProfile: "text"})
	if err != nil {
		t.Fatal(err)
	}
	res := extract(t, p, Doc{Text: "из стали в с. Новое"})
	if m := spanOf(t, res.Spans, "material", "стали"); m.Flags.Has(Ambiguous) {
		t.Fatalf("homonym word: flags = %v", m.Flags)
	}
	if v := spanOf(t, res.Spans, "division", "с. Новое"); !v.Flags.Has(Ambiguous) || !v.Flags.Has(Abbrev) {
		t.Fatalf("ambiguous abbreviation: flags = %v", v.Flags)
	}
}

// A full word of an abbreviation dictionary («погост», abbreviated «пог.»)
// is not an abbreviation: only the abbreviated form marks the span Abbrev.
func TestFullWordOfAbbreviationIsNotAbbrev(t *testing.T) {
	entries := []gazetteer.Entry{{Alias: "погост Новое", Type: "division", Ref: "division:9", Canonical: "Новое"}}
	p, _ := newPipeline(t, entries, "")
	if d := spanOf(t, extract(t, p, Doc{Text: "в погост Новое"}).Spans, "division", "погост Новое"); d.Flags.Has(Abbrev) {
		t.Fatalf("full word: flags = %v", d.Flags)
	}
	if d := spanOf(t, extract(t, p, Doc{Text: "в пог. Новое"}).Spans, "division", "пог. Новое"); !d.Flags.Has(Abbrev) {
		t.Fatalf("abbreviated word: flags = %v", d.Flags)
	}
}
