// Command orthography compares the Modern and PreReform rule sets: case,
// ё→е, pre-reform letters (ѣ, і, ѳ), the final hard sign, and Latin letters
// typed or recognised (OCR) inside Cyrillic words. ReformVariants gives the
// modern spellings of pre-reform adjective endings that the Analyzer tries
// when a form is not in the dictionaries.
//
// Run: go run ./examples/orthography
package main

import (
	"fmt"

	"github.com/amarin/lexicon/textnorm"
)

func main() {
	fmt.Println("word               modern             prereform")
	for _, w := range []string{"Ёлка", "Вѣра", "Іоаннъ", "Ѳеодоръ", "Kот", "Iоаннъ", "Санктъ-Петербургъ", "XIX"} {
		fmt.Printf("%-18s %-18s %s\n", w, textnorm.NormalizeWord(textnorm.Modern, w), textnorm.NormalizeWord(textnorm.PreReform, w))
	}

	fmt.Println()
	for _, w := range []string{"покровскаго", "большаго", "синяго", "новыя", "покровского"} {
		v := textnorm.ReformVariants(w)
		if v == nil {
			fmt.Printf("%-12s → not a pre-reform ending\n", w)

			continue
		}
		fmt.Printf("%-12s → %v\n", w, v)
	}

	// Output:
	// word               modern             prereform
	// Ёлка               елка               елка
	// Вѣра               вѣра               вера
	// Іоаннъ             іоаннъ             иоанн
	// Ѳеодоръ            ѳеодоръ            феодор
	// Kот                кот                кот
	// Iоаннъ             iоаннъ             иоанн
	// Санктъ-Петербургъ  санктъ-петербургъ  санкт-петербург
	// XIX                xix                xix
	//
	// покровскаго  → [покровского]
	// большаго     → [большого большего]
	// синяго       → [синего]
	// новыя        → [новые новой]
	// покровского  → not a pre-reform ending
}
