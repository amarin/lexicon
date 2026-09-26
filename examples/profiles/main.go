// Command profiles parses the same words under three field profiles. A
// Profile picks the dictionary kinds a field is looked up in and, per kind,
// the grammemes a reading must carry — the main defence against homonymy of
// names and common words: in a "name" field «Вера» is a given name, not
// "faith", and «Мороз» a surname, not "frost". Host-defined kinds (here
// "surname" and "toponym") are declared in Options.Kinds.
//
// Run: go run ./examples/profiles
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
const base = `вера	вера	NOUN,inan,femn sing,nomn
вера	вера	NOUN,anim,femn,Name sing,nomn
мороз	мороз	NOUN,inan,masc sing,nomn
покровское	покровском	NOUN,inan,neut,Geox sing,loct
покровский	покровском	ADJF,Qual masc,sing,loct
`

// surnames is a host dictionary of kind "surname".
const surnames = `мороз	мороз	NOUN,anim,masc,Surn sing,nomn
`

func main() {
	reg, err := lexicon.Open(context.Background(), lexicon.Options{
		Kinds: []lexicon.Kind{"surname", "toponym"},
		Builtin: []lexicon.BuiltinDict{
			{Name: "base.demo", Format: lexicon.FormatTSV, Data: []byte(base)},
			{Name: "surname.demo", Format: lexicon.FormatTSV, Data: []byte(surnames)},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = reg.Close() }()

	profiles := []lexicon.Profile{
		{Name: "text"}, // every dictionary, every reading
		{
			Name:      "name",
			Kinds:     []lexicon.Kind{lexicon.KindBase, "surname"},
			Grammemes: map[lexicon.Kind][]string{lexicon.KindBase: {"Name", "Surn", "Patr"}},
		},
		{
			Name:      "place",
			Kinds:     []lexicon.Kind{lexicon.KindBase, "toponym"},
			Grammemes: map[lexicon.Kind][]string{lexicon.KindBase: {"Geox"}},
		},
	}

	a := lexicon.NewAnalyzer(reg, textnorm.Modern, lexicon.AnalyzerOptions{})
	for _, p := range profiles {
		fmt.Printf("profile %q\n", p.Name)
		for _, t := range a.Analyze("Вера Мороз Покровском", p, lexicon.ModeIndex) {
			fmt.Printf("  %-11s %s\n", t.Form, lemmas(t))
		}
	}

	// Output:
	// profile "text"
	//   вера        вера (base.demo: NOUN,inan,femn sing,nomn)
	//   мороз       мороз (base.demo: NOUN,inan,masc sing,nomn)
	//   покровском  покровское[ambiguous] (base.demo: NOUN,inan,neut,Geox sing,loct) | покровский[ambiguous] (base.demo: ADJF,Qual masc,sing,loct)
	// profile "name"
	//   вера        вера (base.demo: NOUN,anim,femn,Name sing,nomn)
	//   мороз       мороз (surname.demo: NOUN,anim,masc,Surn sing,nomn)
	//   покровском  покровском[unknown]
	// profile "place"
	//   вера        вера[unknown]
	//   мороз       мороз[unknown]
	//   покровском  покровское (base.demo: NOUN,inan,neut,Geox sing,loct)
}

// lemmas joins lemma texts with their flags and the dictionary and tag of the
// reading that gave the lemma.
func lemmas(t lexicon.Term) string {
	var out []string
	for _, l := range t.Lemmas {
		s := l.Text
		if l.Flags != 0 {
			s += "[" + l.Flags.String() + "]"
		}
		if l.Kind != "" {
			s += " (" + l.Dict + ": " + l.Tag + ")"
		}
		out = append(out, s)
	}

	return strings.Join(out, " | ")
}
