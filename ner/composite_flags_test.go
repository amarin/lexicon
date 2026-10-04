package ner

import (
	"strings"
	"testing"

	"github.com/amarin/lexicon/gazetteer"
)

// A labelled span with nested parts takes the word-level flags Ambiguous
// and Predicted only from words outside its parts: a part settled its own
// words (decision P12).
func TestCompositeDoesNotRepeatAbbrevAmbiguity(t *testing.T) {
	const yaml = `sets:
  - name: s
    hints:
      - {lemma: село, dotted: true, type: division, dir: right, window: 1, weight: 2, absorb: true}
    patterns:
      - name: place
        elements:
          - group:
              - {type: division}
            role: p
        actions:
          - label: {role: p, type: place}
`
	p, _ := newPipeline(t, testEntries(), yaml, func(c *Config) {
		c.Nesting = map[string][]string{"place": {"division"}}
	})
	res := extract(t, p, Doc{Text: "жил с. Лягушкино"})
	assertBrief(t, res.Spans, "place:с. Лягушкино", "division:с. Лягушкино")
	for _, sp := range res.Spans {
		if !sp.Flags.Has(Abbrev) || sp.Flags.Has(Ambiguous) {
			t.Fatalf("%s: flags = %v", sp.Type, sp.Flags)
		}
	}
}

func TestCompositePredicted(t *testing.T) {
	// «Сидоровке» is known by a predicted lemma only.
	entries := append(testEntries(), gazetteer.Entry{Alias: "Сидоровка", Type: "surname", Ref: "surname:8", Canonical: "Сидоровка"})

	// Every word is inside a part: the part carries the flag, the person does not.
	p, _ := newPipeline(t, entries, personOverNames, personNesting)
	res := extract(t, p, Doc{Text: "Иван Сидоровке"})
	assertBrief(t, res.Spans, "person:Иван Сидоровке", "given_name:Иван", "surname:Сидоровке")
	if who, part := res.Spans[0], res.Spans[2]; who.Flags.Has(Predicted) || !part.Flags.Has(Predicted) {
		t.Fatalf("person flags = %v, surname flags = %v", who.Flags, part.Flags)
	}

	// A word outside the parts still marks the composite.
	const withGap = `
sets:
  - name: persons
    patterns:
      - name: person
        elements:
          - group:
              - {type: given_name}
              - {token: word}
              - {type: surname}
            role: who
        actions:
          - label: {role: who, type: person}
`
	p, _ = newPipeline(t, testEntries(), withGap, personNesting)
	res = extract(t, p, Doc{Text: "Иван Сидоровке Петров"})
	assertBrief(t, res.Spans, "person:Иван Сидоровке Петров", "given_name:Иван", "surname:Петров")
	if who := res.Spans[0]; !who.Flags.Has(Predicted) {
		t.Fatalf("person flags = %v", who.Flags)
	}

	// A labelled span without parts keeps its own words' flags.
	p, _ = newPipeline(t, testEntries(), strings.Replace(surnameAfterGiven, "TYPE", "nickname", 1))
	res = extract(t, p, Doc{Text: "Иван Сидоровке"})
	if nick := spanOf(t, res.Spans, "nickname", "Сидоровке"); !nick.Flags.Has(Predicted) {
		t.Fatalf("nickname flags = %v", nick.Flags)
	}
}
