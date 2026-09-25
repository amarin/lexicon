package textnorm

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// Orthography maps s to the comparison orthography of r: NFC, Latin
// homoglyphs of Cyrillic word parts replaced by r.Homoglyphs (before
// lower-casing, so upper-case-only homoglyphs fold too), lower case, combining
// marks and format characters removed except in r.KeepMarks letters, letters
// replaced by r.Letters. The final hard sign is kept: it is a word rule
// (NormalizeWord). The result is for comparison only; offsets never refer to
// it.
func Orthography(r Rules, s string) string {
	runes := []rune(norm.NFC.String(s))
	foldHomoglyphs(r.Homoglyphs, runes)

	var (
		b   strings.Builder
		enc [utf8.UTFMax]byte
		dec [4 * utf8.UTFMax]byte
	)

	b.Grow(len(s))

	for _, c := range runes {
		c = unicode.ToLower(c)

		switch {
		case slices.Contains(r.KeepMarks, c):
			b.WriteRune(c)
		case unicode.IsMark(c), unicode.Is(unicode.Cf, c):
		default:
			d := norm.NFD.Append(dec[:0], utf8.AppendRune(enc[:0], c)...)
			for len(d) > 0 {
				x, n := utf8.DecodeRune(d)
				d = d[n:]

				if unicode.IsMark(x) {
					continue
				}

				if rep, ok := r.Letters[x]; ok {
					b.WriteString(rep)

					continue
				}

				b.WriteRune(x)
			}
		}
	}

	return b.String()
}

// NormalizeWord is the form of a word for indexing and comparison:
// Orthography plus, when r.DropFinalHardSign is set, removal of a final ъ from
// every hyphen part («санктъ-петербургъ» → «санкт-петербург»).
func NormalizeWord(r Rules, w string) string {
	s := Orthography(r, w)
	if !r.DropFinalHardSign {
		return s
	}

	parts := strings.Split(s, "-")
	for i, p := range parts {
		parts[i] = strings.TrimSuffix(p, "ъ")
	}

	return strings.Join(parts, "-")
}

// foldHomoglyphs replaces, in place, the Latin homoglyphs h of every Cyrillic
// word part of runes (isCyrillicPart). A part is a maximal run of letters,
// combining marks and format characters except U+200B (which the tokenizer
// treats as a space); hyphens and all other code points separate parts.
func foldHomoglyphs(h map[rune]rune, runes []rune) {
	if len(h) == 0 {
		return
	}

	for i := 0; i < len(runes); {
		j := i
		for j < len(runes) && inWordPart(runes[j]) {
			j++
		}

		if j == i {
			i++

			continue
		}

		if isCyrillicPart(h, runes[i:j]) {
			for k := i; k < j; k++ {
				if m, ok := h[runes[k]]; ok {
					runes[k] = m
				}
			}
		}

		i = j
	}
}

// inWordPart reports whether c continues a word part: a letter, a combining
// mark or a format character other than U+200B.
func inWordPart(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsMark(c) || (unicode.Is(unicode.Cf, c) && c != '\u200b')
}

// isCyrillicPart reports whether a word part has at least one Cyrillic letter
// and every other letter of it is a key of h (Rules.Homoglyphs). Non-letters
// (marks, format characters) are ignored.
func isCyrillicPart(h map[rune]rune, part []rune) bool {
	cyr := false

	for _, c := range part {
		switch {
		case !unicode.IsLetter(c):
		case unicode.Is(unicode.Cyrillic, c):
			cyr = true
		default:
			if _, ok := h[c]; !ok {
				return false
			}
		}
	}

	return cyr
}
