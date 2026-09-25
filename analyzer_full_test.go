package lexicon

import (
	"strings"
	"testing"

	"github.com/amarin/lexicon/textnorm"
)

// showFull renders terms like show; terms without lemmas show their raw text.
func showFull(terms []Term) string {
	var parts []string

	for _, t := range terms {
		if len(t.Lemmas) == 0 {
			parts = append(parts, t.Token.Raw)

			continue
		}

		parts = append(parts, show([]Term{t}))
	}

	return strings.Join(parts, "; ")
}

func TestAnalyzeFull(t *testing.T) {
	cases := []struct{ name, text, want string }{
		{
			"all tokens kept", "Кот в селе, 1834 г. Mississippi!",
			"кот=кот[]; в=в[stop]; селе=село[]; ,; 1834=1834[unknown]; " +
				"г.=город[ambiguous,abbrev],год[ambiguous,abbrev]; .; mississippi=mississippi[unknown]; !",
		},
		{"one letter without a dot is a stop word", "с Кота", "с=с[stop]; кота=кот[]"},
		{"compound stays whole", "Санктъ-Петербургъ", "санкт-петербург=санкт-петербург[unknown]"},
		{"hyphenated abbreviation", "кр-нин", "кр-нин=крестьянин[abbrev]"},
		{"pre-reform ending", "Покровскаго", "покровскаго=покровский[reform]"},
	}
	a := newTestAnalyzer()

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := showFull(a.Analyze(c.text, profText, ModeFull)); got != c.want {
				t.Errorf("Analyze(%q, ModeFull) = %q, want %q", c.text, got, c.want)
			}
		})
	}
}

// TestAnalyzeFullOneTermPerToken: term i is token i of Tokenize.
func TestAnalyzeFullOneTermPerToken(t *testing.T) {
	a := newTestAnalyzer()

	for _, text := range []string{"Кр-нин с. Покровскаго, 1834г.", "у X сын Y — 5 лет!", "", "👨\u200d👩 и"} {
		toks := textnorm.Tokenize(textnorm.PreReform, text)

		terms := a.Analyze(text, profName, ModeFull)
		if len(terms) != len(toks) {
			t.Fatalf("Analyze(%q): %d terms for %d tokens", text, len(terms), len(toks))
		}

		for i := range toks {
			if terms[i].Token != toks[i] {
				t.Fatalf("Analyze(%q): term %d token %+v, want %+v", text, i, terms[i].Token, toks[i])
			}
		}
	}
}

// TestAnalyzeFullStopTag: stop words keep their tag for patterns.
func TestAnalyzeFullStopTag(t *testing.T) {
	terms := newTestAnalyzer().Analyze("в", profText, ModeFull)
	if len(terms) != 1 || terms[0].Lemmas[0].Tag != "PREP" || terms[0].Lemmas[0].Flags != FlagStop {
		t.Fatalf("Analyze(в) = %+v", terms)
	}
}
