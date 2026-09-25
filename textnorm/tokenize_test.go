package textnorm

import (
	"slices"
	"strings"
	"testing"
	"unicode/utf16"
	"unicode/utf8"
)

// checkTokens verifies the offsets contract for toks of input: order, byte and
// rune offsets, Raw, grapheme-cluster boundaries as modelled by isExtend, and
// UTF-16 offsets. Shared by unit and fuzz tests.
func checkTokens(t *testing.T, input string, toks []Token) {
	t.Helper()

	prevEnd := 0

	for i, tk := range toks {
		if tk.Start < prevEnd || tk.End <= tk.Start || tk.End > len(input) {
			t.Fatalf("token %d %+v: bad span after %d in %q", i, tk, prevEnd, input)
		}

		if input[tk.Start:tk.End] != tk.Raw {
			t.Fatalf("token %d: input[%d:%d] = %q, Raw %q", i, tk.Start, tk.End, input[tk.Start:tk.End], tk.Raw)
		}

		if got := utf8.RuneCountInString(input[:tk.Start]); got != tk.RuneStart {
			t.Fatalf("token %d %q: RuneStart %d, want %d", i, tk.Raw, tk.RuneStart, got)
		}

		if got := tk.RuneStart + utf8.RuneCountInString(tk.Raw); got != tk.RuneEnd {
			t.Fatalf("token %d %q: RuneEnd %d, want %d", i, tk.Raw, tk.RuneEnd, got)
		}

		if first, _ := utf8.DecodeRuneInString(tk.Raw); isExtend(first) {
			t.Fatalf("token %d %q starts inside a grapheme cluster", i, tk.Raw)
		}

		if tk.End < len(input) {
			if next, _ := utf8.DecodeRuneInString(input[tk.End:]); isExtend(next) {
				t.Fatalf("token %d %q ends inside a grapheme cluster", i, tk.Raw)
			}
		}

		if (tk.Kind == TokenPunct || tk.Kind == TokenSymbol) && tk.Form != "" {
			t.Fatalf("token %d %q (%s): Form %q, want empty", i, tk.Raw, tk.Kind, tk.Form)
		}

		u := UTF16Offsets(input, tk.Start, tk.End)
		if want := len(utf16.Encode([]rune(input[:tk.Start]))); u[0] != want {
			t.Fatalf("token %d %q: UTF-16 start %d, want %d", i, tk.Raw, u[0], want)
		}

		prevEnd = tk.End
	}
}

// tokenLite is the comparable part of a token without offsets.
func tokenLite(tk Token) Token {
	return Token{Raw: tk.Raw, Form: tk.Form, Kind: tk.Kind, Script: tk.Script, Case: tk.Case, Dotted: tk.Dotted, SentenceEnd: tk.SentenceEnd}
}

// cyrForms — forms of Cyrillic words and numbers: what genodex P1 tokenized.
func cyrForms(toks []Token) []string {
	var out []string

	for _, tk := range toks {
		if tk.Kind == TokenNumber || (tk.Kind == TokenWord && tk.Script == ScriptCyrillic) {
			out = append(out, tk.Form)
		}
	}

	return out
}

// TestTokenizeP1: genodex P1 sentence — hyphenated word, dotted abbreviation,
// digits split from letters; punctuation is now kept.
func TestTokenizeP1(t *testing.T) {
	text := "Кр-нин с. Покровскаго, 1834г."
	got := Tokenize(PreReform, text)
	checkTokens(t, text, got)

	want := []Token{
		{Raw: "Кр-нин", Form: "кр-нин", Kind: TokenWord, Script: ScriptCyrillic, Case: CaseTitle},
		{Raw: "с", Form: "с", Kind: TokenWord, Script: ScriptCyrillic, Case: CaseLower, Dotted: true},
		{Raw: ".", Kind: TokenPunct, SentenceEnd: true},
		{Raw: "Покровскаго", Form: "покровскаго", Kind: TokenWord, Script: ScriptCyrillic, Case: CaseTitle},
		{Raw: ",", Kind: TokenPunct},
		{Raw: "1834", Form: "1834", Kind: TokenNumber},
		{Raw: "г", Form: "г", Kind: TokenWord, Script: ScriptCyrillic, Case: CaseLower, Dotted: true},
		{Raw: ".", Kind: TokenPunct, SentenceEnd: true},
	}
	if len(got) != len(want) {
		t.Fatalf("Tokenize(%q) = %+v, want %d tokens", text, got, len(want))
	}

	for i, w := range want {
		if g := tokenLite(got[i]); g != w {
			t.Errorf("token %d = %+v, want %+v", i, g, w)
		}
	}

	if got[3].Start != strings.Index(text, "Покровскаго") {
		t.Errorf("Start of «Покровскаго» = %d", got[3].Start)
	}
}

