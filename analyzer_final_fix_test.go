package lexicon

import (
	"testing"

	"github.com/amarin/lexicon/textnorm"
)

// abbrFake has the real-OpenCorpora shape of one-letter service words: besides
// the service reading there is a letter-name abbreviation reading (Abbr).
var abbrFake = fakeDicts{
	"в":   {rBase("в", "PREP"), rBase("в", "NOUN,inan,masc,Fixd,Abbr sing,nomn")},
	"и":   {rBase("и", "CONJ"), rBase("и", "NOUN,inan,neut,Fixd,Abbr sing,nomn")},
	"я":   {rBase("я", "NOUN,inan,neut,Fixd,Abbr sing,nomn")},
	"кот": {rBase("кот", "NOUN,anim,masc sing,nomn")},
	// «боровскаго» is not in the base; its modern variant is, but only as a
	// non-place reading; the base still predicts junk for the original form.
	"боровскаго": {rBasePredicted("боровскаго", "NOUN,inan,masc,Fixd,Geox sing,gent")},
	"боровского": {rBase("боровский", "ADJF,Qual masc,sing,gent")},
	// An exact reading of the form itself, filtered out by the place profile.
	"стали": {rBase("сталь", "NOUN,inan,femn sing,gent"), rBasePredicted("сталя", "NOUN,Geox sing,gent")},
}

// TestStopIgnoresAbbr: readings tagged Abbr do not keep a service word out of
// the stop words.
func TestStopIgnoresAbbr(t *testing.T) {
	a := NewAnalyzer(abbrFake, textnorm.PreReform, AnalyzerOptions{})

	if got := show(a.Analyze("кот в и я", profText, ModeIndex)); got != "кот=кот[]; я=я[]" {
		t.Fatalf("ModeIndex = %q", got)
	}

	terms := a.Analyze("в", profText, ModeFull)
	if len(terms) != 1 || len(terms[0].Lemmas) != 1 || terms[0].Lemmas[0].Tag != "PREP" || terms[0].Lemmas[0].Flags != FlagStop {
		t.Fatalf("ModeFull(в) = %+v", terms)
	}
}

// TestNoPredictionsAfterExact: once the form or a reform variant has exact
// readings, the profile filter removing them all yields the unknown lemma,
// never predictions.
func TestNoPredictionsAfterExact(t *testing.T) {
	a := NewAnalyzer(abbrFake, textnorm.PreReform, AnalyzerOptions{})

	cases := []struct{ text, want string }{
		{"Боровскаго", "боровскаго=боровскаго[unknown]"},
		{"Боровского", "боровского=боровского[unknown]"},
		{"стали", "стали=стали[unknown]"},
	}
	for _, c := range cases {
		if got := show(a.Analyze(c.text, profPlace, ModeIndex)); got != c.want {
			t.Errorf("Analyze(%q, place) = %q, want %q", c.text, got, c.want)
		}
	}

	if got := show(a.Analyze("Боровскаго", profText, ModeIndex)); got != "боровскаго=боровский[reform]" {
		t.Errorf("Analyze(Боровскаго, text) = %q", got)
	}
}
