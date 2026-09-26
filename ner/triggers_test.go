package ner

import (
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/amarin/lexicon/gazetteer"
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
	if !slices.Contains(sp.Evidence, "trigger «деревне» → division candidate +1") {
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

// TestTriggerRespectsBlocked: a Blocked alias vetoes its type on the range
// it matched. A trigger must not resurrect a fresh candidate of that type
// over the same words, even though filterEarly already removed the blocked
// candidate before applyTriggers runs.
func TestTriggerRespectsBlocked(t *testing.T) {
	entries := append(testEntries(), gazetteer.Entry{
		Alias: "Сидоровке", Type: "division", Flags: gazetteer.Blocked,
	})
	p, _ := newPipeline(t, entries, placesRules)
	assertBrief(t, extract(t, p, Doc{Text: "в деревне Сидоровке"}).Spans)
}

// TestTriggerDirLeft covers collect's left branch: the keyword follows the
// candidate words instead of preceding them.
func TestTriggerDirLeft(t *testing.T) {
	const left = `sets:
  - name: left
    triggers:
      - {lemma: волость, type: division, dir: left, shape: {case: title, script: cyrillic}}
`
	p, _ := newPipeline(t, testEntries(), left)
	res := extract(t, p, Doc{Text: "Сидоровская волость"}, Explain())
	assertBrief(t, res.Spans, "division:Сидоровская")
	sp := res.Spans[0]
	if !slices.Contains(sp.Evidence, "trigger «волость» → division candidate +1") {
		t.Fatalf("evidence = %q", sp.Evidence)
	}
}

// TestTriggerWindowStopsAtPunctuation: a break between the keyword and the
// next word (a comma, or a sentence boundary) stops the walk before it
// collects anything, so no candidate is proposed.
func TestTriggerWindowStopsAtPunctuation(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), placesRules)
	assertBrief(t, extract(t, p, Doc{Text: "Жил в деревне. Сидоровка"}).Spans)
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

// TestTriggerWeightScoresCandidate: the rule's weight is evidence for the
// candidate it proposes, as for a boosted gazetteer span.
func TestTriggerWeightScoresCandidate(t *testing.T) {
	const rule = `sets:
  - name: s
    triggers:
      - {lemma: село, type: division, dir: right, shape: {case: title, script: cyrillic}, weight: %s}
`
	score := func(w string) float32 {
		p, _ := newPipeline(t, testEntries(), fmt.Sprintf(rule, w))
		return spanOf(t, extract(t, p, Doc{Text: "из села Покровское"}).Spans, "division", "Покровское").Score
	}
	if one, five := score("1"), score("5"); five-one != 4 {
		t.Fatalf("weight 1 → %v, weight 5 → %v: want a difference of 4", one, five)
	}
}
