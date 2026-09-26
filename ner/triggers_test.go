package ner

import (
	"reflect"
	"slices"
	"testing"
)

func TestTriggerCandidate(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), placesRules)
	res := extract(t, p, Doc{Text: "родился в деревне Сидоровке"}, Explain())
	assertBrief(t, res.Spans, "division:деревне Сидоровке")
	sp := res.Spans[0]
	if !sp.Flags.Has(Candidate) || !sp.Flags.Has(Predicted) || len(sp.Refs) != 0 ||
		!reflect.DeepEqual(sp.Normal, []string{"сидоровка"}) {
		t.Fatalf("candidate span = %+v", sp)
	}
	if !slices.Contains(sp.Evidence, "trigger «деревне» → division candidate") {
		t.Fatalf("evidence = %q", sp.Evidence)
	}
}

func TestTriggerBoostsExistingMatch(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), placesRules)
	res := extract(t, p, Doc{Text: "деревня Лягушкино"}, Explain())
	assertBrief(t, res.Spans, "division:деревня Лягушкино")
	sp := res.Spans[0]
	if sp.Flags.Has(Candidate) || sp.Refs[0] != "division:1" || !slices.Contains(sp.Evidence, "trigger «деревня» → division +1") {
		t.Fatalf("span = %+v", sp)
	}
}

func TestTriggerShapeStops(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), placesRules)
	assertBrief(t, extract(t, p, Doc{Text: "в деревне жил Иван"}).Spans, "given_name:Иван")
}

func TestNegativeTrigger(t *testing.T) {
	const neg = `sets:
  - name: neg
    triggers:
      - {lemma: улица, type: division, negative: true, weight: 10}
`
	p, _ := newPipeline(t, testEntries(), neg)
	assertBrief(t, extract(t, p, Doc{Text: "улица Лягушкино"}).Spans)
	assertBrief(t, extract(t, p, Doc{Text: "деревня Лягушкино"}).Spans, "division:Лягушкино")
}