// TestTokenizeP1Edges: the genodex P1 edge cases, compared on Cyrillic words and
// numbers only (what P1 emitted).
func TestTokenizeP1Edges(t *testing.T) {
	cases := []struct {
		text  string
		forms []string
	}{
		{"Mississippi и", []string{"и"}},
		{"кот - пёс", []string{"кот", "пес"}},
		{"-кот-", []string{"кот"}},
		{"Бг҃ъ", []string{"бг"}},
		{"Iоаннъ Ѳеодоровъ", []string{"иоанн", "феодоров"}},
		{"Санктъ-Петербургъ", []string{"санкт-петербург"}},
		{"", nil},
	}
	for _, c := range cases {
		toks := Tokenize(PreReform, c.text)
		checkTokens(t, c.text, toks)

		if got := cyrForms(toks); !slices.Equal(got, c.forms) {
			t.Errorf("Tokenize(%q) forms = %v, want %v", c.text, got, c.forms)
		}
	}
}

// TestTokenizeKinds: words of any script, numbers, punctuation, symbols.
func TestTokenizeKinds(t *testing.T) {
	text := "Иванъ, сынъ — 5 лет! Mr. Smith № 7"
	got := Tokenize(PreReform, text)
	checkTokens(t, text, got)

	want := []Token{
		{Raw: "Иванъ", Form: "иван", Kind: TokenWord, Script: ScriptCyrillic, Case: CaseTitle},
		{Raw: ",", Kind: TokenPunct},
		{Raw: "сынъ", Form: "сын", Kind: TokenWord, Script: ScriptCyrillic, Case: CaseLower},
		{Raw: "—", Kind: TokenPunct},
		{Raw: "5", Form: "5", Kind: TokenNumber},
		{Raw: "лет", Form: "лет", Kind: TokenWord, Script: ScriptCyrillic, Case: CaseLower},
		{Raw: "!", Kind: TokenPunct, SentenceEnd: true},
		{Raw: "Mr", Form: "mr", Kind: TokenWord, Script: ScriptLatin, Case: CaseTitle, Dotted: true},
		{Raw: ".", Kind: TokenPunct, SentenceEnd: true},
		{Raw: "Smith", Form: "smith", Kind: TokenWord, Script: ScriptLatin, Case: CaseTitle},
		{Raw: "№", Kind: TokenSymbol},
		{Raw: "7", Form: "7", Kind: TokenNumber},
	}
	if len(got) != len(want) {
		t.Fatalf("Tokenize(%q) = %+v, want %d tokens", text, got, len(want))
	}

	for i, w := range want {
		if g := tokenLite(got[i]); g != w {
			t.Errorf("token %d = %+v, want %+v", i, g, w)
		}
	}
}

// TestTokenizeOffsets: byte and code-point offsets over multi-byte letters,
// combining marks and a 4-byte symbol.
func TestTokenizeOffsets(t *testing.T) {
	text := "Кузнецо\u0301в и \U0001F600 Iоаннъ"
	got := Tokenize(PreReform, text)
	checkTokens(t, text, got)

	raws := []string{"Кузнецо\u0301в", "и", "\U0001F600", "Iоаннъ"}
	if len(got) != len(raws) {
		t.Fatalf("Tokenize(%q) = %+v", text, got)
	}

	for i, raw := range raws {
		start := strings.Index(text, raw)
		if got[i].Raw != raw || got[i].Start != start || got[i].RuneStart != utf8.RuneCountInString(text[:start]) {
			t.Errorf("token %d = %+v, want Raw %q at byte %d", i, got[i], raw, start)
		}
	}

	if got[0].Form != "кузнецов" || got[2].Kind != TokenSymbol || got[3].Script != ScriptCyrillic {
		t.Errorf("tokens %+v", got)
	}
}

// TestTokenizeGraphemes: marks never start a token; ZWJ sequences, flags and
// keycaps stay whole.
func TestTokenizeGraphemes(t *testing.T) {
	cases := []struct {
		text string
		raws []string
	}{
		{"\u0301кот", []string{"кот"}},
		{"кот \u0301пёс", []string{"кот", "пёс"}},
		{"\U0001F468\u200d\U0001F469\u200d\U0001F467 и", []string{"\U0001F468\u200d\U0001F469\u200d\U0001F467", "и"}},
		{"\U0001F1F7\U0001F1FA!", []string{"\U0001F1F7\U0001F1FA", "!"}},
		{"5\u20e3", []string{"5\u20e3"}},
		{".\u0301", []string{".\u0301"}},
		{"кот\u200bпёс", []string{"кот", "пёс"}}, // zero-width space separates words
	}
	for _, c := range cases {
		toks := Tokenize(PreReform, c.text)
		checkTokens(t, c.text, toks)

		var raws []string
		for _, tk := range toks {
			raws = append(raws, tk.Raw)
		}

		if !slices.Equal(raws, c.raws) {
			t.Errorf("Tokenize(%q) = %q, want %q", c.text, raws, c.raws)
		}
	}

	if got := Tokenize(PreReform, "5\u20e3"); got[0].Form != "5" {
		t.Errorf("keycap form %q, want 5", got[0].Form)
	}
}

