package rules

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// sampleRules is written in idiomatic YAML (block style, comments); the
// YAML file carries the provenance manifest as a top-level meta mapping.
const sampleRules = `# Rules used by the rules package tests.
meta:
  source: lexicon tests
  license: CC0-1.0
sets:
  # Always active: place keywords.
  - name: places
    hints:
      - {lemma: деревня, type: division, dir: right, window: 1, weight: 2, absorb: true}
      - {lemma: ул, dotted: true, type: street} # defaults: dir right, window 1, weight 1
    triggers:
      - lemma: деревня|село
        type: division
        window: 1..2
        shape: {case: title, script: cyrillic}
        stop_at: [punct, stop]
        weight: 1
  # Active only for documents tagged period:pre1917.
  - name: pre1917
    when: ["period:pre1917"]
    hints:
      - {lemma: крестьянин, type: surname, dir: right, window: 2}
`

// sampleRulesJSON is sampleRules as a JSON document: JSON is valid YAML,
// so rule files written as JSON keep loading.
const sampleRulesJSON = `{
  "meta": {"source": "lexicon tests", "license": "CC0-1.0"},
  "sets": [
    {"name": "places",
     "hints": [
       {"lemma": "деревня", "type": "division", "dir": "right", "window": 1, "weight": 2, "absorb": true},
       {"lemma": "ул", "dotted": true, "type": "street"}
     ],
     "triggers": [
       {"lemma": "деревня|село", "type": "division", "window": "1..2",
        "shape": {"case": "title", "script": "cyrillic"}, "stop_at": ["punct", "stop"], "weight": 1}
     ]},
    {"name": "pre1917", "when": ["period:pre1917"],
     "hints": [{"lemma": "крестьянин", "type": "surname", "dir": "right", "window": 2}]}
  ]
}`

func TestLoad(t *testing.T) {
	f, err := Load(strings.NewReader(sampleRules))
	if err != nil {
		t.Fatal(err)
	}
	if f.Meta["source"] != "lexicon tests" || len(f.Sets) != 2 {
		t.Fatalf("file = %+v", f)
	}
	places := f.Sets[0]
	if h := places.Hints[0]; h.Lemma != "деревня" || h.Dir != Right || h.Window != 1 || h.Weight != 2 || !h.Absorb {
		t.Fatalf("hint 0 = %+v", h)
	}
	if h := places.Hints[1]; !h.Dotted || h.Dir != Right || h.Window != 0 {
		t.Fatalf("hint 1 defaults = %+v", h)
	}
	tr := places.Triggers[0]
	if tr.Lemma != "деревня|село" || tr.Window != (Window{1, 2}) || tr.Shape.Case != "title" || len(tr.StopAt) != 2 {
		t.Fatalf("trigger = %+v", tr)
	}
	if got := f.Sets[1].When; len(got) != 1 || got[0] != "period:pre1917" {
		t.Fatalf("when = %v", got)
	}
}

// Every set, hint and trigger remembers its source line (decision D15);
// JSON documents get lines too.
func TestLoadPositions(t *testing.T) {
	f, err := LoadNamed("places.yaml", strings.NewReader(sampleRules))
	if err != nil {
		t.Fatal(err)
	}
	places, pre := f.Sets[0], f.Sets[1]
	got := []int{places.line, places.Hints[0].line, places.Hints[1].line, places.Triggers[0].line, pre.line, pre.Hints[0].line}
	if want := []int{7, 9, 10, 12, 19, 22}; !reflect.DeepEqual(got, want) || f.name != "places.yaml" {
		t.Fatalf("lines = %v, name = %q; want %v, places.yaml", got, f.name, want)
	}
	j, err := Load(strings.NewReader(sampleRulesJSON))
	if err != nil {
		t.Fatal(err)
	}
	if got := []int{j.Sets[0].line, j.Sets[0].Triggers[0].line, j.Sets[1].line}; !reflect.DeepEqual(got, []int{4, 10, 13}) {
		t.Fatalf("JSON lines = %v", got)
	}
}

