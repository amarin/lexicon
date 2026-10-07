package dates

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/amarin/lexicon/textnorm"
)

// Date is a calendar date. Year is 1000..9999, Month 1..12, Day 1..31, and the
// combination is a real date (29 February only in a leap year).
type Date struct {
	Year, Month, Day int
	// YearFromContext is true when the text had no year and Year is the
	// context year given to Parse.
	YearFromContext bool
}

const (
	minYear = 1000
	maxYear = 9999
)

// yearPart is an optional four-digit year with an optional "г.", "г" or "года"
// after it; the suffix is not allowed without a year.
const yearPart = `(?:\s*(\d{4})(?:\s*(?:года|г)\.?)?)?`

var (
	// numericRe: "21.01", "21.01.1884", "21/01/1884", "21-01-1884".
	numericRe = regexp.MustCompile(`^(\d{1,2})[./-](\d{1,2})(?:[./-](\d{4})(?:\s*(?:года|г)\.?)?)?$`)
	// wordRe: "21 янв.", "21 генваря 1883", "21-го января 1883 года".
	wordRe = regexp.MustCompile(`^(\d{1,2})(?:-го)?[\s.]*([а-я]+)\.?` + yearPart + `$`)
)

// Parse reads a date from text: "21 янв.", "21 генваря 1883", "21 января 1883
// года", "21.01", "21.01.1884". Numeric forms are day.month[.year] with ".",
// "/" or "-" as the separator. The year from the text wins; otherwise ctxYear
// is used and YearFromContext is set. ctxYear 0 means no context: a date
// without a year then fails. Years must have four digits (1000..9999); a
// two-digit year is rejected as ambiguous. Impossible dates (31 февраля, month
// 13, day 0), unknown month words and any other text give ok=false.
func Parse(text string, ctxYear int) (Date, bool) {
	s := strings.TrimSpace(textnorm.Orthography(textnorm.PreReform, text))
	s = strings.Join(strings.Fields(s), " ")

	var dayS, monthS, yearS string

	if m := numericRe.FindStringSubmatch(s); m != nil {
		dayS, monthS, yearS = m[1], m[2], m[3]
	} else if m := wordRe.FindStringSubmatch(s); m != nil {
		mon, ok := Month(m[2])
		if !ok {
			return Date{}, false
		}

		dayS, monthS, yearS = m[1], strconv.Itoa(mon), m[3]
	} else {
		return Date{}, false
	}

	return build(dayS, monthS, yearS, ctxYear)
}

// build converts the captured digits to a validated Date.
func build(dayS, monthS, yearS string, ctxYear int) (Date, bool) {
	day, err1 := strconv.Atoi(dayS)
	month, err2 := strconv.Atoi(monthS)

	if err1 != nil || err2 != nil {
		return Date{}, false
	}

	d := Date{Month: month, Day: day}

	if yearS != "" {
		y, err := strconv.Atoi(yearS)
		if err != nil {
			return Date{}, false
		}

		d.Year = y
	} else {
		d.Year, d.YearFromContext = ctxYear, true
	}

	if d.Year < minYear || d.Year > maxYear || month < 1 || month > 12 || day < 1 {
		return Date{}, false
	}

	t := time.Date(d.Year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if t.Day() != day || int(t.Month()) != month {
		return Date{}, false
	}

	return d, true
}
