// Command full marks up a pre-reform text with ModeFull: exactly one term per
// token (punctuation included), each with offsets into the original text and
// flagged lemmas — the input for NER, pre-annotation or a review UI. Service
// words keep their lemmas with the "stop" flag instead of being dropped; a
// form found through a modern spelling of its pre-reform ending carries
// "reform"; words the dictionaries do not know are "predicted" (guessed from
// the ending by the base dictionary) or "unknown".
//
// Run: go run ./examples/full
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/textnorm"
)

// base is a stand-in for the OpenCorpora base: lemma<TAB>wordform<TAB>tag.
const base = `иван	иван	NOUN,anim,masc,Name sing,nomn
кузнецов	кузнецов	NOUN,anim,masc,Surn sing,nomn
жить	жил	VERB,impf,intr masc,sing,past,indc
в	в	PREP
покровский	покровского	ADJF,Geox masc,sing,gent
покровский	покровском	ADJF,Geox masc,sing,loct
уезд	уезде	NOUN,inan,masc sing,loct
шуйский	шуйского	ADJF,Geox masc,sing,gent
`

func main() {
	reg, err := lexicon.Open(context.Background(), lexicon.Options{
		Builtin: []lexicon.BuiltinDict{{Name: "base.demo", Format: lexicon.FormatTSV, Data: []byte(base)}},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = reg.Close() }()

	a := lexicon.NewAnalyzer(reg, textnorm.PreReform, lexicon.AnalyzerOptions{})
	text := "Иванъ Кузнецовъ жилъ въ Покровскомъ уѣздѣ Шуйскаго у., Петровъ – тоже"

	for _, t := range a.Analyze(text, lexicon.Profile{Name: "text"}, lexicon.ModeFull) {
		fmt.Printf("%2d-%-2d %-12s %s\n", t.Token.RuneStart, t.Token.RuneEnd, t.Token.Raw, lemmas(t))
	}

	// Output:
	//  0-5  Иванъ        иван
	//  6-15 Кузнецовъ    кузнецов
	// 16-20 жилъ         жить
	// 21-23 въ           в[stop]
	// 24-35 Покровскомъ  покровский
	// 36-41 уѣздѣ        уезд
	// 42-50 Шуйскаго     шуйский[reform]
	// 51-52 у            у[unknown]
	// 52-53 .            -
	// 53-54 ,            -
	// 55-62 Петровъ      петров[predicted]
	// 63-64 –            -
	// 65-69 тоже         тож[predicted]
}

// lemmas joins lemma texts with their flags; "-" for punctuation.
func lemmas(t lexicon.Term) string {
	if t.Lemmas == nil {
		return "-"
	}
	var out []string
	for _, l := range t.Lemmas {
		s := l.Text
		if l.Flags != 0 {
			s += "[" + l.Flags.String() + "]"
		}
		out = append(out, s)
	}

	return strings.Join(out, " ")
}
