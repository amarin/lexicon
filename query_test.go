package lexicon

import "testing"

// TestParseQuery (genodex P1): the last word without anything after it is
// Partial with lemmas; numbers and dotted words are complete; a trailing
// service word is Partial without lemmas.
func TestParseQuery(t *testing.T) {
	cases := []struct {
		q, terms, partial string
	}{
		{"Кота Покро", "кота=кот[]", "покро=покро[unknown]"},
		{"Кота Покро ", "кота=кот[]; покро=покро[unknown]", ""},
		{"в Покро", "", "покро=покро[unknown]"},
		{"Кота", "", "кота=кот[]"},
		{"кот в", "кот=кот[]", "в="},
		{"с.", "с.=село[ambiguous,abbrev],сын[ambiguous,abbrev]", ""},
		{"1834", "1834=1834[unknown]", ""},
		{"Кота!", "кота=кот[]", ""},
		{"", "", ""},
	}
	a := newTestAnalyzer()

	for _, c := range cases {
		got := a.ParseQuery(c.q, profText)

		partial := ""
		if got.Partial != nil {
			partial = show([]Term{*got.Partial})
		}

		if show(got.Terms) != c.terms || partial != c.partial {
			t.Errorf("ParseQuery(%q) = %q / %q, want %q / %q", c.q, show(got.Terms), partial, c.terms, c.partial)
		}
	}
}