// TestTokenizeCase: case classes of words; hyphen parts may each be titled.
func TestTokenizeCase(t *testing.T) {
	cases := map[string]Case{
		"кот": CaseLower, "Кот": CaseTitle, "КОТ": CaseUpper, "И": CaseTitle, "СПб": CaseMixed,
		"Санкт-Петербург": CaseTitle, "Кр-нин": CaseTitle, "санкт-Петербург": CaseMixed, "12": CaseNone,
	}
	for text, want := range cases {
		toks := Tokenize(Modern, text)
		if len(toks) != 1 || toks[0].Case != want {
			t.Errorf("Tokenize(%q) = %+v, want one token with case %s", text, toks, want)
		}
	}
}

// TestTokenizeSentenceEnd: sentence punctuation ends a sentence; a single line
// break does not (hard-wrapped text), a blank line does (spec Q5).
func TestTokenizeSentenceEnd(t *testing.T) {
	cases := []struct {
		text string
		want []bool
	}{
		{"кот\nпёс. Кот", []bool{false, false, true, false}}, // кот, пёс, ., Кот
		{"Нижний\r\nНовгород", []bool{false, false}},         // CRLF is one break
		{"кот\n\nпёс", []bool{true, false}},                  // blank line
		{"кот\n  \r\n\tпёс", []bool{true, false}},            // blank line with spaces
		{"кот\u2029пёс", []bool{true, false}},                // paragraph separator
	}
	for _, c := range cases {
		got := Tokenize(Modern, c.text)
		checkTokens(t, c.text, got)

		if len(got) != len(c.want) {
			t.Fatalf("Tokenize(%q) = %+v", c.text, got)
		}

		for i, w := range c.want {
			if got[i].SentenceEnd != w {
				t.Errorf("%q token %d %q: SentenceEnd %v, want %v", c.text, i, got[i].Raw, got[i].SentenceEnd, w)
			}
		}
	}
}

// TestTokenizeScriptSplit: a hyphen part with Cyrillic letters and only Latin
// homoglyphs besides is one Cyrillic word, homoglyphs replaced in Form and Raw
// untouched; other mixed parts split at Cyrillic/non-Cyrillic boundaries;
// parts without Cyrillic letters stay as they are (decision D2). Latin i/I is
// a homoglyph only in PreReform.
func TestTokenizeScriptSplit(t *testing.T) {
	cases := []struct {
		r    Rules
		text string
		want []string // Raw/script/Form
	}{
		{Modern, "Kот", []string{"Kот/cyrillic/кот"}},
		{PreReform, "Kот", []string{"Kот/cyrillic/кот"}},
		{Modern, "HOBЫЙ", []string{"HOBЫЙ/cyrillic/новый"}}, // upper-case H, O, B fold too
		{Modern, "Kотw", []string{"K/latin/k", "от/cyrillic/от", "w/latin/w"}},
		{PreReform, "Iоаннъ", []string{"Iоаннъ/cyrillic/иоанн"}},
		{Modern, "Iоаннъ", []string{"I/latin/i", "оаннъ/cyrillic/оаннъ"}},
		{PreReform, "XIX-го", []string{"XIX/latin/xix", "-/none/", "го/cyrillic/го"}},
		{Modern, "Kот-Cмирнов", []string{"Kот-Cмирнов/cyrillic/кот-смирнов"}},
		{Modern, "Kот-Smith", []string{"Kот/cyrillic/кот", "-/none/", "Smith/latin/smith"}},
		{Modern, "MOCKBA", []string{"MOCKBA/latin/mockba"}}, // no Cyrillic letter
		{Modern, "αβγ", []string{"αβγ/mixed/αβγ"}},
	}
	for _, c := range cases {
		toks := Tokenize(c.r, c.text)
		checkTokens(t, c.text, toks)

		var got []string
		for _, tk := range toks {
			got = append(got, tk.Raw+"/"+tk.Script.String()+"/"+tk.Form)
		}

		if !slices.Equal(got, c.want) {
			t.Errorf("Tokenize(%s, %q) = %v, want %v", c.r.Name, c.text, got, c.want)
		}
	}
}
