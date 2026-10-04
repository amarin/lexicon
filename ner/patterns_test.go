package ner

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/amarin/lexicon/gazetteer"
)

// patternRules = place hints plus pattern sets over the testEntries types.
const patternRules = `meta:
  source: lexicon tests
sets:
  - name: places
    hints:
      - {lemma: деревня, type: division, dir: right, window: 1, weight: 2, absorb: true}
      - {lemma: уезд, type: division, dir: left, window: 1, weight: 2, absorb: true}

  # Peasant records before 1917: in «крестьянин … Иван Петров» the word after
  # the given name is a short patronymic, not a surname.
  - name: peasant-naming
    when: ["period:pre1917", "estate:peasant"]
    patterns:
      - name: peasant-patronymic
        elements:
          - type: estate
          - {any: true, repeat: "*"} # repeat markers are quoted: * and ? are YAML indicators
          - type: given_name
          - {type: surname|patronymic, role: patr}
        actions:
          - relabel: {role: patr, type: patronymic}

  # «Иван Петров, 25 лет» → an age span over «25».
  - name: ages
    patterns:
      - name: age
        elements:
          - {type: given_name}
          - {type: surname|patronymic, repeat: "?"}
          - {token: punct, repeat: "?"}
          - {token: number, role: age}
          - lemma: год
        actions:
          - label: {role: age, type: age}
`

var peasantTags = []string{"period:pre1917", "estate:peasant"}

const peasantText = "крестьянин деревни Лягушкиной Иван Петров"

func TestPeasantPatronymicRelabel(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), patternRules)

	res := extract(t, p, Doc{Text: peasantText})
	assertBrief(t, res.Spans, "estate:крестьянин", "division:деревни Лягушкиной", "given_name:Иван", "surname:Петров")

	res = extract(t, p, Doc{Text: peasantText, Tags: peasantTags}, Explain())
	assertBrief(t, res.Spans, "estate:крестьянин", "division:деревни Лягушкиной", "given_name:Иван", "patronymic:Петров")
	sp := spanOf(t, res.Spans, "patronymic", "Петров")
	if !sp.Flags.Has(Candidate) || sp.Flags.Has(Ambiguous) || sp.Refs != nil ||
		!reflect.DeepEqual(sp.Normal, []string{"петров"}) || sp.Score != 4 {
		t.Fatalf("relabelled span = %+v", sp)
	}
	if !slices.Contains(sp.Evidence, "pattern peasant-patronymic relabel surname → patronymic") {
		t.Fatalf("evidence = %q", sp.Evidence)
	}
	// The surname reading is kept as an alternative with its ref.
	if len(sp.Alternatives) != 1 {
		t.Fatalf("alternatives = %+v", sp.Alternatives)
	}
	alt := sp.Alternatives[0]
	if alt.Type != "surname" || !reflect.DeepEqual(alt.Refs, []string{"surname:1"}) ||
		!reflect.DeepEqual(alt.Normal, []string{"Петров"}) || alt.Score != 3 ||
		!slices.Contains(alt.Evidence, "relabelled by pattern peasant-patronymic") {
		t.Fatalf("alternative = %+v", alt)
	}
}

func TestRelabelPrefersExistingTarget(t *testing.T) {
	entries := append(testEntries(), gazetteer.Entry{Alias: "Петров", Type: "patronymic", Ref: "patronymic:7", Canonical: "Петрович"})
	p, _ := newPipeline(t, entries, patternRules)

	surname := Alternative{Type: "surname", Refs: []string{"surname:1"}, Normal: []string{"Петров"}, Score: 3}

	// Without the peasant context both readings score the same: the tie is
	// flagged and the surname is an alternative (D7, D14).
	res := extract(t, p, Doc{Text: peasantText})
	tie := spanOf(t, res.Spans, "patronymic", "Петров")
	if !tie.Flags.Has(Ambiguous) || !reflect.DeepEqual(tie.Alternatives, []Alternative{surname}) {
		t.Fatalf("tie = %+v", tie)
	}

	// With it the existing patronymic wins with its own ref; the surname
	// reading stays as the only alternative, without Ambiguous.
	res = extract(t, p, Doc{Text: peasantText, Tags: peasantTags})
	sp := spanOf(t, res.Spans, "patronymic", "Петров")
	if sp.Flags.Has(Candidate) || sp.Flags.Has(Ambiguous) || !reflect.DeepEqual(sp.Refs, []string{"patronymic:7"}) ||
		!reflect.DeepEqual(sp.Normal, []string{"Петрович"}) || sp.Score != 4 ||
		!reflect.DeepEqual(sp.Alternatives, []Alternative{surname}) {
		t.Fatalf("existing patronymic must win with its ref: %+v", sp)
	}
}

