package rules

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/textnorm"
)

type fakeLattice struct {
	terms []lexicon.Term
	spans map[int][]LatticeSpan
}

func (f *fakeLattice) Len() int                    { return len(f.terms) }
func (f *fakeLattice) Term(i int) *lexicon.Term    { return &f.terms[i] }
func (f *fakeLattice) SpansAt(i int) []LatticeSpan { return f.spans[i] }

// w is a word term; its case is title when raw starts with an upper-case letter.
func w(raw, lemma, tag string) lexicon.Term {
	c := textnorm.CaseLower
	if r, _ := utf8.DecodeRuneInString(raw); unicode.IsUpper(r) {
		c = textnorm.CaseTitle
	}
	t := lexicon.Term{
		Token: textnorm.Token{Raw: raw, Kind: textnorm.TokenWord, Case: c, Script: textnorm.ScriptCyrillic},
		Form:  strings.ToLower(raw),
	}
	if lemma != "" {
		t.Lemmas = []lexicon.Lemma{{Text: lemma, Tag: tag}}
	}
	return t
}

func punct(raw string) lexicon.Term {
	return lexicon.Term{Token: textnorm.Token{Raw: raw, Kind: textnorm.TokenPunct}}
}

func num(raw string) lexicon.Term {
	return lexicon.Term{Token: textnorm.Token{Raw: raw, Kind: textnorm.TokenNumber}, Form: raw}
}

