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
		{
			"full words of an abbreviation dictionary are not abbreviated", "Сын деревня город с. г.",
			"сын=сын[]; деревня=деревня[]; город=город[]; " +
				"с.=село[ambiguous,abbrev],сын[ambiguous,abbrev]; .; г.=город[ambiguous,abbrev],год[ambiguous,abbrev]; .",
		},
		{"full word before a sentence dot", "город.", "город=город[]; ."},
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

// TestAnalyzeFullAbbrevRegistry: a registry whose base lacks the full words
// (genodex without a base file) reads them from the abbreviation dictionary
// only; they are still not abbreviated.
func TestAnalyzeFullAbbrevRegistry(t *testing.T) {
	a := NewAnalyzer(openTest(t, t.TempDir(), nil), textnorm.PreReform, AnalyzerOptions{})

	const want = "сын=сын[]; деревня=деревня[]; с.=село[ambiguous,abbrev],сын[ambiguous,abbrev]; .; дер.=деревня[abbrev]; ."
	if got := showFull(a.Analyze("сын деревня с. дер.", profText, ModeFull)); got != want {
		t.Errorf("Analyze = %q, want %q", got, want)
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
