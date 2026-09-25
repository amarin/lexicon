package main

import (
	"bytes"
	"testing"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/textnorm"
)

func TestFormatTerms(t *testing.T) {
	var buf bytes.Buffer

	formatTerms(&buf, []lexicon.Term{
		{Token: textnorm.Token{Raw: "с"}, Form: "с.", Lemmas: []lexicon.Lemma{
			{Text: "село", Flags: lexicon.FlagAbbrev | lexicon.FlagAmbiguous},
			{Text: "сын", Flags: lexicon.FlagAbbrev | lexicon.FlagAmbiguous},
		}},
		{Token: textnorm.Token{Raw: "."}},
		{Token: textnorm.Token{Raw: "Кота"}, Form: "кота", Lemmas: []lexicon.Lemma{{Text: "кот"}}},
	})

	want := "с\tс.\tсело[ambiguous,abbrev] сын[ambiguous,abbrev]\n.\t\t\nКота\tкота\tкот\n"
	if buf.String() != want {
		t.Fatalf("formatTerms =\n%q\nwant\n%q", buf.String(), want)
	}
}
