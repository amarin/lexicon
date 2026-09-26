package rules

import (
	"reflect"
	"strings"
	"testing"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/textnorm"
)

func mustBook(t *testing.T, src string) *Book {
	t.Helper()
	f, err := Load(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	b, err := Compile(f)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func term(raw string, dotted bool, kind textnorm.TokenKind, c textnorm.Case, lemmas ...lexicon.Lemma) *lexicon.Term {
	return &lexicon.Term{
		Token:  textnorm.Token{Raw: raw, Kind: kind, Case: c, Script: textnorm.ScriptCyrillic, Dotted: dotted},
		Form:   strings.ToLower(raw),
		Lemmas: lemmas,
	}
}

func TestBookActive(t *testing.T) {
	b := mustBook(t, sampleRules)
	if got := b.Active(nil); !reflect.DeepEqual(got.Sets, []string{"places"}) || len(got.Hints) != 2 || len(got.Triggers) != 1 {
		t.Fatalf("no tags: %+v", got)
	}
	got := b.Active([]string{"period:pre1917", "record:birth"})
	if !reflect.DeepEqual(got.Sets, []string{"places", "pre1917"}) || len(got.Hints) != 3 || got.Hints[2].Set != "pre1917" {
		t.Fatalf("with tags: %+v", got)
	}
	if h := got.Hints[1]; h.Window != 1 || h.Weight != 1 {
		t.Fatalf("hint defaults not applied: %+v", h)
	}
	if !reflect.DeepEqual(b.Sets(), []string{"places", "pre1917"}) {
		t.Fatalf("Sets() = %v", b.Sets())
	}
}

func TestBookVersion(t *testing.T) {
	a, b := mustBook(t, sampleRules), mustBook(t, sampleRules)
	if a.Version() == "" || a.Version() != b.Version() {
		t.Fatal("version must be stable")
	}
	c := mustBook(t, strings.Replace(sampleRules, `weight: 2`, `weight: 3`, 1))
	if c.Version() == a.Version() {
		t.Fatal("version must change with the rules")
	}
	empty, err := Compile()
	if err != nil || empty.Version() == "" || len(empty.Active(nil).Hints) != 0 {
		t.Fatalf("empty book: %v", err)
	}
}

func TestKeywordsAndShapes(t *testing.T) {
	b := mustBook(t, sampleRules)
	act := b.Active(nil)
	derevnya, ul := act.Hints[0], act.Hints[1]
	tr := act.Triggers[0]

	full := term("деревни", false, textnorm.TokenWord, textnorm.CaseLower, lexicon.Lemma{Text: "деревня"})
	abbr := term("дер", true, textnorm.TokenWord, textnorm.CaseLower, lexicon.Lemma{Text: "деревня", Flags: lexicon.FlagAbbrev})
	if !derevnya.Keyword(full) || !derevnya.Keyword(abbr) || !tr.Keyword(full) {
		t.Fatal("lemma keyword must match inflected and abbreviated forms")
	}
	dotted := term("ул", true, textnorm.TokenWord, textnorm.CaseLower, lexicon.Lemma{Text: "улица", Flags: lexicon.FlagAbbrev})
	// Mimic v0.1's Analyzer.fullTerm, which gives a dotted abbreviation
	// Term.Form the trailing dot (e.g. "ул.") — pins keyword.match's
	// TrimSuffix behaviour against real Form values, not just the helper's.
	dotted.Form = "ул."
	undotted := term("ул", false, textnorm.TokenWord, textnorm.CaseLower)
	if !ul.Keyword(dotted) || ul.Keyword(undotted) {
		t.Fatal("dotted hint must require the dot")
	}
	unknown := term("мир", false, textnorm.TokenWord, textnorm.CaseLower, lexicon.Lemma{Text: "деревня", Flags: lexicon.FlagUnknown})
	if derevnya.Keyword(unknown) {
		t.Fatal("unknown lemmas must not match")
	}

	title := term("Сидоровке", false, textnorm.TokenWord, textnorm.CaseTitle, lexicon.Lemma{Text: "сидоровка", Flags: lexicon.FlagPredicted})
	lower := term("жил", false, textnorm.TokenWord, textnorm.CaseLower)
	stop := term("В", false, textnorm.TokenWord, textnorm.CaseTitle, lexicon.Lemma{Text: "в", Flags: lexicon.FlagStop})
	if !tr.Accepts(title) || tr.Accepts(lower) || tr.Accepts(stop) {
		t.Fatal("shape/stop checks are wrong")
	}
}

func TestCompileErrors(t *testing.T) {
	for name, src := range map[string]string{
		"empty set name":  `sets: [{name: ""}]`,
		"duplicate set":   `sets: [{name: a}, {name: a}]`,
		"hint type":       `sets: [{name: a, hints: [{lemma: ул}]}]`,
		"hint lemma":      `sets: [{name: a, hints: [{lemma: " | ", type: street}]}]`,
		"hint window":     `sets: [{name: a, hints: [{lemma: ул, type: street, window: 9}]}]`,
		"trigger both":    `sets: [{name: a, triggers: [{lemma: село, type: division, dir: both}]}]`,
		"trigger range":   `sets: [{name: a, triggers: [{lemma: село, type: division, window: 3..2}]}]`,
		"trigger shape":   `sets: [{name: a, triggers: [{lemma: село, type: division, shape: {case: camel}}]}]`,
		"trigger stop_at": `sets: [{name: a, triggers: [{lemma: село, type: division, stop_at: [comma]}]}]`,
	} {
		f, err := Load(strings.NewReader(src))
		if err != nil {
			t.Fatalf("%s: load: %v", name, err)
		}
		// Flow-style sources are one line: every error points at <input>:1.
		if _, err := Compile(f); err == nil || !strings.HasPrefix(err.Error(), "<input>:1: ") {
			t.Errorf("%s: err = %v, want prefix <input>:1:", name, err)
		}
	}
}

// Compile errors read "<file>:<line>: <set>/<rule>: <message>" (decision D15).
func TestCompileErrorLines(t *testing.T) {
	load := func(name, src string) File {
		t.Helper()
		f, err := LoadNamed(name, strings.NewReader(src))
		if err != nil {
			t.Fatal(err)
		}
		return f
	}
	for name, tc := range map[string]struct {
		files []File
		want  string
	}{
		"hint window": {[]File{load("places.yaml",
			"sets:\n  - name: places\n    hints:\n      - {lemma: деревня, type: division}\n      - {lemma: ул, type: street, window: 9}\n")},
			"places.yaml:5: places/hint 1: window 9 out of 1..8"},
		"trigger dir": {[]File{load("",
			"sets:\n  - name: s\n    triggers:\n      - lemma: село\n        type: division\n        dir: both\n")},
			"<input>:4: s/trigger 0: triggers need dir left or right"},
		"empty set name": {[]File{load("a.yaml", "sets:\n  - name: a\n  - when: [x]\n")},
			"a.yaml:3: #1: empty set name"},
		"duplicate across files": {[]File{load("a.yaml", "sets:\n  - name: places\n"), load("b.yaml", "sets:\n  - {name: other}\n  - name: places\n")},
			"b.yaml:3: places: duplicate set name (first defined at a.yaml:2)"},
		"built in Go": {[]File{{Sets: []RuleSet{{Name: "g", Hints: []Hint{{Lemma: "ул"}}}}}},
			"<input>: g/hint 0: empty type"},
	} {
		if _, err := Compile(tc.files...); err == nil || err.Error() != tc.want {
			t.Errorf("%s: err = %v, want %q", name, err, tc.want)
		}
	}
}
