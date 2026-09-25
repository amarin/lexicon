package textnorm

import "unicode/utf8"

// SplitHyphen returns the hyphen-separated parts of a word token, each with
// offsets into the same input, its own Form and Case, the word's Script;
// Dotted and SentenceEnd go to the last part. nil when t is not a word or has
// no inner hyphen (or when its Form does not split into the same number of
// parts, which the built-in rule sets never cause).
func SplitHyphen(t Token) []Token {
	if t.Kind != TokenWord {
		return nil
	}

	raws := hyphenParts(t.Raw)
	if len(raws) < 2 {
		return nil
	}

	forms := hyphenParts(t.Form)
	if len(forms) != len(raws) {
		return nil
	}

	out := make([]Token, len(raws))
	for i, s := range raws {
		raw := t.Raw[s[0]:s[1]]
		out[i] = Token{
			Raw: raw, Form: t.Form[forms[i][0]:forms[i][1]],
			Start: t.Start + s[0], End: t.Start + s[1],
			RuneStart: t.RuneStart + s[2], RuneEnd: t.RuneStart + s[3],
			Kind: TokenWord, Script: t.Script, Case: caseOf(raw),
		}
	}

	last := &out[len(out)-1]
	last.Dotted, last.SentenceEnd = t.Dotted, t.SentenceEnd

	return out
}

// hyphenParts returns [byteStart, byteEnd, runeStart, runeEnd] of every
// hyphen-separated part of s (one element when s has no hyphen).
func hyphenParts(s string) [][4]int {
	var out [][4]int

	start, runeStart, n := 0, 0, 0

	for off, c := range s {
		if isHyphen(c) {
			out = append(out, [4]int{start, off, runeStart, n})
			start, runeStart = off+utf8.RuneLen(c), n+1
		}

		n++
	}

	return append(out, [4]int{start, len(s), runeStart, n})
}
