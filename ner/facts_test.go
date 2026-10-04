package ner

import (
	"reflect"
	"testing"

	"github.com/amarin/lexicon/gazetteer"
)

// factRules: a person over given name + surname, then kinship and age
// facts whose arguments are a whole person when there is one.
const factRules = `
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
  - name: kinship
    patterns:
      # «Мария, дочь крестьянина Ивана» → child_of(child, parent).
      - name: child-of
        elements:
          - {type: person|given_name, role: child}
          - {token: punct, repeat: "?"}
          - lemma: сын|дочь
          - {type: estate, repeat: "?"}
          - {type: person|given_name, role: parent}
        actions:
          - emit: {kind: child_of, args: {child: child, parent: parent}}
      # «Иван Петров, 25 лет» → an age span over «25» and age(person, age).
      - name: age
        elements:
          - {type: person|given_name, role: person}
          - {token: punct, repeat: "?"}
          - {token: number, role: age}
          - lemma: год
        actions:
          - label: {role: age, type: age}
          - emit: {kind: age, args: {person: person, age: age}}
`

func TestFactChildOf(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), factRules)
	res := extract(t, p, Doc{Text: "Мария, дочь крестьянина Ивана"})
	assertBrief(t, res.Spans, "given_name:Мария", "estate:крестьянина", "given_name:Ивана")
	want := []Fact{{Kind: "child_of", Args: map[string]int{"child": 0, "parent": 2}, Rule: "child-of"}}
	if !reflect.DeepEqual(res.Facts, want) {
		t.Fatalf("facts = %+v", res.Facts)
	}
}

// A fact argument can be a whole person (decision P13).
func TestFactPersonArgument(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), factRules, personNesting)
	res := extract(t, p, Doc{Text: "Мария, дочь Ивана Петрова"})
	assertBrief(t, res.Spans, "given_name:Мария", "person:Ивана Петрова", "given_name:Ивана", "surname:Петрова")
	want := []Fact{{Kind: "child_of", Args: map[string]int{"child": 0, "parent": 1}, Rule: "child-of"}}
	if !reflect.DeepEqual(res.Facts, want) {
		t.Fatalf("facts = %+v", res.Facts)
	}
}

func TestFactAge(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), factRules, personNesting)
	res := extract(t, p, Doc{Text: "Иван Петров, 25 лет"})
	assertBrief(t, res.Spans, "person:Иван Петров", "given_name:Иван", "surname:Петров", "age:25")
	want := []Fact{{Kind: "age", Args: map[string]int{"person": 0, "age": 3}, Rule: "age"}}
	if !reflect.DeepEqual(res.Facts, want) {
		t.Fatalf("facts = %+v (spans %q)", res.Facts, brief(res.Spans))
	}
}

func TestFactDroppedWhenArgumentLoses(t *testing.T) {
	entries := append(testEntries(), gazetteer.Entry{Alias: "Иван Петров", Type: "ship", Ref: "ship:1", Canonical: "Иван Петров"})
	p, _ := newPipeline(t, entries, factRules)
	res := extract(t, p, Doc{Text: "Мария, дочь Ивана Петрова"})
	assertBrief(t, res.Spans, "given_name:Мария", "ship:Ивана Петрова")
	if len(res.Facts) != 0 {
		t.Fatalf("fact with a lost argument survived: %+v", res.Facts)
	}
}

func TestFactsFollowTypesFilter(t *testing.T) {
	p, _ := newPipeline(t, testEntries(), factRules)
	const text = "Мария, дочь крестьянина Ивана"
	res := extract(t, p, Doc{Text: text, Types: []string{"given_name"}})
	if len(res.Facts) != 1 || !reflect.DeepEqual(res.Facts[0].Args, map[string]int{"child": 0, "parent": 1}) {
		t.Fatalf("facts = %+v", res.Facts)
	}
	if res := extract(t, p, Doc{Text: text, Types: []string{"estate"}}); len(res.Facts) != 0 {
		t.Fatalf("facts over filtered spans: %+v", res.Facts)
	}
}

func TestFactsDeduplicated(t *testing.T) {
	const twice = factRules + `  - name: kinship-copy
    patterns:
      - name: child-of
        elements:
          - {type: given_name, role: child}
          - {token: punct, repeat: "?"}
          - lemma: сын|дочь
          - {type: estate, repeat: "?"}
          - {type: given_name, role: parent}
        actions:
          - emit: {kind: child_of, args: {child: child, parent: parent}}
`
	p, _ := newPipeline(t, testEntries(), twice)
	if res := extract(t, p, Doc{Text: "Мария, дочь крестьянина Ивана"}); len(res.Facts) != 1 {
		t.Fatalf("facts = %+v", res.Facts)
	}
}
