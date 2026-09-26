// Command index turns text into search-index terms (ModeIndex): one term per
// indexable word with its lemmas, service words dropped, numbers kept, a
// hyphenated word indexed whole and by parts. The base dictionary is a tiny
// TSV registered in code, so the example needs no download; with it,
// «кормить» is unknown and «Иван-кот» is guessed from its ending — the real
// OpenCorpora base knows both (go run ./examples/base).
//
// Run: go run ./examples/index
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/textnorm"
)

// base is a stand-in for the OpenCorpora base: lemma<TAB>wordform<TAB>tag.
const base = `кот	кот	NOUN,anim,masc sing,nomn
кот	кота	NOUN,anim,masc sing,gent
кот	котом	NOUN,anim,masc sing,ablt
пёс	пса	NOUN,anim,masc sing,gent
сталь	стали	NOUN,inan,femn sing,gent
стать	стали	VERB,perf,intr plur,past,indc
и	и	CONJ
с	с	PREP
жить	жил	VERB,impf,intr masc,sing,past,indc
в	в	PREP
год	году	NOUN,inan,masc sing,loct
`

// lemmas prints the lemmas of a term as the CLI does: text[flags].
func lemmas(t lexicon.Term) string {
	var s string
	for i, l := range t.Lemmas {
		if i > 0 {
			s += " "
		}
		s += l.Text
		if l.Flags != 0 {
			s += "[" + l.Flags.String() + "]"
		}
	}

	return s
}

func main() {
	ctx := context.Background()

	reg, err := lexicon.Open(ctx, lexicon.Options{
		Builtin: []lexicon.BuiltinDict{{Name: "base.demo", Format: lexicon.FormatTSV, Data: []byte(base)}},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = reg.Close() }()

	a := lexicon.NewAnalyzer(reg, textnorm.Modern, lexicon.AnalyzerOptions{})
	text := "Кота и пса стали кормить в 1834 году; Иван-кот"

	for _, t := range a.Analyze(text, lexicon.Profile{Name: "text"}, lexicon.ModeIndex) {
		fmt.Printf("%-9s %-9s %s\n", t.Token.Raw, t.Form, lemmas(t))
	}

	// Output:
	// Кота      кота      кот
	// пса       пса       пес
	// стали     стали     сталь[ambiguous] стать[ambiguous]
	// кормить   кормить   кормить[unknown]
	// 1834      1834      1834[unknown]
	// году      году      год
	// Иван-кот  иван-кот  иван-кот[predicted]
	// Иван      иван      иван[unknown]
	// кот       кот       кот
}
