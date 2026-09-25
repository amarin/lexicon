package textnorm

import "testing"

// TestOrthographyPreReform: case, pre-reform letters, combining marks and format
// characters; й is kept; the final hard sign stays (NormalizeWord drops it).
func TestOrthographyPreReform(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Ивановъ", "ивановъ"},
		{"Покровскій", "покровский"},
		{"Ѳеодоръ", "феодоръ"},
		{"Евдокія", "евдокия"},
		{"Iоаннъ", "иоаннъ"}, // Latin I inside a Cyrillic word
		{"Бѣлёвъ", "белевъ"},
		{"Бг҃ъ", "бгъ"}, // titlo
		{"Андрей", "андрей"},
		{"Андрей", "андрей"},      // й typed decomposed
		{"Кузнецо́в", "кузнецов"}, // stress mark
		{"Ѡ ѯ ѱ ꙋ ѧ ѵ", "о кс пс у я и"},
		{"кр‐нин", "кр-нин"},           // U+2010 hyphen
		{"Mississippi", "mississippi"}, // no Cyrillic letter: Latin i stays Latin
		{"Kот", "кот"},                 // Latin K: homoglyph inside a Cyrillic word
		{"Kотw", "kотw"},               // w is not a homoglyph: nothing replaced
		{"Пет­ров", "петров"},          // soft hyphen (format character) dropped
	}
	for _, c := range cases {
		if got := Orthography(PreReform, c.in); got != c.want {
			t.Errorf("Orthography(PreReform, %q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestOrthographyModern: only case, marks, ё, the U+2010 hyphen and Latin
// homoglyphs inside Cyrillic word parts change.
func TestOrthographyModern(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Ивановъ", "ивановъ"},
		{"Ёлкин", "елкин"},
		{"Покровскій", "покровскій"},
		{"Ѳеодоръ", "ѳеодоръ"},
		{"Iоаннъ", "iоаннъ"}, // Latin I is a homoglyph only in PreReform
		{"Андрей", "андрей"},
		{"Кузнецо́в", "кузнецов"},
		{"кр‐нин", "кр-нин"},
		{"Kот", "кот"},
		{"HOBЫЙ", "новый"},   // upper-case H, O, B fold before lower-casing
		{"XIX-го", "xix-го"}, // a part without Cyrillic letters is untouched
		{"MOCKBA", "mockba"}, // no Cyrillic letter at all: stays Latin
	}
	for _, c := range cases {
		if got := Orthography(Modern, c.in); got != c.want {
			t.Errorf("Orthography(Modern, %q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestNormalizeWord: PreReform drops the final hard sign of every hyphen part;
// Modern keeps it.
func TestNormalizeWord(t *testing.T) {
	cases := []struct {
		r        Rules
		in, want string
	}{
		{PreReform, "Ивановъ", "иванов"},
		{PreReform, "Санктъ-Петербургъ", "санкт-петербург"},
		{PreReform, "Бг҃ъ", "бг"},
		{PreReform, "объявленіе", "объявление"}, // inner ъ stays
		{Modern, "Ивановъ", "ивановъ"},
	}
	for _, c := range cases {
		if got := NormalizeWord(c.r, c.in); got != c.want {
			t.Errorf("NormalizeWord(%s, %q) = %q, want %q", c.r.Name, c.in, got, c.want)
		}
	}
}

// TestRuleSets: names and versions feed Analyzer.Version.
func TestRuleSets(t *testing.T) {
	if Modern.Name != "modern" || PreReform.Name != "prereform" {
		t.Fatalf("names %q, %q", Modern.Name, PreReform.Name)
	}

	if Modern.Version == "" || PreReform.Version == "" {
		t.Fatal("empty rule set version")
	}
}
