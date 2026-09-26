// Command tokenize shows the textnorm tokenizer: every token of the input —
// words, numbers, punctuation, symbols — with byte and code-point offsets
// into the input exactly as passed, so input[Start:End] is always the token's
// text, whatever normalization produced its Form. UTF16Offsets converts
// offsets for a JavaScript client; SplitHyphen gives the parts of a
// hyphenated word with their own offsets.
//
// Run: go run ./examples/tokenize
package main

import (
	"fmt"

	"github.com/amarin/lexicon/textnorm"
)

func main() {
	input := "Ёлка 🎄 у Санктъ-Петербурга."

	fmt.Println("raw                  form                 bytes  runes  utf16  kind")
	for _, t := range textnorm.Tokenize(textnorm.PreReform, input) {
		u := textnorm.UTF16Offsets(input, t.Start, t.End)
		fmt.Printf("%-20s %-20s %2d-%-2d  %2d-%-2d  %2d-%-2d  %s\n",
			quote(t.Raw), quote(t.Form), t.Start, t.End, t.RuneStart, t.RuneEnd, u[0], u[1], kind(t))

		for _, p := range textnorm.SplitHyphen(t) {
			fmt.Printf("  part %-15s %-20s %2d-%-2d\n", quote(p.Raw), quote(p.Form), p.Start, p.End)
		}
	}

	// Output:
	// raw                  form                 bytes  runes  utf16  kind
	// «Ёлка»               «елка»                0-8    0-4    0-4   word
	// «🎄»                  «»                    9-13   5-6    5-7   symbol
	// «у»                  «у»                  14-16   7-8    8-9   word
	// «Санктъ-Петербурга»  «санкт-петербурга»   17-50   9-26  10-27  word, dotted
	//   part «Санктъ»        «санкт»              17-29
	//   part «Петербурга»    «петербурга»         30-50
	// «.»                  «»                   50-51  26-27  27-28  punct, sentence end
}

// quote wraps s in «» so empty forms are visible.
func quote(s string) string { return "«" + s + "»" }

// kind names a token kind and its marks.
func kind(t textnorm.Token) string {
	s := map[textnorm.TokenKind]string{
		textnorm.TokenWord: "word", textnorm.TokenNumber: "number",
		textnorm.TokenPunct: "punct", textnorm.TokenSymbol: "symbol",
	}[t.Kind]
	if t.Dotted {
		s += ", dotted"
	}
	if t.SentenceEnd {
		s += ", sentence end"
	}

	return s
}
