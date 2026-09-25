package textnorm

import (
	"testing"
	"unicode/utf8"
)

// TestSplitHyphen: parts keep offsets into the input; Dotted goes to the last
// part; tokens without an inner hyphen have no parts.
func TestSplitHyphen(t *testing.T) {
	text := "в Санктъ‐Петербургъ."
	toks := Tokenize(PreReform, text)

	whole := toks[1]
	if whole.Raw != "Санктъ‐Петербургъ" || whole.Form != "санкт-петербург" || !whole.Dotted {
		t.Fatalf("whole token %+v", whole)
	}

	parts := SplitHyphen(whole)
	want := []struct {
		raw, form string
		dotted    bool
	}{{"Санктъ", "санкт", false}, {"Петербургъ", "петербург", true}}

	if len(parts) != len(want) {
		t.Fatalf("SplitHyphen = %+v", parts)
	}

	for i, w := range want {
		p := parts[i]
		if p.Raw != w.raw || p.Form != w.form || p.Dotted != w.dotted || p.Kind != TokenWord ||
			p.Script != ScriptCyrillic || p.Case != CaseTitle {
			t.Errorf("part %d = %+v, want %+v", i, p, w)
		}

		if text[p.Start:p.End] != p.Raw {
			t.Errorf("part %d: text[%d:%d] = %q", i, p.Start, p.End, text[p.Start:p.End])
		}

		if p.RuneStart != utf8.RuneCountInString(text[:p.Start]) || p.RuneEnd != p.RuneStart+utf8.RuneCountInString(p.Raw) {
			t.Errorf("part %d rune offsets %d..%d", i, p.RuneStart, p.RuneEnd)
		}
	}

	if SplitHyphen(toks[0]) != nil {
		t.Error("SplitHyphen(«в») != nil")
	}

	if SplitHyphen(toks[2]) != nil {
		t.Error("SplitHyphen(«.») != nil")
	}
}
