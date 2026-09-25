package lexicon

import (
	"testing"

	"github.com/amarin/lexicon/textnorm"
)

// FuzzAnalyze: ModeFull is 1:1 with tokens; ModeIndex terms have valid
// offsets, a form and lemmas; a Partial query term ends the query.
func FuzzAnalyze(f *testing.F) {
	for _, s := range []string{
		"Кр-нин с. Покровскаго, 1834г.", "кот в селе", "Санктъ-Петербургъ", "с. Кота ц.", "Mississippi и 👨\u200d👩", "a\xffб",
	} {
		f.Add(s)
	}

	a := newTestAnalyzer()

	f.Fuzz(func(t *testing.T, text string) {
		toks := textnorm.Tokenize(textnorm.PreReform, text)

		full := a.Analyze(text, profText, ModeFull)
		if len(full) != len(toks) {
			t.Fatalf("%d full terms for %d tokens", len(full), len(toks))
		}

		for i := range full {
			if full[i].Token != toks[i] {
				t.Fatalf("term %d token %+v, want %+v", i, full[i].Token, toks[i])
			}
		}

		for _, term := range a.Analyze(text, profName, ModeIndex) {
			tk := term.Token
			if text[tk.Start:tk.End] != tk.Raw || term.Form == "" || len(term.Lemmas) == 0 {
				t.Fatalf("index term %+v", term)
			}
		}

		if q := a.ParseQuery(text, profText); q.Partial != nil && q.Partial.Token.End != len(text) {
			t.Fatalf("Partial %+v does not end the query", q.Partial)
		}
	})
}
