package dates_test

import (
	"fmt"

	"github.com/amarin/lexicon/dates"
)

func ExampleParse() {
	// The year comes from the text when it is there.
	d, ok := dates.Parse("21 генваря 1883 года", 1884)
	fmt.Println(d.Year, d.Month, d.Day, d.YearFromContext, ok)

	// Otherwise the context year (for example the year of the record column) is used.
	d, ok = dates.Parse("3 сент.", 1885)
	fmt.Println(d.Year, d.Month, d.Day, d.YearFromContext, ok)

	// No year and no context, or an impossible date: not a date.
	_, ok = dates.Parse("3 сент.", 0)
	fmt.Println(ok)
	_, ok = dates.Parse("31 февраля 1885", 0)
	fmt.Println(ok)
	// Output:
	// 1883 1 21 false true
	// 1885 9 3 true true
	// false
	// false
}

func ExampleMonth() {
	for _, w := range []string{"іюня", "генваря", "сент.", "май", "мартышка"} {
		n, ok := dates.Month(w)
		fmt.Println(w, n, ok)
	}
	// Output:
	// іюня 6 true
	// генваря 1 true
	// сент. 9 true
	// май 5 true
	// мартышка 0 false
}
