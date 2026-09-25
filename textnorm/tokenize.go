package textnorm

import "unicode"

// zwj is ZERO WIDTH JOINER.
const zwj = '\u200d'

// Tokenize splits input into tokens with byte and code-point offsets into
// input exactly as passed (input[t.Start:t.End] == t.Raw):
//
//  1. a word is a maximal run of letters of one category — Cyrillic or
//     non-Cyrillic — with the extending code points of each letter, joined
//     over single inner hyphens (-, U+2010) between letters of that category;
//     a hyphen part with a Cyrillic letter whose other letters are all keys
//     of r.Homoglyphs is Cyrillic as a whole («Kот», Form «кот»); in other
//     parts each letter keeps its own category («Kотw» → «K», «от», «w»);
//  2. a number is a run of ASCII digits;
//  3. any other visible code point is a one-cluster punct or symbol token; a
//     regional-indicator pair is one symbol; ZWJ joins a following pictographic
//     symbol;
//  4. extending code points (marks, format characters except U+200B, emoji
//     modifiers) never start a token: they extend the preceding one, or are
//     skipped after whitespace and at the start;
//  5. whitespace (and U+200B) and control characters are not tokens; a line
//     break sets SentenceEnd on the previous token, as do the punctuation
//     tokens . ! ? ; ….
//
// Grapheme clusters are approximated by rule 4 and the ZWJ/regional-indicator
// rules; Prepend characters are not modelled.
func Tokenize(r Rules, input string) []Token {
	t := tokenizer{rules: r, in: input, p: make([]runePos, 0, len(input))}
	for off, c := range input {
		t.p = append(t.p, runePos{off: off, r: c})
	}

	t.classify()
	t.scan()

	return t.out
}

func isDigit(r rune) bool { return r >= '0' && r <= '9' }

func isHyphen(r rune) bool { return r == '-' || r == '\u2010' }

// isExtend: code points that belong to the preceding grapheme cluster.
func isExtend(r rune) bool {
	return unicode.IsMark(r) ||
		(unicode.Is(unicode.Cf, r) && r != '\u200b') ||
		(r >= 0x1F3FB && r <= 0x1F3FF)
}

func isSpace(r rune) bool { return unicode.IsSpace(r) || r == '\u200b' }

func isLineBreak(r rune) bool {
	switch r {
	case '\n', '\r', '\v', '\f', '\u0085', '\u2028', '\u2029':
		return true
	}

	return false
}

func isSentencePunct(r rune) bool {
	switch r {
	case '.', '!', '?', ';', '…':
		return true
	}

	return false
}

func isRegionalIndicator(r rune) bool { return r >= 0x1F1E6 && r <= 0x1F1FF }
