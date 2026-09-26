package ner

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/amarin/lexicon/gazetteer"
)

func resolveOf(nesting map[string]map[string]bool, cands ...*candidate) []string {
	s := &state{p: &Pipeline{nesting: nesting}, cands: cands}
	chosen, nested := s.resolve()
	var out []string
	for _, c := range chosen {
		mark := ""
		if nested[c] {
			mark = " nested"
		}
		if c.flags.Has(Ambiguous) {
			mark += " ambiguous"
		}
		var alts []string
		for _, a := range s.alternatives(c) {
			alts = append(alts, a.Type)
		}
		if len(alts) > 0 {
			mark += " alt=" + strings.Join(alts, ",")
		}
		out = append(out, fmt.Sprintf("%s[%d,%d)%s", c.typ, c.start, c.end, mark))
	}
	return out
}

func cand(typ string, start, end int, score float64) *candidate {
	return &candidate{typ: typ, start: start, end: end, score: score}
}

func TestResolve(t *testing.T) {
	cases := []struct {
		name    string
		nesting map[string]map[string]bool
		cands   []*candidate
		want    []string
	}{
		{"fragments beat a weaker whole", nil,
			[]*candidate{cand("a", 0, 2, 5), cand("b", 0, 1, 3), cand("c", 1, 2, 3)},
			[]string{"b[0,1)", "c[1,2)"}},
		{"crossing spans conflict", nil,
			[]*candidate{cand("a", 0, 2, 4), cand("b", 1, 3, 3)},
			[]string{"a[0,2)"}},
		{"allowed nesting keeps both", map[string]map[string]bool{"x": {"y": true}},
			[]*candidate{cand("x", 0, 3, 5), cand("y", 1, 2, 2)},
			[]string{"x[0,3)", "y[1,2) nested"}},
		{"nesting only for configured pairs", map[string]map[string]bool{"x": {"z": true}},
			[]*candidate{cand("x", 0, 3, 5), cand("y", 1, 2, 2)},
			[]string{"x[0,3)"}},
		{"non-positive scores never win", nil,
			[]*candidate{cand("a", 0, 1, 0), cand("b", 1, 2, -1)},
			nil},
		{"type tie is flagged, the loser is an alternative", nil,
			[]*candidate{cand("surname", 0, 1, 3), cand("patronymic", 0, 1, 3)},
			[]string{"patronymic[0,1) ambiguous alt=surname"}},
		{"lower-scored types on the same range are alternatives, best first", nil,
			[]*candidate{cand("given_name", 0, 1, 1), cand("surname", 0, 1, 3), cand("patronymic", 0, 1, 2)},
			[]string{"surname[0,1) alt=patronymic,given_name"}},
		{"equal-scored alternatives are ordered by type", nil,
			[]*candidate{cand("x", 0, 1, 5), cand("c", 0, 1, 1), cand("b", 0, 1, 1)},
			[]string{"x[0,1) alt=b,c"}},
		{"other ranges are never alternatives", nil,
			[]*candidate{cand("a", 0, 2, 4), cand("b", 1, 3, 3), cand("c", 0, 1, 0.5)},
			[]string{"a[0,2)"}},
		{"non-positive losers are not alternatives", nil,
			[]*candidate{cand("a", 0, 1, 2), cand("b", 0, 1, 0)},
			[]string{"a[0,1)"}},
	}
	for _, c := range cases {
		if got := resolveOf(c.nesting, c.cands...); fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

// TestResolveFloatNoiseTie: two candidates whose scores are mathematically
// equal but arrive by different runtime float64 additions (a classic
// 0.1+0.2 vs 0.3 case) must still be treated as an exact tie, not decided
// by float noise in the last bits.
func TestResolveFloatNoiseTie(t *testing.T) {
	a, b := 0.1, 0.2
	c1 := &candidate{typ: "patronymic", start: 0, end: 1, bonus: a + b}
	c2 := &candidate{typ: "surname", start: 0, end: 1, bonus: 0.3}
	if c1.bonus == c2.bonus {
		t.Fatalf("test setup: want raw float64 noise, got exact equality (%v == %v)", c1.bonus, c2.bonus)
	}
	s := &state{p: &Pipeline{weights: Weights{}}, cands: []*candidate{c1, c2}, explain: true}
	s.scoreAll()
	if c1.score != c2.score {
		t.Fatalf("quantized scores must be exactly equal: %v vs %v", c1.score, c2.score)
	}
	chosen, _ := s.resolve()
	if len(chosen) != 1 || chosen[0] != c1 {
		t.Fatalf("want patronymic as the sole, deterministic winner: %+v", chosen)
	}
	if !c1.flags.Has(Ambiguous) {
		t.Fatalf("tie must set Ambiguous on the winner: %+v", c1)
	}
	alts := s.alternatives(c1)
	if len(alts) != 1 || alts[0].Type != "surname" {
		t.Fatalf("loser must be kept as an alternative of the winner: %+v", alts)
	}
	found := false
	for _, e := range c1.evidence {
		if e == "tie with surname" {
			found = true
		}
	}
	if !found {
		t.Fatalf("winner must record the tie: %+v", c1.evidence)
	}
}

func TestExtractResolvesOverlaps(t *testing.T) {
	entries := append(testEntries(),
		gazetteer.Entry{Alias: "Иван Петров", Type: "ship", Ref: "ship:1", Canonical: "Иван Петров"},
		gazetteer.Entry{Alias: "Вера", Type: "given_name", Ref: "given_name:4", Canonical: "Вера"},
	)
	p, _ := newPipeline(t, entries, "")
	assertBrief(t, extract(t, p, Doc{Text: "Иван Петров"}).Spans, "ship:Иван Петров")
	// The Types filter runs after resolution: the losing given_name is not resurrected.
	assertBrief(t, extract(t, p, Doc{Text: "Иван Петров", Types: []string{"given_name"}}).Spans)

	res := extract(t, p, Doc{Text: "Вера Петрова"})
	assertBrief(t, res.Spans, "given_name:Вера", "surname:Петрова")
	vera := res.Spans[0]
	if !vera.Flags.Has(Ambiguous) || len(vera.Refs) != 2 || vera.Score != 2.75 {
		t.Fatalf("ambiguous span = %+v", vera)
	}
	if s := res.Spans[1].Score; s != 2 {
		t.Fatalf("lemma-only one-word score = %v, want 2", s)
	}
}

func TestExtractAlternatives(t *testing.T) {
	entries := append(testEntries(),
		gazetteer.Entry{Alias: "Петров", Type: "patronymic", Ref: "patronymic:1", Canonical: "Петрович"})

	// Equal scores: deterministic winner by type name, flagged, the loser kept.
	p, _ := newPipeline(t, entries, "")
	res := extract(t, p, Doc{Text: "Иван Петров"})
	assertBrief(t, res.Spans, "given_name:Иван", "patronymic:Петров")
	sp := res.Spans[1]
	want := []Alternative{{Type: "surname", Refs: []string{"surname:1"}, Normal: []string{"Петров"}, Score: 3}}
	if !sp.Flags.Has(Ambiguous) || !reflect.DeepEqual(sp.Alternatives, want) {
		t.Fatalf("tie = %+v", sp)
	}
	if res.Spans[0].Alternatives != nil {
		t.Fatalf("unrivalled span has alternatives: %+v", res.Spans[0])
	}

	// Explain fills the alternative's evidence and notes the tie on the winner.
	sp = extract(t, p, Doc{Text: "Иван Петров"}, Explain()).Spans[1]
	if !slices.Contains(sp.Alternatives[0].Evidence, "surface match «Петров» (test)") ||
		!slices.Contains(sp.Evidence, "tie with surname") {
		t.Fatalf("explained tie = %+v", sp)
	}

	// A type weight breaks the tie: the loser is still an alternative, the winner is not ambiguous.
	p, _ = newPipeline(t, entries, "", func(c *Config) { c.Weights.Types = map[string]float32{"surname": 0.5} })
	res = extract(t, p, Doc{Text: "Иван Петров"})
	sp = spanOf(t, res.Spans, "surname", "Петров")
	want = []Alternative{{Type: "patronymic", Refs: []string{"patronymic:1"}, Normal: []string{"Петрович"}, Score: 3}}
	if sp.Flags.Has(Ambiguous) || sp.Score != 3.5 || !reflect.DeepEqual(sp.Alternatives, want) {
		t.Fatalf("weighted winner = %+v", sp)
	}
	// The Types filter looks at the winner only: an alternative's type does not keep the span.
	assertBrief(t, extract(t, p, Doc{Text: "Иван Петров", Types: []string{"patronymic"}}).Spans)
}
