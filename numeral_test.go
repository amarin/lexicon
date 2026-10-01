package lexicon

import (
	"math"
	"strings"
	"testing"
)

// TestNumeralGrammemes: the Russian count agreement for a nominative phrase,
// by the last two digits (pymorphy2 make_agree_with_number).
func TestNumeralGrammemes(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{0, "gent plur"},
		{1, "nomn sing"},
		{2, "gent sing"}, {3, "gent sing"}, {4, "gent sing"},
		{5, "gent plur"}, {9, "gent plur"}, {10, "gent plur"},
		{11, "gent plur"}, {12, "gent plur"}, {13, "gent plur"}, {14, "gent plur"},
		{15, "gent plur"}, {20, "gent plur"},
		{21, "nomn sing"}, {22, "gent sing"}, {25, "gent plur"},
		{100, "gent plur"}, {101, "nomn sing"}, {102, "gent sing"},
		{111, "gent plur"}, {112, "gent plur"}, {114, "gent plur"}, {121, "nomn sing"},
		{-1, "nomn sing"}, {-3, "gent sing"}, {-11, "gent plur"},
		{math.MinInt, "gent plur"},
	}

	for _, c := range cases {
		if got := strings.Join(NumeralGrammemes(c.n), " "); got != c.want {
			t.Errorf("NumeralGrammemes(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}

// TestNumeralGrammemesFreshSlice: the caller may append to the result.
func TestNumeralGrammemesFreshSlice(t *testing.T) {
	a := NumeralGrammemes(1)
	a[0] = "x"

	if got := NumeralGrammemes(1)[0]; got != "nomn" {
		t.Fatalf("shared slice: %q", got)
	}
}
