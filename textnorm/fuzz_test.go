package textnorm

import (
	"testing"
	"unicode/utf16"
)

// FuzzTokenize: the offsets contract holds for any input and both rule sets;
// hyphen parts keep it too and cover the word without gaps except hyphens.
func FuzzTokenize(f *testing.F) {
	for _, s := range []string{
		"Кр-нин с. Покровскаго, 1834г.", "Кузнецо\u0301в и \U0001F600 Iоаннъ", "\U0001F468\u200d\U0001F469\u200d\U0001F467 \U0001F1F7\U0001F1FA 5\u20e3",
		"\u0301кот \u0301пёс", "Санктъ\u2010Петербургъ.", "XIX-го Kот Kотw HOBЫЙ αβγ", "a\xffб\xe2\x82", "кот\u200bпёс\n\n",
	} {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		for _, r := range []Rules{Modern, PreReform} {
			toks := Tokenize(r, input)
			checkTokens(t, input, toks)

			for _, tk := range toks {
				parts := SplitHyphen(tk)
				if parts == nil {
					continue
				}

				checkTokens(t, input, parts)

				if parts[0].Start != tk.Start || parts[len(parts)-1].End != tk.End {
					t.Fatalf("parts %+v do not cover %+v", parts, tk)
				}
			}
		}
	})
}

// FuzzUTF16Offsets: UTF16Offsets agrees with utf16.Encode at code-point
// boundaries.
func FuzzUTF16Offsets(f *testing.F) {
	for _, s := range []string{"a\U0001F600б", "Кузнецо\u0301в", "\xff\xfe", ""} {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		var offs []int
		for off := range input {
			offs = append(offs, off)
		}

		offs = append(offs, len(input))
		got := UTF16Offsets(input, offs...)

		for i, off := range offs {
			if want := len(utf16.Encode([]rune(input[:off]))); got[i] != want {
				t.Fatalf("UTF16Offsets(%q, %d) = %d, want %d", input, off, got[i], want)
			}
		}
	})
}
