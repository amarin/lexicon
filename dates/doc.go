// Package dates parses calendar dates as written in pre-reform Russian parish
// and civil records: month words in the nominative and genitive case,
// abbreviations ("янв.", "сент."), pre-reform spelling ("іюня", "генваря") and
// numeric forms ("21.01", "21.01.1884"). A year missing from the text can be
// taken from a caller-supplied context year. Dates are Gregorian calendar
// dates; the package does not convert between Julian and Gregorian calendars.
//
// Spelling is normalized with [textnorm.PreReform], so ѣ, і, ѳ and case
// variants are handled by the same rules the analyzer uses.
package dates
