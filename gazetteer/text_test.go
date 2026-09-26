package gazetteer

import (
	"testing"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/internal/fakedict"
)

func prepared(t *testing.T, text string) *Text {
	t.Helper()
	return Prepare(fakedict.NewAnalyzer().Analyze(text, fakedict.Profiles()["text"], lexicon.ModeFull))
}

func TestTextBreaks(t *testing.T) {
	cases := []struct {
		text   string
		n      int
		breaks []bool
	}{
		{"дер. Лягушкиной Иван", 3, []bool{false, false, true}}, // abbreviation dot is transparent
		{"Иван Петров. Мария", 3, []bool{false, true, true}},    // sentence end
		{"Иван, Петров", 2, []bool{true, true}},                 // comma breaks
		{"И. Петров", 2, []bool{false, true}},                   // initial
	}
	for _, c := range cases {
		tx := prepared(t, c.text)
		if tx.Len() != c.n {
			t.Fatalf("%q: Len() = %d, want %d", c.text, tx.Len(), c.n)
		}
		for p, want := range c.breaks {
			if got := tx.Break(p); got != want {
				t.Errorf("%q: Break(%d) = %v, want %v", c.text, p, got, want)
			}
		}
	}
	tx := prepared(t, "дер. Лягушкиной Иван")
	if !tx.Joined(0, 2) || !tx.Joined(1, 1) {
		t.Fatal("Joined must follow Break")
	}
	if tx.Terms()[tx.TermIndex(1)].Token.Raw != "Лягушкиной" {
		t.Fatal("TermIndex must skip the dot term")
	}
}

func TestTextSentences(t *testing.T) {
	tx := prepared(t, "Жил в дер. Лягушкиной Иван Петров. Мария")
	terms := tx.Terms()
	segs := tx.Sentences()
	if len(segs) != 2 || segs[0][0] != 0 || segs[len(segs)-1][1] != len(terms) {
		t.Fatalf("sentences = %v", segs)
	}
	for i := 1; i < len(segs); i++ {
		if segs[i][0] != segs[i-1][1] {
			t.Fatalf("sentences must be contiguous: %v", segs)
		}
	}
	find := func(raw string) int {
		for i, tm := range terms {
			if tm.Token.Raw == raw {
				return i
			}
		}
		t.Fatalf("no term %q", raw)
		return -1
	}
	in := func(i int) int {
		for k, s := range segs {
			if i >= s[0] && i < s[1] {
				return k
			}
		}
		return -1
	}
	if in(find("дер")) != in(find("Петров")) || in(find("Петров")) == in(find("Мария")) {
		t.Fatalf("wrong split: %v", segs)
	}
}
