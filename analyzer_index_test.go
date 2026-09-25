package lexicon

import (
	"fmt"
	"strings"
	"testing"

	"github.com/amarin/lexicon/textnorm"
)

func newTestAnalyzer() *Analyzer {
	return NewAnalyzer(fake, textnorm.PreReform, AnalyzerOptions{})
}

// show renders "form=lemma[flags],lemma[flags]; …" for comparisons.
func show(terms []Term) string {
	var parts []string

	for _, t := range terms {
		var ls []string
		for _, l := range t.Lemmas {
			ls = append(ls, fmt.Sprintf("%s[%s]", l.Text, l.Flags))
		}

		parts = append(parts, t.Form+"="+strings.Join(ls, ","))
	}

	return strings.Join(parts, "; ")
}

// TestAnalyzeIndexP1 is genodex P1 TestAnalyze (regression suite of ModeIndex).
// Only change: «покровскаго» carries the new reform flag (decision D12).
func TestAnalyzeIndexP1(t *testing.T) {
	cases := []struct {
		name, text string
		p          Profile
		want       string
	}{
		{"lemma", "Кота", profText, "кота=кот[]"},
		{"lemma normalized", "пса", profText, "пса=пес[]"},
		{"homonymy", "стали", profText, "стали=сталь[ambiguous],стать[ambiguous]"},
		{"stop word", "кот в селе", profText, "кот=кот[]; селе=село[]"},
		{"dotted abbreviation", "с. Кота", profText, "с.=село[ambiguous,abbrev],сын[ambiguous,abbrev]; кота=кот[]"},
		{"dot is not an abbreviation", "кот.", profText, "кот=кот[]"},
		{"hyphenated abbreviation", "кр-нин", profText, "кр-нин=крестьянин[abbrev]"},
		{"compound word", "Санктъ-Петербургъ", profText, "санкт-петербург=санкт-петербург[unknown]; санкт=санкт[unknown]; петербург=петербург[]"},
		{"plain abbreviation and word", "губ", profText, "губ=губа[ambiguous],губерния[ambiguous,abbrev]"},
		{"abbreviations never predict", "дер", profText, "дер=деревня[abbrev]"},
		{"user dictionaries never predict", "мороза", profName, "мороза=мороза[unknown]"},
		{"predicted lemma", "Дорожкиной", profText, "дорожкиной=дорожкина[predicted]"},
		{"exact beats predicted", "Кузнецова", profText, "кузнецова=кузнецов[]"},
		{"unknown word and number", "Дарожкн 1834", profText, "дарожкн=дарожкн[unknown]; 1834=1834[unknown]"},
		{"pre-reform ending", "Покровскаго", profText, "покровскаго=покровский[reform]"},
		{"name: common words are not parsed", "стали", profName, "стали=стали[unknown]"},
		{"name: a Name reading passes", "Ивана", profName, "ивана=иван[]"},
		{"place: abbreviations are parsed", "с.", profPlace, "с.=село[ambiguous,abbrev],сын[ambiguous,abbrev]"},
	}
	a := newTestAnalyzer()

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := show(a.Analyze(c.text, c.p, ModeIndex)); got != c.want {
				t.Errorf("Analyze(%q) = %q, want %q", c.text, got, c.want)
			}
		})
	}
}

// TestAnalyzeIndexV01: rules lexicon enforces beyond the P1 code (D2, D14).
// «Kот» (Latin K) is an intentional deviation from P1: P1 split the word at
// the script boundary and indexed «от»; lexicon replaces Latin homoglyphs
// inside Cyrillic words (owner decision 2026-09-24, Q1).
func TestAnalyzeIndexV01(t *testing.T) {
	cases := []struct{ name, text, want string }{
		{"Latin homoglyphs inside a Cyrillic word (P1: «от»)", "Kот", "кот=кот[]"},
		{"other Latin letters split the word", "Kотw", "от=от[unknown]"},
		{"one-letter dotted abbreviation is ambiguous", "ц. Кота", "ц.=церковь[ambiguous,abbrev]; кота=кот[]"},
		{"one letter without a dot is not an abbreviation", "с Кота", "кота=кот[]"},
		{"Latin words are not indexed", "Mississippi кот", "кот=кот[]"},
		{"punctuation is not indexed", "кот, кот!", "кот=кот[]; кот=кот[]"},
	}
	a := newTestAnalyzer()

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := show(a.Analyze(c.text, profText, ModeIndex)); got != c.want {
				t.Errorf("Analyze(%q) = %q, want %q", c.text, got, c.want)
			}
		})
	}
}

// TestAnalyzeKeepsReading: a lemma carries tag, kind and dictionary of the
// first reading that produced it (grammemes for NER).
func TestAnalyzeKeepsReading(t *testing.T) {
	terms := newTestAnalyzer().Analyze("Ивана", profText, ModeIndex)
	want := Lemma{Text: "иван", Tag: "NOUN,anim,masc,Name sing,gent", Kind: KindBase, Dict: "base.test"}

	if len(terms) != 1 || len(terms[0].Lemmas) != 1 || terms[0].Lemmas[0] != want {
		t.Fatalf("Analyze(Ивана) = %+v", terms)
	}
}

// TestAnalyzeKeepsToken: terms carry source tokens; compound parts carry their
// part tokens (D13).
func TestAnalyzeKeepsToken(t *testing.T) {
	a := newTestAnalyzer()

	text := "в селе Кота"
	terms := a.Analyze(text, profText, ModeIndex)

	last := terms[len(terms)-1].Token
	if text[last.Start:last.End] != "Кота" {
		t.Fatalf("token %+v", last)
	}

	text = "в Санктъ-Петербургъ"
	terms = a.Analyze(text, profText, ModeIndex)

	part := terms[len(terms)-1].Token
	if part.Raw != "Петербургъ" || text[part.Start:part.End] != part.Raw {
		t.Fatalf("part token %+v", part)
	}
}

func TestAnalyzerVersion(t *testing.T) {
	if got := newTestAnalyzer().Version(); got != "analyzer-1/prereform-2/fake-1" {
		t.Fatalf("Version = %q", got)
	}
}
