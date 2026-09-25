package textnorm

import (
	"slices"
	"testing"
)

// TestReformVariants: pre-reform adjective endings give modern variants; other
// words give none. -ой/-ей and -ого/-его depend on the last stem letter.
func TestReformVariants(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"покровскаго", []string{"покровского"}},
		{"синяго", []string{"синего"}},
		{"хорошаго", []string{"хорошого", "хорошего"}}, // sibilant: both tried
		{"большаго", []string{"большого", "большего"}},
		{"новыя", []string{"новые", "новой"}},
		{"покровския", []string{"покровские", "покровской"}}, // к/г/х → -ой
		{"синия", []string{"синие", "синей"}},
		{"кот", nil},
	}
	for _, c := range cases {
		if got := ReformVariants(c.in); !slices.Equal(got, c.want) {
			t.Errorf("ReformVariants(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
