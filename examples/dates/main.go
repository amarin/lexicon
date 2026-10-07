// Command dates reads dates from the columns of a pre-reform record: month
// words in the genitive and abbreviated, pre-reform spelling, numeric forms,
// and a context year (the year of the register page) for dates written
// without one. Parse reports false for anything that is not a real date, so
// the host keeps the raw text. Dates are taken as written: there is no
// Julian/Gregorian conversion.
//
// Run: go run ./examples/dates
package main

import (
	"fmt"

	"github.com/amarin/lexicon/dates"
)

func main() {
	// ctxYear is the year of the register page; the year in the text wins.
	const pageYear = 1884

	for _, text := range []string{
		"21 генваря 1883 года", // the year from the text
		"1 іюня 1883 г.",       // pre-reform spelling and a suffix
		"3 сент.",              // no year: the page year
		"21-го января",         // an ordinal day
		"21.01",                // numeric, no year
		"21.01.1884",           // numeric with a year
		"31 февраля",           // impossible
		"21.01.84",             // a two-digit year is ambiguous
		"21 мартышка",          // not a month
	} {
		d, ok := dates.Parse(text, pageYear)
		if !ok {
			fmt.Printf("%-22s not a date\n", text)

			continue
		}

		fmt.Printf("%-22s %04d-%02d-%02d fromContext=%v\n", text, d.Year, d.Month, d.Day, d.YearFromContext)
	}

	// Without a context year a date with no year fails.
	_, ok := dates.Parse("3 сент.", 0)
	fmt.Println("no context:", ok)

	// A lone month word, for example a column heading.
	n, ok := dates.Month("ѳевраля")
	fmt.Println("ѳевраля:", n, ok)

	// Output:
	// 21 генваря 1883 года   1883-01-21 fromContext=false
	// 1 іюня 1883 г.         1883-06-01 fromContext=false
	// 3 сент.                1884-09-03 fromContext=true
	// 21-го января           1884-01-21 fromContext=true
	// 21.01                  1884-01-21 fromContext=true
	// 21.01.1884             1884-01-21 fromContext=false
	// 31 февраля             not a date
	// 21.01.84               not a date
	// 21 мартышка            not a date
	// no context: false
	// ѳевраля: 2 true
}