func TestLoadJSON(t *testing.T) {
	fy, err := Load(strings.NewReader(sampleRules))
	if err != nil {
		t.Fatal(err)
	}
	fj, err := Load(strings.NewReader(sampleRulesJSON))
	if err != nil {
		t.Fatalf("JSON document must load: %v", err)
	}
	// Source lines differ between the two layouts; the canonical encoding
	// (exported fields only, as in Book.Version) must not.
	ey, err := yaml.Marshal(fy)
	if err != nil {
		t.Fatal(err)
	}
	ej, err := yaml.Marshal(fj)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(ey, ej) {
		t.Fatalf("JSON and YAML differ:\n%s\n%s", ej, ey)
	}
}

func TestLoadErrors(t *testing.T) {
	for name, src := range map[string]string{
		"unknown field":  `sets: [{name: x, colour: red}]`,
		"bad direction":  `sets: [{name: x, hints: [{lemma: a, type: t, dir: up}]}]`,
		"bad window":     `sets: [{name: x, triggers: [{lemma: a, type: t, window: 1-3}]}]`,
		"window mapping": `sets: [{name: x, triggers: [{lemma: a, type: t, window: {min: 1, max: 2}}]}]`,
		"not a mapping":  "- a\n- b\n",
		"empty":          "",
		"two documents":  "sets: []\n---\nsets: []\n",
		"json trailing":  `{"sets": []} {"sets": []}`,
	} {
		if _, err := Load(strings.NewReader(src)); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}

// Load errors read "<file>:<line>: <message>" (decision D15).
func TestLoadErrorLines(t *testing.T) {
	for name, tc := range map[string]struct {
		src  string
		want string
	}{
		"unknown field": {"sets:\n  - name: x\n    colour: red\n", "rules.yaml:3: field colour not found"},
		"bad direction": {"sets:\n  - name: x\n    hints:\n      - lemma: a\n        type: t\n        dir: up\n", `rules.yaml:6: unknown direction "up"`},
		"bad window":    {"sets:\n  - name: x\n    triggers:\n      - lemma: a\n        type: t\n        window: 1-3\n", `rules.yaml:6: window: bad value "1-3"`},
		"syntax":        {"sets:\n  - name: x\n    hints: [\n", "rules.yaml:"},
		"two documents": {"sets: []\n---\nsets: []\n", "rules.yaml:2: a rule file holds one YAML document"}, // the "---" line
		"empty":         {"", "rules.yaml: empty rule file"},
	} {
		_, err := LoadNamed("rules.yaml", strings.NewReader(tc.src))
		if err == nil || !strings.HasPrefix(err.Error(), tc.want) {
			t.Errorf("%s: err = %v, want prefix %q", name, err, tc.want)
		}
	}
	if _, err := Load(strings.NewReader("sets:\n  - name: x\n    colour: red\n")); err == nil || !strings.HasPrefix(err.Error(), "<input>:3: ") {
		t.Errorf("unnamed input: err = %v, want prefix <input>:3:", err)
	}
}

func TestWindowForms(t *testing.T) {
	for src, want := range map[string]Window{
		`sets: [{name: x, triggers: [{lemma: a, type: t, window: 3}]}]`:    {1, 3},
		`sets: [{name: x, triggers: [{lemma: a, type: t, window: 2..4}]}]`: {2, 4},
		`sets: [{name: x, triggers: [{lemma: a, type: t, window: "2"}]}]`:  {1, 2},
	} {
		f, err := Load(strings.NewReader(src))
		if err != nil || f.Sets[0].Triggers[0].Window != want {
			t.Errorf("%s: %+v %v", src, f, err)
		}
	}
}

func TestLoadFile(t *testing.T) {
	dir := t.TempDir()
	for name, data := range map[string]string{"r.yaml": sampleRules, "r.yml": sampleRules, "r.json": sampleRulesJSON} {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
		if f, err := LoadFile(path); err != nil || len(f.Sets) != 2 {
			t.Fatalf("LoadFile(%s): %+v %v", name, f, err)
		}
	}
	bad := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(bad, []byte("sets:\n  - name: x\n    colour: red\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadFile(bad); err == nil || !strings.HasPrefix(err.Error(), bad+":3: ") {
		t.Fatalf("bad file: err = %v, want %s:3: prefix", err, bad)
	}
	if _, err := LoadFile(filepath.Join(dir, "missing.yaml")); err == nil {
		t.Fatal("missing file: no error")
	}
}
