package ner

import (
	"context"
	"reflect"
	"testing"

	"github.com/amarin/lexicon/gazetteer"
	"github.com/amarin/lexicon/internal/fakedict"
	"github.com/amarin/lexicon/rules"
)

// personsPipeline builds a pipeline over nertest's person fixtures.
func personsPipeline(t testing.TB, nest bool) *Pipeline {
	t.Helper()
	an := fakedict.NewAnalyzer()
	gz, err := gazetteer.New(context.Background(), gazetteer.Config{
		Analyzer: an, TypeProfiles: fakedict.TypeProfiles(), DefaultProfile: fakedict.Profiles()["text"],
		Sources: []gazetteer.Source{gazetteer.NewTSVSource("persons", "../nertest/testdata/persons.tsv")},
	})
	if err != nil {
		t.Fatal(err)
	}
	rf, err := rules.LoadFile("../nertest/testdata/persons.rules.yaml")
	if err != nil {
		t.Fatal(err)
	}
	book, err := rules.Compile(rf)
	if err != nil {
		t.Fatal(err)
	}
	cfg := Config{Analyzer: an, Gazetteer: gz, Rules: book, Profiles: fakedict.Profiles(), DefaultProfile: "text"}
	if nest {
		personNesting(&cfg)
	}
	p, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestPersonParts(t *testing.T) {
	p := personsPipeline(t, true)

	// Dictionary parts keep their refs; the person is a composite.
	res := extract(t, p, Doc{Text: "Иван Петров Сидоров"})
	assertBrief(t, res.Spans, "person:Иван Петров Сидоров", "given_name:Иван", "patronymic:Петров", "surname:Сидоров")
	who := res.Spans[0]
	if who.Flags.Has(Candidate) || who.Refs != nil || !reflect.DeepEqual(who.Normal, []string{"Иван Петров Сидоров"}) {
		t.Fatalf("person = %+v", who)
	}
	for i, ref := range []string{"given_name:1", "patronymic:1", "surname:10"} {
		if part := res.Spans[i+1]; !part.Flags.Has(Nested) || !reflect.DeepEqual(part.Refs, []string{ref}) {
			t.Fatalf("part %d = %+v", i, part)
		}
	}

	// An unknown word inside a name is a Candidate part without refs.
	res = extract(t, p, Doc{Text: "Анна Михаилова Кузнецова"})
	if part := spanOf(t, res.Spans, "patronymic", "Михаилова"); !part.Flags.Has(Candidate) || part.Refs != nil {
		t.Fatalf("unknown patronymic = %+v", part)
	}
	if who := spanOf(t, res.Spans, "person", "Анна Михаилова Кузнецова"); who.Flags.Has(Candidate) {
		t.Fatalf("person over known parts must not be a candidate: %+v", who)
	}

	// A surname that is also a place name keeps the place as an alternative.
	res = extract(t, p, Doc{Text: "Анна Петрова Головина"})
	sur := spanOf(t, res.Spans, "surname", "Головина")
	if !sur.Flags.Has(Candidate) || sur.Flags.Has(Ambiguous) || len(sur.Alternatives) != 1 ||
		sur.Alternatives[0].Type != "division" || !reflect.DeepEqual(sur.Alternatives[0].Refs, []string{"division:9"}) {
		t.Fatalf("surname = %+v", sur)
	}

	// A given-name reading in the patronymic slot keeps the name as an alternative.
	res = extract(t, p, Doc{Text: "Алексей Степанов Сидоров"})
	patr := spanOf(t, res.Spans, "patronymic", "Степанов")
	if len(patr.Alternatives) != 1 || patr.Alternatives[0].Type != "given_name" ||
		!reflect.DeepEqual(patr.Alternatives[0].Refs, []string{"given_name:6"}) {
		t.Fatalf("patronymic = %+v", patr)
	}
}

// Without the host's nesting pairs the parts are output and the person is not.
func TestPersonsWithoutNesting(t *testing.T) {
	p := personsPipeline(t, false)
	assertBrief(t, extract(t, p, Doc{Text: "Иван Петров Сидоров"}).Spans,
		"given_name:Иван", "patronymic:Петров", "surname:Сидоров")
}
