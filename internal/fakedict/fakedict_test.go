package fakedict

import (
	"testing"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/textnorm"
)

func hasLemma(t lexicon.Term, text string, flag lexicon.Flag) bool {
	for _, l := range t.Lemmas {
		if l.Text == text && l.Flags&flag == flag {
			return true
		}
	}
	return false
}

func TestParsePredictsOnlyWithoutExactReading(t *testing.T) {
	d := Genealogy()
	rs := d.Parse("Сидоровке", AllKinds())
	if len(rs) != 1 || !rs[0].Predicted || rs[0].Normal != "сидоровка" {
		t.Fatalf("Сидоровке: %+v", rs)
	}
	rs = d.Parse("Лягушкиной", AllKinds())
	if len(rs) != 1 || rs[0].Predicted || rs[0].Normal != "лягушкина" {
		t.Fatalf("Лягушкиной: %+v", rs)
	}
	if rs := d.Parse("Сидоровке", []lexicon.Kind{KindToponym}); rs != nil {
		t.Fatalf("prediction without base kind: %+v", rs)
	}
}

// TestAnalyzerContract pins the v0.1 behaviour this milestone relies on
// (assumptions A2-A3 of the v0.2 plan). If it fails, reconcile with v0.1
// before implementing anything else.
func TestAnalyzerContract(t *testing.T) {
	const text = "дер. Лягушкиной Иван, 25 лет."
	terms := NewAnalyzer().Analyze(text, Profiles()["text"], lexicon.ModeFull)
	got := map[string]lexicon.Term{}
	for _, tm := range terms {
		if text[tm.Token.Start:tm.Token.End] != tm.Token.Raw {
			t.Fatalf("offsets of %q are wrong: %+v", tm.Token.Raw, tm.Token)
		}
		got[tm.Token.Raw] = tm
	}
	if der := got["дер"]; !der.Token.Dotted || !hasLemma(der, "деревня", lexicon.FlagAbbrev) {
		t.Fatalf("дер: %+v", der)
	}
	if dot, ok := got["."]; !ok || dot.Token.Kind != textnorm.TokenPunct {
		t.Fatalf("ModeFull must keep '.' as a Punct term: %+v", terms)
	}
	if _, ok := got[","]; !ok {
		t.Fatalf("ModeFull must keep ',' as a term: %+v", terms)
	}
	if !hasLemma(got["Лягушкиной"], "лягушкина", 0) {
		t.Fatalf("Лягушкиной: %+v", got["Лягушкиной"])
	}
	if iv := got["Иван"]; iv.Token.Case != textnorm.CaseTitle || iv.Token.Script != textnorm.ScriptCyrillic || !hasLemma(iv, "иван", 0) {
		t.Fatalf("Иван: %+v", iv)
	}
	if n := got["25"]; n.Token.Kind != textnorm.TokenNumber {
		t.Fatalf("25: %+v", n)
	}
	if !hasLemma(got["лет"], "год", 0) {
		t.Fatalf("лет: %+v", got["лет"])
	}
}
