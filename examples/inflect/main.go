// Command inflect asks the registry for word forms by grammemes: plural
// headings and number-dependent labels («1 уезд», «2 уезда», «5 уездов»).
// Registry.Inflect picks the source reading by grammemes and returns the
// forms that have the wanted ones; NumeralGrammemes gives the grammemes for
// a count. A word the dictionaries do not know gives no forms — the host
// keeps a fallback. The base here is a tiny TSV in code; a real host uses
// the OpenCorpora base (see examples/base).
//
// Run: go run ./examples/inflect
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/amarin/lexicon"
)

// base stands in for the OpenCorpora base: lemma<TAB>wordform<TAB>tag.
const base = `уезд	уезд	NOUN,inan,masc sing,nomn
уезд	уезда	NOUN,inan,masc sing,gent
уезд	уезды	NOUN,inan,masc plur,nomn
уезд	уездов	NOUN,inan,masc plur,gent
волость	волость	NOUN,inan,femn sing,nomn
волость	волости	NOUN,inan,femn sing,gent
волость	волости	NOUN,inan,femn plur,nomn
волость	волостей	NOUN,inan,femn plur,gent
`

// noun selects the source reading: the nominative singular of a noun.
var noun = []string{"NOUN", "nomn", "sing"}

func main() {
	reg, err := lexicon.Open(context.Background(), lexicon.Options{
		Builtin: []lexicon.BuiltinDict{{Name: "base.example", Format: lexicon.FormatTSV, Data: []byte(base)}},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = reg.Close() }()

	fmt.Println("headings:", plural(reg, "уезд"), plural(reg, "волость"), plural(reg, "стан"))

	for _, word := range []string{"уезд", "волость", "стан"} {
		var labels []string
		for _, n := range []int{1, 2, 5, 11, 21} {
			labels = append(labels, count(reg, n, word))
		}
		fmt.Println(strings.Join(labels, ", "))
	}

	// Output:
	// headings: уезды волости стан
	// 1 уезд, 2 уезда, 5 уездов, 11 уездов, 21 уезд
	// 1 волость, 2 волости, 5 волостей, 11 волостей, 21 волость
	// 1 × стан, 2 × стан, 5 × стан, 11 × стан, 21 × стан
}

// plural is the nominative plural of word, or word itself when the
// dictionaries have no form.
func plural(reg *lexicon.Registry, word string) string {
	if forms := reg.Inflect(word, nil, noun, []string{"plur", "nomn"}); len(forms) > 0 {
		return forms[0]
	}

	return word
}

// count is «n word» with the word agreeing with n, or a neutral «n × word»
// when the dictionaries have no form.
func count(reg *lexicon.Registry, n int, word string) string {
	if forms := reg.Inflect(word, nil, noun, lexicon.NumeralGrammemes(n)); len(forms) > 0 {
		return fmt.Sprintf("%d %s", n, forms[0])
	}

	return fmt.Sprintf("%d × %s", n, word)
}