func TestPatternBoost(t *testing.T) {
	const boost = `
sets:
  - name: b
    patterns:
      - name: estate-boost
        elements: [{type: estate, role: e}]
        actions: [{boost: {role: e, weight: 2}}]
`
	p, _ := newPipeline(t, testEntries(), boost)
	res := extract(t, p, Doc{Text: "крестьянин"}, Explain())
	sp := spanOf(t, res.Spans, "estate", "крестьянин")
	if sp.Score != 5 || !slices.Contains(sp.Evidence, "pattern estate-boost boost +2") {
		t.Fatalf("boosted span = %+v", sp)
	}
}

func TestLabelCreatesSpan(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), patternRules)
	res := extract(t, p, Doc{Text: "Иван Петров, 25 лет"}, Explain())
	assertBrief(t, res.Spans, "given_name:Иван", "surname:Петров", "age:25")
	age := spanOf(t, res.Spans, "age", "25")
	if !age.Flags.Has(Candidate) || !reflect.DeepEqual(age.Normal, []string{"25"}) || age.Refs != nil ||
		age.Alternatives != nil || age.Score != 2 || !slices.Contains(age.Evidence, "labelled by pattern age") {
		t.Fatalf("age span = %+v", age)
	}
}

const surnameAfterGiven = `
sets:
  - name: s
    patterns:
      - name: after-given
        elements:
          - {type: given_name}
          - {token: word, role: s}
        actions:
          - label: {role: s, type: TYPE}
`

// A label over a range that already has a candidate of the type boosts it
// instead of creating a duplicate.
func TestLabelKeepsExistingCandidate(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), strings.Replace(surnameAfterGiven, "TYPE", "surname", 1))
	res := extract(t, p, Doc{Text: "Иван Петров"})
	assertBrief(t, res.Spans, "given_name:Иван", "surname:Петров")
	sp := spanOf(t, res.Spans, "surname", "Петров")
	if sp.Flags.Has(Candidate) || !reflect.DeepEqual(sp.Refs, []string{"surname:1"}) || sp.Score != 4 {
		t.Fatalf("surname = %+v", sp)
	}
}

// A Blocked alias vetoes its type on the range for labels too.
func TestLabelRespectsBlocked(t *testing.T) {
	rules := strings.Replace(surnameAfterGiven, "TYPE", "nickname", 1)
	p, _ := newPipeline(t, testEntries(), rules)
	sp := spanOf(t, extract(t, p, Doc{Text: "Иван Петров"}).Spans, "surname", "Петров")
	if len(sp.Alternatives) != 1 || sp.Alternatives[0].Type != "nickname" {
		t.Fatalf("the labelled nickname must lose to the surname and stay an alternative: %+v", sp)
	}
	blocked := append(testEntries(), gazetteer.Entry{Alias: "Петров", Type: "nickname", Ref: "nickname:9", Flags: gazetteer.Blocked})
	p, _ = newPipeline(t, blocked, rules)
	sp = spanOf(t, extract(t, p, Doc{Text: "Иван Петров"}).Spans, "surname", "Петров")
	if len(sp.Alternatives) != 0 {
		t.Fatalf("a blocked type was labelled: %+v", sp)
	}
}

