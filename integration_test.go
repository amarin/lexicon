//go:build integration

package lexicon

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/amarin/lexicon/textnorm"
)

// realBase reads the base dictionary from $LEXICON_BASE_DAT (see
// `lexicon dicts fetch`).
func realBase(t testing.TB) []byte {
	t.Helper()

	path := os.Getenv("LEXICON_BASE_DAT")
	if path == "" {
		t.Skip("LEXICON_BASE_DAT is not set (path to base.opencorpora.dat)")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return data
}

// scribeAbbrev is a subset of the genodex scribe abbreviations.
var scribeAbbrev = BuiltinDict{
	Name: "abbrev.scribes", Kind: KindAbbrev, Format: FormatTSV,
	Data: []byte("село\tс.\tNOUN\nсын\tс.\tNOUN\nгород\tг.\tNOUN\nгод\tг.\tNOUN\nкрестьянин\tкр-нин\tNOUN\n"),
}

func openReal(t testing.TB) (*Registry, *Analyzer) {
	t.Helper()

	r, err := Open(t.Context(), Options{Base: realBase(t), Builtin: []BuiltinDict{scribeAbbrev}})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = r.Close() })

	return r, NewAnalyzer(r, textnorm.PreReform, AnalyzerOptions{})
}

// lemmaWith reports whether terms of form have lemma text with all flags.
func lemmaWith(terms []Term, form, text string, flags Flag) bool {
	for _, t := range terms {
		if t.Form != form {
			continue
		}

		if slices.ContainsFunc(t.Lemmas, func(l Lemma) bool { return l.Text == text && l.Flags&flags == flags }) {
			return true
		}
	}

	return false
}

// TestRealBaseP1Sentence: the genodex P1 live check on the real dictionary.
func TestRealBaseP1Sentence(t *testing.T) {
	_, a := openReal(t)

	terms := a.Analyze("Кр-нин с. Покровскаго Ивановъ, 1834 г.", Profile{Name: "text"}, ModeIndex)

	checks := []struct {
		form, lemma string
		flags       Flag
	}{
		{"кр-нин", "крестьянин", FlagAbbrev},
		{"с.", "село", FlagAbbrev | FlagAmbiguous},
		{"с.", "сын", FlagAbbrev | FlagAmbiguous},
		{"покровскаго", "покровский", FlagReform},
		{"иванов", "иванов", 0},
		{"1834", "1834", FlagUnknown},
		{"г.", "город", FlagAbbrev | FlagAmbiguous},
	}
	for _, c := range checks {
		if !lemmaWith(terms, c.form, c.lemma, c.flags) {
			t.Errorf("no %s → %s[%s] in %+v", c.form, c.lemma, c.flags, terms)
		}
	}
}

// TestRealBasePrediction: an out-of-dictionary surname is predicted by the
// base dictionary only.
func TestRealBasePrediction(t *testing.T) {
	r, a := openReal(t)

	// «Бырдыкиной» is a made-up surname: absent from OpenCorpora, predictable by its ending.
	rs := r.Parse("бырдыкиной", nil)
	if len(rs) == 0 {
		t.Fatal("no prediction for «бырдыкиной»")
	}

	for _, rd := range rs {
		if !rd.Predicted || rd.Kind != KindBase {
			t.Fatalf("reading %+v", rd)
		}
	}

	terms := a.Analyze("Бырдыкиной", Profile{Name: "text"}, ModeFull)
	if len(terms) != 1 || !strings.Contains(terms[0].Lemmas[0].Flags.String(), "predicted") {
		t.Fatalf("terms %+v", terms)
	}
}

func BenchmarkRealBaseAnalyzeFull(b *testing.B) {
	_, a := openReal(b)
	a.Analyze(analyzerBenchText, Profile{Name: "text"}, ModeFull)

	b.SetBytes(int64(len(analyzerBenchText)))
	b.ReportAllocs()

	for b.Loop() {
		a.Analyze(analyzerBenchText, Profile{Name: "text"}, ModeFull)
	}
}
