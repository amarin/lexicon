// Command abbrev expands abbreviations of archival records with a dictionary
// of kind "abbrev": a dotted word is looked up with its dot («с.», «у.»), a
// hyphenated one whole («кр-нин»). A one-letter dotted abbreviation is always
// ambiguous; an undotted one-letter word («с» — a preposition) never takes an
// abbreviation reading. Abbreviation lookups ignore the profile's kinds
// (behaviour of 0.1.0, see docs/en/scenarios.md).
//
// Run: go run ./examples/abbrev
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
const base = `с	с	PREP
жена	женой	NOUN,anim,femn sing,ablt
`

// abbrevs maps each abbreviation (the wordform column, with its dot or
// hyphen) to the full word (the lemma column).
const abbrevs = `крестьянин	кр-нин	NOUN
село	с.	NOUN
сын	с.	NOUN
уезд	у.	NOUN
губерния	губ.	NOUN
`

func main() {
	reg, err := lexicon.Open(context.Background(), lexicon.Options{
		Builtin: []lexicon.BuiltinDict{
			{Name: "base.demo", Format: lexicon.FormatTSV, Data: []byte(base)},
			{Name: "abbrev.records", Format: lexicon.FormatTSV, Data: []byte(abbrevs)},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = reg.Close() }()

	a := lexicon.NewAnalyzer(reg, textnorm.PreReform, lexicon.AnalyzerOptions{})
	text := "Кр-нин с. Ивановка Шуйского у., губ. Владимирская, с женой"

	for _, t := range a.Analyze(text, lexicon.Profile{Name: "text"}, lexicon.ModeIndex) {
		fmt.Printf("%-13s %-13s %s\n", t.Token.Raw, t.Form, lemmas(t))
	}

	// Output:
	// Кр-нин        кр-нин        крестьянин[abbrev]
	// с             с.            село[ambiguous,abbrev] сын[ambiguous,abbrev]
	// Ивановка      ивановка      ивановка[unknown]
	// Шуйского      шуйского      шуйского[unknown]
	// у             у.            уезд[ambiguous,abbrev]
	// губ           губ.          губерния[abbrev]
	// Владимирская  владимирская  владимирская[unknown]
	// женой         женой         жена
}

// lemmas joins lemma texts with their flags.
func lemmas(t lexicon.Term) string {
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
