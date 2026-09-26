package gazetteer

import (
	"reflect"
	"testing"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/textnorm"
)

func TestAlternatives(t *testing.T) {
	word := func(form string, lemmas ...lexicon.Lemma) *lexicon.Term {
		return &lexicon.Term{Token: textnorm.Token{Raw: form, Kind: textnorm.TokenWord}, Form: form, Lemmas: lemmas}
	}
	cases := []struct {
		name string
		term *lexicon.Term
		want []string
	}{
		{"known", word("лягушкиной", lexicon.Lemma{Text: "лягушкина"}), []string{"лягушкина"}},
		{"dedup", word("петров", lexicon.Lemma{Text: "петров"}, lexicon.Lemma{Text: "петров", Kind: "surname"}), []string{"петров"}},
		{"ambiguous", word("с", lexicon.Lemma{Text: "село"}, lexicon.Lemma{Text: "сын"}), []string{"село", "сын"}},
		{"unknown flag", word("сидоровке", lexicon.Lemma{Text: "сидоровке", Flags: lexicon.FlagUnknown}), []string{"сидоровке"}},
		{"no lemmas", word("иваново"), []string{"иваново"}},
		{"punct", &lexicon.Term{Token: textnorm.Token{Raw: ",", Kind: textnorm.TokenPunct}}, nil},
	}
	for _, c := range cases {
		if got := Alternatives(c.term); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestCombine(t *testing.T) {
	out, capped := Combine([][]string{{"село", "сын"}, {"иваново"}}, MaxLemmaKeys)
	want := [][]string{{"село", "иваново"}, {"сын", "иваново"}}
	if capped || !reflect.DeepEqual(out, want) {
		t.Fatalf("got %v %v", out, capped)
	}
	four := [][]string{{"1", "2"}, {"1", "2"}, {"1", "2"}, {"1", "2"}}
	out, capped = Combine(four, 8)
	if !capped || len(out) != 8 || !reflect.DeepEqual(out[7], []string{"1", "2", "2", "2"}) {
		t.Fatalf("capped: %v %v", out, capped)
	}
	out, capped = Combine(four[:3], 8)
	if capped || len(out) != 8 {
		t.Fatalf("exactly 8 must not be capped: %d %v", len(out), capped)
	}
	if out, _ := Combine([][]string{{"a"}, {}}, 8); out != nil {
		t.Fatalf("empty position must yield nothing: %v", out)
	}
}
