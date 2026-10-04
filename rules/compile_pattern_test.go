package rules

import (
	"reflect"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// mustPattern decodes one pattern written in YAML (flow style keeps the
// table-driven cases on one line) and compiles it.
func mustPattern(t *testing.T, src string) *Program {
	t.Helper()
	var p Pattern
	if err := yaml.Unmarshal([]byte(src), &p); err != nil {
		t.Fatal(err)
	}
	prog, err := compilePattern("test", p)
	if err != nil {
		t.Fatal(err)
	}
	return prog
}

func TestCompilePattern(t *testing.T) {
	prog := mustPattern(t, `
name: age
elements:
  - {type: given_name, role: person}
  - {token: number, role: age}
  - lemma: год
actions:
  - label: {role: age, type: age}
  - boost: {role: person}
  - emit: {kind: age, args: {person: person, age: age}}
`)
	if prog.Name() != "age" || prog.Set() != "test" || !reflect.DeepEqual(prog.Roles(), []string{"person", "age"}) {
		t.Fatalf("program = %+v", prog)
	}
	acts := prog.Actions()
	if acts[0].Label.Weight != 1 || acts[1].Boost.Weight != 1 {
		t.Fatalf("default weights not applied: %+v %+v", acts[0].Label, acts[1].Boost)
	}
}

// Conditions combine (decision P14).
func TestCompilePatternConditions(t *testing.T) {
	for name, src := range map[string]string{
		"conjunction": `{name: p, elements: [{token: word, shape: {case: title}, not: {type: a|b}, role: x}], actions: [{label: {role: x, type: t}}]}`,
		"type+shape":  `{name: p, elements: [{type: a, shape: {case: title}, role: x}], actions: [{boost: {role: x}}]}`,
		"not alone":   `{name: p, elements: [{not: {lemma: сын}, role: x}], actions: [{label: {role: x, type: t}}]}`,
		"group label": `{name: p, elements: [{group: [{type: a}, {type: b}], role: x}], actions: [{label: {role: x, type: t}}, {emit: {kind: k, args: {who: x}}}]}`,
	} {
		var p Pattern
		if err := yaml.Unmarshal([]byte(src), &p); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if _, err := compilePattern("test", p); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestCompilePatternErrors(t *testing.T) {
	for name, src := range map[string]string{
		"no name":           `{elements: [{any: true}], actions: [{boost: {role: x}}]}`,
		"no elements":       `{name: p, actions: [{boost: {role: x}}]}`,
		"no actions":        `{name: p, elements: [{any: true}]}`,
		"any+type":          `{name: p, elements: [{any: true, type: a, role: x}], actions: [{boost: {role: x}}]}`,
		"empty element":     `{name: p, elements: [{role: x}], actions: [{label: {role: x, type: t}}]}`,
		"group+condition":   `{name: p, elements: [{any: true, group: [{any: true}]}], actions: [{boost: {role: x}}]}`,
		"bad token kind":    `{name: p, elements: [{token: emoji, role: x}], actions: [{label: {role: x, type: t}}]}`,
		"bad shape":         `{name: p, elements: [{shape: {case: camel}, role: x}], actions: [{label: {role: x, type: t}}]}`,
		"not with role":     `{name: p, elements: [{token: word, not: {type: a, role: y}, role: x}], actions: [{label: {role: x, type: t}}]}`,
		"not with repeat":   `{name: p, elements: [{token: word, not: {type: a, repeat: "?"}, role: x}], actions: [{label: {role: x, type: t}}]}`,
		"empty not":         `{name: p, elements: [{token: word, not: {}, role: x}], actions: [{label: {role: x, type: t}}]}`,
		"nested not":        `{name: p, elements: [{token: word, not: {not: {type: a}}, role: x}], actions: [{label: {role: x, type: t}}]}`,
		"unknown role":      `{name: p, elements: [{type: a, role: x}], actions: [{boost: {role: y}}]}`,
		"relabel no type":   `{name: p, elements: [{type: a, role: x}], actions: [{relabel: {role: x}}]}`,
		"relabel term role": `{name: p, elements: [{token: word, role: x}], actions: [{relabel: {role: x, type: t}}]}`,
		"label span role":   `{name: p, elements: [{type: a, role: x}], actions: [{label: {role: x, type: t}}]}`,
		"label no type":     `{name: p, elements: [{token: word, role: x}], actions: [{label: {role: x}}]}`,
		"two actions":       `{name: p, elements: [{type: a, role: x}], actions: [{boost: {role: x}, relabel: {role: x, type: b}}]}`,
		"emit term role":    `{name: p, elements: [{token: number, role: n}], actions: [{emit: {kind: k, args: {a: n}}}]}`,
		"emit no args":      `{name: p, elements: [{type: a, role: x}], actions: [{emit: {kind: k}}]}`,
		"emit no kind":      `{name: p, elements: [{type: a, role: x}], actions: [{emit: {args: {a: x}}}]}`,
		"boost term role":   `{name: p, elements: [{token: number, role: n}], actions: [{boost: {role: n}}]}`,
	} {
		var p Pattern
		if err := yaml.Unmarshal([]byte(src), &p); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if _, err := compilePattern("test", p); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
	// Unknown repeat markers fail at load (TestLoadPatternsBadRepeat); a
	// Repeat value built in Go is still checked.
	bad := Pattern{Name: "p", Elements: []Element{{Any: true, Repeat: Repeat(9)}}, Actions: []Action{{Boost: &Boost{Role: "x"}}}}
	if _, err := compilePattern("test", bad); err == nil || !strings.Contains(err.Error(), "repeat") {
		t.Errorf("repeat built in Go: err = %v", err)
	}
}

// Pattern compile errors name file, line, set and pattern (decision P11):
// the offending element's line (inside groups and not too), else the pattern's.
func TestCompilePatternErrorLines(t *testing.T) {
	for name, tc := range map[string]struct{ src, want string }{
		"element in group": {"sets:\n  - name: kinship\n    patterns:\n      - name: child-of\n        elements:\n          - {type: given_name, role: child}\n          - group:\n              - {token: emoji}\n        actions:\n          - boost: {role: child}\n",
			`kin.yaml:8: kinship/pattern "child-of": unknown token kind "emoji" (want word, number, punct or symbol)`},
		"not": {"sets:\n  - name: s\n    patterns:\n      - name: p\n        elements:\n          - token: word\n            not: {type: a, role: y}\n            role: x\n        actions:\n          - label: {role: x, type: t}\n",
			`kin.yaml:7: s/pattern "p": not takes only type, lemma, grammeme, token and shape`},
		"no actions": {"sets:\n  - name: kinship\n    patterns:\n      - name: no-actions\n        elements:\n          - {any: true}\n",
			`kin.yaml:4: kinship/pattern "no-actions": no actions`},
		"unnamed": {"sets:\n  - name: s\n    patterns:\n      - elements: [{any: true}]\n        actions: [{boost: {role: x}}]\n",
			"kin.yaml:4: s/pattern 0: pattern without name"},
	} {
		f, err := LoadNamed("kin.yaml", strings.NewReader(tc.src))
		if err != nil {
			t.Fatalf("%s: load: %v", name, err)
		}
		if _, err := Compile(f); err == nil || err.Error() != tc.want {
			t.Errorf("%s: err = %v, want %q", name, err, tc.want)
		}
	}
}

func TestBookActivatesPatternsByTags(t *testing.T) {
	f, err := Load(strings.NewReader(testPatterns))
	if err != nil {
		t.Fatal(err)
	}
	b, err := Compile(f)
	if err != nil {
		t.Fatal(err)
	}
	names := func(ps []*Program) []string {
		var out []string
		for _, p := range ps {
			out = append(out, p.Name())
		}
		return out
	}
	base := []string{"unknown-patronymic", "person", "age"}
	if got := names(b.Active(nil).Patterns); !reflect.DeepEqual(got, base) {
		t.Fatalf("no tags: %v", got)
	}
	if got := names(b.Active([]string{"period:pre1917"}).Patterns); !reflect.DeepEqual(got, base) {
		t.Fatalf("one of two tags must not activate peasant-naming: %v", got)
	}
	got := names(b.Active([]string{"estate:peasant", "period:pre1917"}).Patterns)
	if !reflect.DeepEqual(got, []string{"unknown-patronymic", "person", "peasant-patronymic", "age"}) {
		t.Fatalf("all tags: %v", got)
	}
}