const personOverNames = `
sets:
  - name: persons
    patterns:
      - name: person
        elements:
          - group:
              - {type: given_name}
              - {type: surname}
            role: who
        actions:
          - label: {role: who, type: person}
`

func personNesting(c *Config) {
	c.Nesting = map[string][]string{"person": {"given_name", "patronymic", "surname"}}
}

// A label over dictionary spans is a composite, not a Candidate; with the
// nesting pairs the parts are output inside it.
func TestLabelOverSpansIsNotCandidate(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), personOverNames, personNesting)
	res := extract(t, p, Doc{Text: "Иван Петров"})
	assertBrief(t, res.Spans, "person:Иван Петров", "given_name:Иван", "surname:Петров")
	who := spanOf(t, res.Spans, "person", "Иван Петров")
	if who.Flags.Has(Candidate) || who.Flags.Has(Nested) || who.Refs != nil || who.Score != 3.5 {
		t.Fatalf("person = %+v", who)
	}
	if g := spanOf(t, res.Spans, "given_name", "Иван"); !g.Flags.Has(Nested) || !reflect.DeepEqual(g.Refs, []string{"given_name:1"}) {
		t.Fatalf("given name = %+v", g)
	}
	// Without the nesting pairs the parts outscore the whole.
	p, _ = newPipeline(t, testEntries(), personOverNames)
	assertBrief(t, extract(t, p, Doc{Text: "Иван Петров"}).Spans, "given_name:Иван", "surname:Петров")
}

// A part that requires context (a surname that is also a common word) is
// kept when a pattern assembles a span over it: the match is its context.
func TestLabelGivesContextToParts(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), personOverNames, personNesting)
	res := extract(t, p, Doc{Text: "Иван Мороз"}, Explain())
	assertBrief(t, res.Spans, "person:Иван Мороз", "given_name:Иван", "surname:Мороз")
	if who := res.Spans[0]; !reflect.DeepEqual(who.Normal, []string{"Иван Мороз"}) {
		t.Fatalf("person = %+v", who)
	}
	sur := res.Spans[2]
	if !reflect.DeepEqual(sur.Refs, []string{"surname:2"}) || !slices.Contains(sur.Evidence, "pattern person gives context") {
		t.Fatalf("surname = %+v", sur)
	}
	// No pattern match, no context.
	assertBrief(t, extract(t, p, Doc{Text: "ударил мороз"}).Spans)
}

// Patterns run per sentence: a match in a later sentence gets offsets into
// the whole text.
func TestPatternInLaterSentence(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), patternRules)
	const text = "Он пришёл. Иван Петров, 25 лет"
	res := extract(t, p, Doc{Text: text})
	assertBrief(t, res.Spans, "given_name:Иван", "surname:Петров", "age:25")
	age := res.Spans[2]
	start := strings.Index(text, "25")
	if age.Start != start || age.End != start+2 || age.RuneStart != 24 || age.RuneEnd != 26 || text[age.Start:age.End] != "25" {
		t.Fatalf("age = %+v", age)
	}
}

// Doc.Types filters the output after resolution: without the person its
// parts stay (still Nested), without the parts the person stays.
func TestTypesFilterOnComposite(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), personOverNames, personNesting)
	res := extract(t, p, Doc{Text: "Иван Петров", Types: []string{"given_name", "surname"}})
	assertBrief(t, res.Spans, "given_name:Иван", "surname:Петров")
	for _, sp := range res.Spans {
		if !sp.Flags.Has(Nested) {
			t.Fatalf("part = %+v", sp)
		}
	}
	res = extract(t, p, Doc{Text: "Иван Петров", Types: []string{"person"}})
	assertBrief(t, res.Spans, "person:Иван Петров")
	if who := res.Spans[0]; who.Flags.Has(Nested) || !reflect.DeepEqual(who.Normal, []string{"Иван Петров"}) {
		t.Fatalf("person = %+v", who)
	}
}
