package textnorm

import (
	"cmp"
	"fmt"
	"slices"
	"unicode/utf16"
	"unicode/utf8"
)

// UTF16Offsets converts byte offsets of input to UTF-16 code-unit offsets (for
// JavaScript clients). Offsets may come in any order; the result is in the
// same order. An offset inside a code point counts that code point; invalid
// UTF-8 bytes count as one unit each (U+FFFD). Panics on an offset outside
// [0, len(input)].
func UTF16Offsets(input string, byteOffs ...int) []int {
	order := make([]int, len(byteOffs))
	for i := range order {
		order[i] = i
	}

	slices.SortFunc(order, func(a, b int) int { return cmp.Compare(byteOffs[a], byteOffs[b]) })

	out := make([]int, len(byteOffs))
	pos, units := 0, 0

	for _, k := range order {
		off := byteOffs[k]
		if off < 0 || off > len(input) {
			panic(fmt.Sprintf("textnorm: byte offset %d out of range [0, %d]", off, len(input)))
		}

		for pos < off {
			c, size := utf8.DecodeRuneInString(input[pos:])
			units += utf16.RuneLen(c)
			pos += size
		}

		out[k] = units
	}

	return out
}
