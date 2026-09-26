package textnorm_test

import (
	"fmt"

	"github.com/amarin/lexicon/textnorm"
)

// Every token carries byte and code-point offsets into the input as passed:
// input[Start:End] is the token's text whatever its normalized Form.
func ExampleTokenize() {
	input := "Иванъ, 1834 г."
	for _, t := range textnorm.Tokenize(textnorm.PreReform, input) {
		fmt.Printf("%q form=%q bytes=%d-%d runes=%d-%d dotted=%t\n",
			input[t.Start:t.End], t.Form, t.Start, t.End, t.RuneStart, t.RuneEnd, t.Dotted)
	}
	// Output:
	// "Иванъ" form="иван" bytes=0-10 runes=0-5 dotted=false
	// "," form="" bytes=10-11 runes=5-6 dotted=false
	// "1834" form="1834" bytes=12-16 runes=7-11 dotted=false
	// "г" form="г" bytes=17-19 runes=12-13 dotted=true
	// "." form="" bytes=19-20 runes=13-14 dotted=false
}

// NormalizeWord is the form a word is indexed and compared by.
func ExampleNormalizeWord() {
	for _, w := range []string{"Ёлка", "Вѣра", "Санктъ-Петербургъ", "Kот"} {
		fmt.Println(textnorm.NormalizeWord(textnorm.Modern, w), textnorm.NormalizeWord(textnorm.PreReform, w))
	}
	// Output:
	// елка елка
	// вѣра вера
	// санктъ-петербургъ санкт-петербург
	// кот кот
}

// Orthography maps a whole string for comparison; the final hard sign is a
// word rule and stays.
func ExampleOrthography() {
	fmt.Println(textnorm.Orthography(textnorm.PreReform, "Ѳеодоръ Iоанновичъ"))
	// Output: феодоръ иоанновичъ
}

// Pre-reform adjective endings have modern variants, tried in order.
func ExampleReformVariants() {
	fmt.Println(textnorm.ReformVariants("покровскаго"))
	fmt.Println(textnorm.ReformVariants("хорошаго"))
	fmt.Println(textnorm.ReformVariants("покровского") == nil)
	// Output:
	// [покровского]
	// [хорошого хорошего]
	// true
}

// The parts of a hyphenated word keep offsets into the same input.
func ExampleSplitHyphen() {
	input := "Санктъ-Петербургъ"
	for _, p := range textnorm.SplitHyphen(textnorm.Tokenize(textnorm.PreReform, input)[0]) {
		fmt.Println(input[p.Start:p.End], p.Form)
	}
	// Output:
	// Санктъ санкт
	// Петербургъ петербург
}

// JavaScript strings index UTF-16 code units: an emoji takes two.
func ExampleUTF16Offsets() {
	input := "🎄 Ёлка"
	t := textnorm.Tokenize(textnorm.Modern, input)[1]
	fmt.Println(t.Start, t.End, t.RuneStart, t.RuneEnd, textnorm.UTF16Offsets(input, t.Start, t.End))
	// Output: 5 13 2 6 [3 7]
}
