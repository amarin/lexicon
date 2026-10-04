package rules

import (
	"bytes"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// testPatterns is a host-style rule file used across the pattern tests.
// Line numbers matter: TestLoadPatterns checks them.
const testPatterns = `meta:
  source: lexicon tests
sets:
  - name: persons
    patterns:
      - name: unknown-patronymic
        elements:
          - {type: given_name}
          - {token: word, shape: {case: title}, not: {type: given_name|patronymic|surname}, role: p}
          - {type: surname}
        actions:
          - label: {role: p, type: patronymic}
      - name: person
        elements:
          - group:
              - {type: given_name}
              - {type: patronymic|surname}
              - {type: surname, repeat: "?"} # repeat markers are quoted: * and ? are YAML indicators
            role: who
        actions:
          - label: {role: who, type: person}
  - name: peasant-naming
    when: ["period:pre1917", "estate:peasant"]
    patterns:
      - name: peasant-patronymic
        elements:
          - {type: given_name}
          - {type: surname, role: patr}
        actions:
          - relabel: {role: patr, type: patronymic}
  - name: kinship
    patterns:
      - name: age
        elements:
          - {type: given_name, role: person}
          - {token: punct, repeat: "?"}
          - {token: number, role: age}
          - lemma: год
        actions:
          - label: {role: age, type: age}
          - emit: {kind: age, args: {person: person, age: age}}
`

func TestLoadPatterns(t *testing.T) {
	f, err := Load(strings.NewReader(testPatterns))
	if err != nil {
		t.Fatal(err)
	}
	p := f.Sets[0].Patterns[0]
	if p.Name != "unknown-patronymic" || len(p.Elements) != 3 {
		t.Fatalf("pattern = %+v", p)
	}
	e := p.Elements[1]
	if e.Token != "word" || e.Shape.Case != "title" || e.Role != "p" ||
		e.Not == nil || e.Not.Type != "given_name|patronymic|surname" {
		t.Fatalf("element = %+v", e)
	}
	if p.line != 6 || e.line != 9 || e.Not.line != 9 {
		t.Fatalf("lines: pattern %d, element %d, not %d; want 6, 9, 9", p.line, e.line, e.Not.line)
	}
	if l := p.Actions[0].Label; l == nil || l.Role != "p" || l.Type != "patronymic" {
		t.Fatalf("label = %+v", p.Actions[0])
	}
	grp := f.Sets[0].Patterns[1].Elements[0]
	if len(grp.Group) != 3 || grp.Role != "who" || grp.Group[2].Repeat != RepeatOptional || grp.Group[2].line != 18 {
		t.Fatalf("group = %+v", grp)
	}
	if got := f.Sets[1].Patterns[0]; got.line != 25 || got.Actions[0].Relabel == nil {
		t.Fatalf("peasant-patronymic = %+v", got)
	}
	age := f.Sets[2].Patterns[0]
	if em := age.Actions[1].Emit; em == nil || em.Kind != "age" || em.Args["age"] != "age" {
		t.Fatalf("emit = %+v", age.Actions[1])
	}
	if !e.hasCondition() || grp.hasCondition() || !(Element{Not: &Element{Type: "a"}}).hasCondition() {
		t.Fatal("hasCondition is wrong")
	}
	// Repeat and Not survive a marshal round trip (Book.Version hashes this encoding).
	out, err := yaml.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	back, err := Load(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("round trip: %v\n%s", err, out)
	}
	if back.Sets[0].Patterns[1].Elements[0].Group[2].Repeat != RepeatOptional || back.Sets[0].Patterns[0].Elements[1].Not == nil {
		t.Fatalf("round trip lost data:\n%s", out)
	}
}

// An unquoted * is a YAML alias indicator: the file must fail to load with
// a line number instead of silently producing an empty repeat.
func TestLoadPatternsUnquotedRepeat(t *testing.T) {
	src := "sets:\n  - name: s\n    patterns:\n      - name: p\n        elements:\n          - {any: true, repeat: *}\n        actions:\n          - boost: {role: x}\n"
	if _, err := Load(strings.NewReader(src)); err == nil || !strings.HasPrefix(err.Error(), "<input>:6: ") {
		t.Fatalf("err = %v, want an <input>:6: error", err)
	}
}

// Repeat is typed: an unknown marker fails at load time with file and line.
func TestLoadPatternsBadRepeat(t *testing.T) {
	for name, tc := range map[string]struct{ src, want string }{
		"braces":  {"sets:\n  - name: s\n    patterns:\n      - name: p\n        elements:\n          - {any: true, repeat: \"{2}\"}\n        actions:\n          - boost: {role: x}\n", `kin.yaml:6: unknown repeat "{2}"`},
		"word":    {"sets:\n  - name: s\n    patterns:\n      - name: p\n        elements:\n          - any: true\n            repeat: optional\n        actions:\n          - boost: {role: x}\n", `kin.yaml:7: unknown repeat "optional"`},
		"mapping": {"sets:\n  - name: s\n    patterns:\n      - name: p\n        elements:\n          - {any: true, repeat: {min: 2}}\n        actions:\n          - boost: {role: x}\n", "kin.yaml:6: repeat: want"},
	} {
		if _, err := LoadNamed("kin.yaml", strings.NewReader(tc.src)); err == nil || !strings.HasPrefix(err.Error(), tc.want) {
			t.Errorf("%s: err = %v, want prefix %q", name, err, tc.want)
		}
	}
}