func render(ms []PatternMatch, roles ...string) string {
	var parts []string
	for _, m := range ms {
		s := fmt.Sprintf("[%d,%d)", m.Start, m.End)
		for _, r := range roles {
			if c, ok := m.Roles[r]; ok {
				s += fmt.Sprintf(" %s=[%d,%d)#%d", r, c.Start, c.End, c.Span)
			}
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, " ")
}

// «крестьянин деревни Лягушкиной Иван Петров» with candidates 0 estate,
// 1 division (2 terms), 2 given_name, 3 surname.
func peasantLattice() *fakeLattice {
	return &fakeLattice{
		terms: []lexicon.Term{
			w("крестьянин", "крестьянин", "NOUN"), w("деревни", "деревня", "NOUN"),
			w("Лягушкиной", "лягушкина", "Geox"), w("Иван", "иван", "Name"), w("Петров", "петров", "Surn"),
		},
		spans: map[int][]LatticeSpan{
			0: {{ID: 0, Type: "estate", End: 1}},
			1: {{ID: 1, Type: "division", End: 3}},
			3: {{ID: 2, Type: "given_name", End: 4}},
			4: {{ID: 3, Type: "surname", End: 5}},
		},
	}
}

// «Иван Петров Сидоров Анна Михаилова Кузнецова»: two persons back to
// back. «Михаилова» has no candidate unless knownPatronymic.
func twoPersons(knownPatronymic bool) *fakeLattice {
	l := &fakeLattice{
		terms: []lexicon.Term{
			w("Иван", "иван", "Name"), w("Петров", "", ""), w("Сидоров", "", ""),
			w("Анна", "", ""), w("Михаилова", "", ""), w("Кузнецова", "", ""),
		},
		spans: map[int][]LatticeSpan{
			0: {{ID: 0, Type: "given_name", End: 1}},
			1: {{ID: 1, Type: "patronymic", End: 2}},
			2: {{ID: 2, Type: "surname", End: 3}},
			3: {{ID: 3, Type: "given_name", End: 4}},
			5: {{ID: 4, Type: "surname", End: 6}},
		},
	}
	if knownPatronymic {
		l.spans[4] = []LatticeSpan{{ID: 5, Type: "patronymic", End: 5}}
	}
	return l
}

const unknownPatronymic = `{name: p, elements: [{type: given_name},
	{token: word, shape: {case: title}, not: {type: given_name|patronymic|surname}, role: p},
	{type: surname}], actions: [{label: {role: p, type: patronymic}}]}`

const personGroup = `{name: p, elements: [{group: [{type: given_name}, {type: patronymic|surname},
	{type: surname, repeat: "?"}], role: who}], actions: [{label: {role: who, type: person}}]}`

func TestFindAllPeasantPatronymic(t *testing.T) {
	prog := mustPattern(t, `
name: peasant-patronymic
elements:
  - type: estate
  - {any: true, repeat: "*"}
  - type: given_name
  - {type: surname|patronymic, role: patr}
actions:
  - relabel: {role: patr, type: patronymic}
`)
	ms := prog.FindAll(peasantLattice())
	if got := render(ms, "patr"); got != "[0,5) patr=[4,5)#3" {
		t.Fatalf("got %s", got)
	}
	// Every candidate a type element consumed is reported, role or not.
	if !reflect.DeepEqual(ms[0].Spans, []int{0, 2, 3}) {
		t.Fatalf("consumed spans = %v", ms[0].Spans)
	}
}

func TestFindAllPersons(t *testing.T) {
	// The unknown-word slot takes «Михаилова» and never a typed word.
	ms := mustPattern(t, unknownPatronymic).FindAll(twoPersons(false))
	if got := render(ms, "p"); got != "[3,6) p=[4,5)#-1" {
		t.Fatalf("unknown patronymic: %s", got)
	}
	if !reflect.DeepEqual(ms[0].Spans, []int{3, 4}) {
		t.Fatalf("consumed spans = %v", ms[0].Spans)
	}
	// One match per person; the second needs its patronymic to be a span.
	if got := render(mustPattern(t, personGroup).FindAll(twoPersons(false)), "who"); got != "[0,3) who=[0,3)#-1" {
		t.Fatalf("persons, unknown patronymic: %s", got)
	}
	if got := render(mustPattern(t, personGroup).FindAll(twoPersons(true)), "who"); got != "[0,3) who=[0,3)#-1 [3,6) who=[3,6)#-1" {
		t.Fatalf("persons: %s", got)
	}
}

func TestFindAllElements(t *testing.T) {
	daughter := &fakeLattice{
		terms: []lexicon.Term{w("Мария", "мария", "Name"), punct(","), w("дочь", "дочь", "NOUN"),
			w("крестьянина", "крестьянин", "NOUN"), w("Ивана", "иван", "Name")},
		spans: map[int][]LatticeSpan{
			0: {{ID: 7, Type: "given_name", End: 1}},
			3: {{ID: 8, Type: "estate", End: 4}},
			4: {{ID: 9, Type: "given_name", End: 5}},
		},
	}
	age := &fakeLattice{terms: []lexicon.Term{num("25"), w("лет", "год", "NOUN"), punct(","), num("3"), w("года", "год", "NOUN")}}
	// «дер. Головина»: one division span that absorbed its keyword.
	absorbed := &fakeLattice{
		terms: []lexicon.Term{w("дер", "деревня", "NOUN"), punct("."), w("Головина", "", "")},
		spans: map[int][]LatticeSpan{0: {{ID: 0, Type: "division", End: 3}}},
	}
	bare := &fakeLattice{
		terms: []lexicon.Term{w("Головина", "", "")},
		spans: map[int][]LatticeSpan{0: {{ID: 0, Type: "division", End: 1}}},
	}
	// «Иван Анна Кузнецова»: the slot after a given name must not take a given name.
	names := &fakeLattice{
		terms: []lexicon.Term{w("Иван", "", ""), w("Анна", "", ""), w("Кузнецова", "", "")},
		spans: map[int][]LatticeSpan{
			0: {{ID: 0, Type: "given_name", End: 1}},
			1: {{ID: 1, Type: "given_name", End: 2}},
			2: {{ID: 2, Type: "surname", End: 3}},
		},
	}
	cases := []struct {
		name, pattern string
		l             Lattice
		roles         []string
		want          string
	}{
		{"optional punct, optional span", `{name: p, elements: [{type: given_name, role: child},
			{token: punct, repeat: "?"}, {lemma: сын|дочь}, {type: estate, repeat: "?"},
			{type: given_name, role: parent}], actions: [{boost: {role: child}}]}`,
			daughter, []string{"child", "parent"}, "[0,5) child=[0,1)#7 parent=[4,5)#9"},
		{"group role captures terms", `{name: p, elements: [{group: [{token: number}, {lemma: год}], role: age}],
			actions: [{label: {role: age, type: age}}]}`,
			age, []string{"age"}, "[0,2) age=[0,2)#-1 [3,5) age=[3,5)#-1"},
		{"grammeme", `{name: p, elements: [{grammeme: Name, role: n}], actions: [{label: {role: n, type: t}}]}`,
			daughter, []string{"n"}, "[0,1) n=[0,1)#-1 [4,5) n=[4,5)#-1"},
		{"plus", `{name: p, elements: [{token: word|number, repeat: "+", role: w}], actions: [{label: {role: w, type: t}}]}`,
			age, []string{"w"}, "[0,2) w=[1,2)#-1 [3,5) w=[4,5)#-1"},
		{"empty-loop guard", `{name: p, elements: [{group: [{any: true, repeat: "?"}], repeat: "*"},
			{lemma: год, role: y}], actions: [{label: {role: y, type: t}}]}`,
			age, []string{"y"}, "[0,5) y=[4,5)#-1"},
		{"no match", `{name: p, elements: [{lemma: сын}, {type: given_name, role: x}], actions: [{boost: {role: x}}]}`,
			daughter, []string{"x"}, ""},
		{"span words fail the shape", `{name: p, elements: [{type: division, shape: {case: title}, role: s}], actions: [{boost: {role: s}}]}`,
			absorbed, []string{"s"}, ""},
		{"span words pass the shape", `{name: p, elements: [{type: division, shape: {case: title}, role: s}], actions: [{boost: {role: s}}]}`,
			bare, []string{"s"}, "[0,1) s=[0,1)#0"},
		{"not a given name", `{name: p, elements: [{type: given_name},
			{token: word, shape: {case: title}, not: {type: given_name}, role: x}], actions: [{label: {role: x, type: t}}]}`,
			names, []string{"x"}, "[1,3) x=[2,3)#-1"},
		{"not alone consumes a term", `{name: p, elements: [{type: given_name}, {not: {lemma: сын|дочь}, role: x}],
			actions: [{label: {role: x, type: t}}]}`,
			twoPersons(false), []string{"x"}, "[0,2) x=[1,2)#-1 [3,5) x=[4,5)#-1"},
		{"not blocks a term", `{name: p, elements: [{type: given_name}, {token: punct}, {not: {lemma: сын|дочь}, role: x}],
			actions: [{label: {role: x, type: t}}]}`,
			daughter, []string{"x"}, ""},
	}
	for _, c := range cases {
		if got := render(mustPattern(t, c.pattern).FindAll(c.l), c.roles...); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}
