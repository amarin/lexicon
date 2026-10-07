package dates

import (
	"strings"
	"testing"
)

func TestMonth(t *testing.T) {
	cases := map[string]int{
		"январь": 1, "января": 1, "генваря": 1, "генварь": 1, "янв.": 1, "Янв": 1,
		"февраля": 2, "ѳевраля": 2, "фев.": 2, "февр.": 2,
		"март": 3, "марта": 3, "мар.": 3,
		"апреля": 4, "апр.": 4,
		"май": 5, "мая": 5, "МАЯ": 5,
		"іюня": 6, "июня": 6, "июнь": 6, "июн.": 6,
		"іюля": 7, "июля": 7, "июл.": 7,
		"августа": 8, "авг.": 8,
		"сентября": 9, "сент.": 9, "сен.": 9, "сентябрь": 9,
		"октября": 10, "окт.": 10,
		"ноября": 11, "нояб.": 11, "ноя": 11,
		"декабря": 12, "дек.": 12, "декабрь": 12, " декабря ": 12,
	}
	for in, want := range cases {
		got, ok := Month(in)
		if !ok || got != want {
			t.Errorf("Month(%q) = %d, %v; want %d", in, got, ok, want)
		}
	}

	for _, in := range []string{"", ".", "мартышка", "январьь", "13", "lorem", "ма", "янв.."} {
		if n, ok := Month(in); ok {
			t.Errorf("Month(%q) = %d, true; want false", in, n)
		}
	}
}

func TestParse(t *testing.T) {
	cases := []struct {
		name string
		in   string
		ctx  int
		want Date
		ok   bool
	}{
		{"abbrev with context", "21 янв.", 1884, Date{1884, 1, 21, true}, true},
		{"genitive pre-reform year", "21 генваря 1883", 1884, Date{1883, 1, 21, false}, true},
		{"text year wins", "21 января 1883 года", 1900, Date{1883, 1, 21, false}, true},
		{"г. suffix", "1 іюня 1883 г.", 0, Date{1883, 6, 1, false}, true},
		{"г suffix no dot", "1 июня 1883г", 0, Date{1883, 6, 1, false}, true},
		{"month only suffix", "5 мая г.", 1890, Date{1890, 5, 5, true}, true},
		{"sept abbrev", "3 сент.", 1885, Date{1885, 9, 3, true}, true},
		{"abbrev dot glued year", "3 сент. 1885", 0, Date{1885, 9, 3, false}, true},
		{"dots around month", "3.сент.1885", 0, Date{1885, 9, 3, false}, true},
		{"ordinal day", "21-го января 1883", 0, Date{1883, 1, 21, false}, true},
		{"yat", "14 ѳевраля 1850", 0, Date{1850, 2, 14, false}, true},
		{"upper case", "21 ЯНВАРЯ", 1884, Date{1884, 1, 21, true}, true},
		{"extra spaces", "  21   января  1883 ", 0, Date{1883, 1, 21, false}, true},
		{"numeric no year", "21.01", 1884, Date{1884, 1, 21, true}, true},
		{"numeric year", "21.01.1884", 1900, Date{1884, 1, 21, false}, true},
		{"numeric slash", "5/2/1884", 0, Date{1884, 2, 5, false}, true},
		{"numeric year suffix", "21.01.1884 г.", 0, Date{1884, 1, 21, false}, true},
		{"leap day", "29 февраля 1884", 0, Date{1884, 2, 29, false}, true},
		{"leap day ctx", "29 февр.", 1884, Date{1884, 2, 29, true}, true},
		{"no leap 1900", "29 февраля 1900", 0, Date{}, false},
		{"no leap ctx", "29.02", 1885, Date{}, false},
		{"leap 2000", "29.02.2000", 0, Date{2000, 2, 29, false}, true},

		{"no year no ctx", "21 янв.", 0, Date{}, false},
		{"numeric no year no ctx", "21.01", 0, Date{}, false},
		{"31 feb", "31 февраля 1884", 0, Date{}, false},
		{"31 apr", "31 апреля", 1884, Date{}, false},
		{"month 13", "1.13", 1884, Date{}, false},
		{"0.13", "0.13", 1884, Date{}, false},
		{"day 0", "0 января 1884", 0, Date{}, false},
		{"day 32", "32 января 1884", 0, Date{}, false},
		{"month 0", "5.00.1884", 0, Date{}, false},
		{"two digit year", "21.01.84", 0, Date{}, false},
		{"unknown month", "21 мартышка 1884", 0, Date{}, false},
		{"garbage", "что-то", 1884, Date{}, false},
		{"empty", "", 1884, Date{}, false},
		{"month only", "января", 1884, Date{}, false},
		{"trailing junk", "21 января 1883 года рожд.", 0, Date{}, false},
		{"negative ctx", "21 января", -5, Date{}, false},
		{"ctx out of range", "21 января", 12345, Date{}, false},
		{"text year out of range", "21 января 0999", 0, Date{}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := Parse(c.in, c.ctx)
			if ok != c.ok || (ok && got != c.want) {
				t.Errorf("Parse(%q, %d) = %+v, %v; want %+v, %v", c.in, c.ctx, got, ok, c.want, c.ok)
			}
		})
	}
}

func TestParseNeverPanics(t *testing.T) {
	inputs := []string{
		"", " ", "\x00", "\xff\xfe", strings.Repeat("9", 10000), strings.Repeat("я", 10000),
		"99999999999999999999.1.1", "1.99999999999999999999", "٣ января", "ⅩⅩ января",
		"21 ́января", "21 ян​варя 1883", "i̇юня", "1 июня 1883 г.\n", "🙂", "-1.1.1884",
	}
	for _, in := range inputs {
		for _, ctx := range []int{0, 1884, -1, 1 << 40} {
			Parse(in, ctx)
			Month(in)
		}
	}
}

func FuzzParse(f *testing.F) {
	for _, s := range []string{"21 янв.", "21.01.1884", "29 февраля 1884", "іюня", "", "\xff"} {
		f.Add(s, 1884)
	}

	f.Fuzz(func(t *testing.T, s string, ctx int) {
		d, ok := Parse(s, ctx)
		if !ok {
			return
		}

		if d.Month < 1 || d.Month > 12 || d.Day < 1 || d.Day > 31 || d.Year < minYear || d.Year > maxYear {
			t.Fatalf("Parse(%q, %d) = %+v out of range", s, ctx, d)
		}

		if d.YearFromContext && d.Year != ctx {
			t.Fatalf("Parse(%q, %d) = %+v: context flag without context year", s, ctx, d)
		}
	})
}
