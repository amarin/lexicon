package dates

import (
	"strings"

	"github.com/amarin/lexicon/textnorm"
)

// months maps a normalized (lower case, modern spelling) month word, without a
// trailing dot, to its number. Pre-reform letters never appear here:
// textnorm.PreReform has already folded them ("іюня" → "июня").
var months = map[string]int{
	"январь": 1, "января": 1, "генварь": 1, "генваря": 1, "янв": 1, "генв": 1,
	"февраль": 2, "февраля": 2, "фев": 2, "февр": 2,
	"март": 3, "мартъ": 3, "марта": 3, "мар": 3,
	"апрель": 4, "апреля": 4, "апр": 4, "апрел": 4,
	"май": 5, "мая": 5,
	"июнь": 6, "июня": 6, "июн": 6,
	"июль": 7, "июля": 7, "июл": 7,
	"август": 8, "августъ": 8, "августа": 8, "авг": 8,
	"сентябрь": 9, "сентября": 9, "сен": 9, "сент": 9, "сентяб": 9,
	"октябрь": 10, "октября": 10, "окт": 10, "октяб": 10,
	"ноябрь": 11, "ноября": 11, "ноя": 11, "нояб": 11, "ноябр": 11,
	"декабрь": 12, "декабря": 12, "дек": 12, "декаб": 12,
}

// Month returns the month number 1..12 for a month word: nominative and
// genitive case, abbreviations with or without a trailing dot ("янв.",
// "сент."), pre-reform spelling ("іюня", "генваря", "ѳевраля"), any letter
// case. Surrounding spaces are ignored. It reports false for anything else.
func Month(word string) (int, bool) {
	w := strings.TrimSuffix(strings.Trim(textnorm.Orthography(textnorm.PreReform, word), " \t\r\n\u00a0"), ".")
	n, ok := months[w]

	return n, ok
}
