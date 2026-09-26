// Command query parses search-box input with ParseQuery: complete words get
// the same terms as the index (ModeIndex), and the last word, while the user
// is still typing it, becomes Partial — search it by prefix over indexed
// forms and by its lemmas. The same Analyzer builds the index and parses the
// query, so the two cannot diverge.
//
// Run: go run ./examples/query
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
const base = `кот	кот	NOUN,anim,masc sing,nomn
кот	кота	NOUN,anim,masc sing,gent
кузнецов	кузнецов	NOUN,anim,masc,Surn sing,nomn
кузнецов	кузнецова	NOUN,anim,masc,Surn sing,gent
в	в	PREP
`

func main() {
	reg, err := lexicon.Open(context.Background(), lexicon.Options{
		Builtin: []lexicon.BuiltinDict{{Name: "base.demo", Format: lexicon.FormatTSV, Data: []byte(base)}},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = reg.Close() }()

	a := lexicon.NewAnalyzer(reg, textnorm.Modern, lexicon.AnalyzerOptions{})
	text := lexicon.Profile{Name: "text"}

	for _, q := range []string{"Кузнецова", "Кузнецова ", "кота в", "кота Кузн"} {
		got := a.ParseQuery(q, text)
		fmt.Printf("%-13q terms: %-20s partial: %s\n", q, terms(got.Terms), partial(got.Partial))
	}

	// Output:
	// "Кузнецова"   terms: -                    partial: кузнецова* or кузнецов
	// "Кузнецова "  terms: кузнецова=кузнецов   partial: -
	// "кота в"      terms: кота=кот             partial: в* (service word: prefix only)
	// "кота Кузн"   terms: кота=кот             partial: кузн* or кузн[unknown]
}

// terms lists form=lemma|lemma for complete words.
func terms(ts []lexicon.Term) string {
	var out []string
	for _, t := range ts {
		out = append(out, t.Form+"="+lemmas(t))
	}
	if out == nil {
		return "-"
	}

	return strings.Join(out, " ")
}

// partial shows the prefix and the lemmas to search it by.
func partial(t *lexicon.Term) string {
	if t == nil {
		return "-"
	}
	if t.Lemmas == nil {
		return t.Form + "* (service word: prefix only)"
	}

	return t.Form + "* or " + lemmas(*t)
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

	return strings.Join(out, "|")
}
