package ner

import (
	"reflect"
	"testing"

	"github.com/amarin/lexicon/gazetteer"
)

// «ответчик Петров»: the defining word is context, the person is the
// surname alone (decision P13).
const personByDefiningWord = `
sets:
  - name: persons
    patterns:
      - name: person-by-defining-word
        elements:
          - {lemma: крестьянин|ответчик|истец}
          - group:
              - {type: surname}
            role: who
        actions:
          - label: {role: who, type: person}
`

func TestSameRangeNestingForLabelledSpan(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), personByDefiningWord, personNesting)
	res := extract(t, p, Doc{Text: "ответчик Петров"})
	assertBrief(t, res.Spans, "person:Петров", "surname:Петров")
	who, part := res.Spans[0], res.Spans[1]
	if who.Flags.Has(Ambiguous) || who.Flags.Has(Candidate) || who.Flags.Has(Nested) || who.Alternatives != nil ||
		!reflect.DeepEqual(who.Normal, []string{"Петров"}) {
		t.Fatalf("person = %+v", who)
	}
	if !part.Flags.Has(Nested) || part.Flags.Has(Ambiguous) || part.Alternatives != nil ||
		!reflect.DeepEqual(part.Refs, []string{"surname:1"}) {
		t.Fatalf("surname = %+v", part)
	}
	if who.Start != part.Start || who.End != part.End {
		t.Fatalf("ranges differ: %+v %+v", who, part)
	}
}

// Without the nesting pair the two readings compete as before: the
// dictionary span wins, the labelled one is its alternative.
func TestSameRangeWithoutNestingCompetes(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), personByDefiningWord)
	res := extract(t, p, Doc{Text: "ответчик Петров"})
	assertBrief(t, res.Spans, "surname:Петров")
	sp := res.Spans[0]
	if sp.Flags.Has(Ambiguous) || len(sp.Alternatives) != 1 || sp.Alternatives[0].Type != "person" {
		t.Fatalf("surname = %+v", sp)
	}
}

// Two dictionary candidates on one range still compete, even when Nesting
// lists the pair: same-range nesting is only for labelled spans.
func TestSameRangeDictionarySpansStillCompete(t *testing.T) {
	entries := append(testEntries(), gazetteer.Entry{Alias: "Петров", Type: "patronymic", Ref: "patronymic:7", Canonical: "Петрович"})
	p, _ := newPipeline(t, entries, "", func(c *Config) {
		c.Nesting = map[string][]string{"patronymic": {"surname"}, "surname": {"patronymic"}}
	})
	res := extract(t, p, Doc{Text: "Петров"})
	if len(res.Spans) != 1 || !res.Spans[0].Flags.Has(Ambiguous) || len(res.Spans[0].Alternatives) != 1 {
		t.Fatalf("spans = %+v", res.Spans)
	}
}

// The normal form of a composite is built from its parts.
func TestCompositeNormalFromParts(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), personOverNames, personNesting)
	res := extract(t, p, Doc{Text: "видел Ивана Петрова"})
	assertBrief(t, res.Spans, "person:Ивана Петрова", "given_name:Ивана", "surname:Петрова")
	if who := res.Spans[0]; !reflect.DeepEqual(who.Normal, []string{"Иван Петров"}) || who.Flags.Has(Ambiguous) {
		t.Fatalf("person = %+v", who)
	}
}

// «ответчик Лаптев»: an unknown word labelled surname, then a person over
// it. The labelled part nests on the range of the person labelled later.
const unknownSurnameByDefiningWord = `
sets:
  - name: persons
    patterns:
      - name: unknown-surname-by-defining-word
        elements:
          - {lemma: ответчик|истец}
          - {token: word, shape: {case: title}, not: {type: given_name|patronymic|surname}, role: s}
        actions:
          - label: {role: s, type: surname}
      - name: person-by-defining-word
        elements:
          - {lemma: крестьянин|ответчик|истец}
          - group:
              - {type: surname}
            role: who
        actions:
          - label: {role: who, type: person}
`

func TestSameRangeNestingOfLabelledPart(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), unknownSurnameByDefiningWord, personNesting)
	res := extract(t, p, Doc{Text: "ответчик Лаптев"})
	assertBrief(t, res.Spans, "person:Лаптев", "surname:Лаптев")
	who, part := res.Spans[0], res.Spans[1]
	if who.Flags.Has(Ambiguous) || who.Flags.Has(Candidate) || who.Flags.Has(Nested) || who.Alternatives != nil {
		t.Fatalf("person = %+v", who)
	}
	if !part.Flags.Has(Nested) || !part.Flags.Has(Candidate) || part.Refs != nil {
		t.Fatalf("surname = %+v", part)
	}
}

// A third reading of the range competes with the part, not with the person
// that holds the part on its whole range: alternatives and ties live on
// the part.
func TestSameRangeAlternativesGoToThePart(t *testing.T) {
	entries := append(testEntries(), gazetteer.Entry{Alias: "Лягушкина", Type: "surname", Ref: "surname:7", Canonical: "Лягушкина"})
	p, _ := newPipeline(t, entries, personByDefiningWord, personNesting)
	for _, text := range []string{"ответчик Лягушкиной", "ответчик Лягушкина"} { // lemma match, surface match
		res := extract(t, p, Doc{Text: text})
		if len(res.Spans) != 2 {
			t.Fatalf("%s: spans = %q", text, brief(res.Spans))
		}
		who, part := res.Spans[0], res.Spans[1]
		if who.Type != "person" || who.Flags.Has(Ambiguous) || who.Alternatives != nil {
			t.Fatalf("%s: person = %+v", text, who)
		}
		if part.Type != "surname" || !part.Flags.Has(Nested) || len(part.Alternatives) != 1 || part.Alternatives[0].Type != "division" {
			t.Fatalf("%s: surname = %+v", text, part)
		}
		// The two dictionary readings score the same: the tie is the part's.
		if !part.Flags.Has(Ambiguous) {
			t.Fatalf("%s: surname = %+v", text, part)
		}
	}
}
