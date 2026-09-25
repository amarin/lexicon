package textnorm

import "maps"

// homoglyphs are Latin letters that look like Cyrillic ones; both rule sets
// replace them inside Cyrillic word parts (Rules.Homoglyphs, decision D5).
var homoglyphs = map[rune]rune{
	'A': 'А', 'a': 'а', 'B': 'В', 'C': 'С', 'c': 'с', 'E': 'Е', 'e': 'е',
	'H': 'Н', 'K': 'К', 'k': 'к', 'M': 'М', 'O': 'О', 'o': 'о', 'P': 'Р',
	'p': 'р', 'T': 'Т', 'X': 'Х', 'x': 'х', 'y': 'у', 'Y': 'У',
}

// Modern: lower case, NFC, combining marks stripped except in й, ё→е (through
// decomposition), U+2010 hyphen → '-', Latin homoglyphs inside Cyrillic words.
var Modern = Rules{
	Name:       "modern",
	Version:    "2",
	Letters:    map[rune]string{'‐': "-"},
	KeepMarks:  []rune{'й'},
	Homoglyphs: homoglyphs,
}

// PreReform: Modern plus pre-reform and Church Slavonic letters, Latin i/I as
// homoglyphs of і/І and the final hard sign (genodex search P1 table).
// ї and ѷ need no entries: decomposition gives і and ѵ plus a dropped mark.
var PreReform = Rules{
	Name:    "prereform",
	Version: "2",
	Letters: map[rune]string{
		'ѣ': "е", 'і': "и", 'ѵ': "и",
		'ѳ': "ф", 'ѡ': "о", 'ꙋ': "у", 'ѹ': "у", 'ѧ': "я", 'ѯ': "кс", 'ѱ': "пс",
		'‐': "-",
	},
	KeepMarks:         []rune{'й'},
	DropFinalHardSign: true,
	Homoglyphs: func() map[rune]rune {
		h := maps.Clone(homoglyphs)
		h['i'], h['I'] = 'і', 'І'

		return h
	}(),
}
