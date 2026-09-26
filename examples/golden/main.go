// Command golden measures extraction quality against a golden set: JSON
// lines, each a text with the spans it must yield (surface text and type).
// nertest.Run reports strict (exact bytes) and partial (overlap) precision
// and recall per type plus every missed or spurious span; Report.Check
// turns that into a regression gate for tests or CI. A case's context is
// host data: WithTags derives document tags from it, which switch rule
// sets on. The morphology is a tiny TSV registered in code, so the example
// needs no download.
//
// Run: go run ./examples/golden
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/gazetteer"
	"github.com/amarin/lexicon/ner"
	"github.com/amarin/lexicon/nertest"
	"github.com/amarin/lexicon/rules"
	"github.com/amarin/lexicon/textnorm"
)

// base is a stand-in for the OpenCorpora base: lemma<TAB>wordform<TAB>tag.
const base = `крестьянин	крестьянин	NOUN,anim,masc sing,nomn
иван	иван	NOUN,anim,masc,Name sing,nomn
иван	ивана	NOUN,anim,masc,Name sing,gent
мороз	мороз	NOUN,inan,masc sing,nomn
ударить	ударил	VERB,perf,tran masc,sing,past,indc
`

// people is the gazetteer: type<TAB>ref<TAB>canonical<TAB>alias[<TAB>flags].
const people = `given_name	g1	Иван	Иван
surname	s1	Мороз	Мороз	requires_context
`

// ruleFile makes «крестьянин X» a surname context for peasant records.
const ruleFile = `
sets:
  - name: peasant
    when: ["estate:peasant"]
    hints:
      - {lemma: крестьянин, type: surname, window: 2}
`

// cases is the golden set; "context" is host metadata of the document.
const cases = `{"id": "frost", "text": "ударил мороз", "spans": []}
{"id": "peasant", "text": "крестьянин Иван Мороз", "context": {"estate": "peasant"}, "spans": [{"text": "Иван", "type": "given_name"}, {"text": "Мороз", "type": "surname"}]}
`

func main() {
	ctx := context.Background()

	p, closeFn := pipeline(ctx)
	defer closeFn()

	cs, err := nertest.LoadCases(strings.NewReader(cases))
	if err != nil {
		log.Fatal(err)
	}

	// Without WithTags the context is ignored and the surname is missed.
	rep, err := nertest.Run(ctx, p, cs)
	if err != nil {
		log.Fatal(err)
	}
	if err := rep.Write(os.Stdout); err != nil {
		log.Fatal(err)
	}
	fmt.Println("check:", rep.Check(1, 1))

	// The host maps its document metadata to tags: estate=peasant → estate:peasant.
	byContext := nertest.WithTags(func(c nertest.Case) []string {
		var tags []string
		for k, v := range c.Context {
			tags = append(tags, k+":"+v)
		}
		return tags
	})
	rep, err = nertest.Run(ctx, p, cs, byContext)
	if err != nil {
		log.Fatal(err)
	}
	if err := rep.Write(os.Stdout); err != nil {
		log.Fatal(err)
	}
	fmt.Println("check:", rep.Check(1, 1))

	// Output:
	// TYPE        P      R      F1     P~     R~     F1~    TP  FP  FN
	// given_name  1.000  1.000  1.000  1.000  1.000  1.000  1   0   0
	// surname     1.000  0.000  0.000  1.000  0.000  0.000  0   0   1
	// cases: 2, failures: 1
	// missed	peasant	surname	«Мороз»
	// check: [surname: strict recall 0.000 < 1.000]
	// TYPE        P      R      F1     P~     R~     F1~    TP  FP  FN
	// given_name  1.000  1.000  1.000  1.000  1.000  1.000  1   0   0
	// surname     1.000  1.000  1.000  1.000  1.000  1.000  1   0   0
	// cases: 2, failures: 0
	// check: []
}

// pipeline builds the extractor under test.
func pipeline(ctx context.Context) (*ner.Pipeline, func()) {
	reg, err := lexicon.Open(ctx, lexicon.Options{
		Builtin: []lexicon.BuiltinDict{{Name: "base.demo", Format: lexicon.FormatTSV, Data: []byte(base)}},
	})
	if err != nil {
		log.Fatal(err)
	}

	an := lexicon.NewAnalyzer(reg, textnorm.Modern, lexicon.AnalyzerOptions{})
	text := lexicon.Profile{Name: "text"}

	gz, err := gazetteer.New(ctx, gazetteer.Config{
		Analyzer: an, DefaultProfile: text,
		Sources: []gazetteer.Source{gazetteer.NewTSVSourceData("people", []byte(people))},
	})
	if err != nil {
		log.Fatal(err)
	}

	rf, err := rules.LoadNamed("peasant.yaml", strings.NewReader(ruleFile))
	if err != nil {
		log.Fatal(err)
	}
	book, err := rules.Compile(rf)
	if err != nil {
		log.Fatal(err)
	}

	p, err := ner.New(ner.Config{
		Analyzer: an, Gazetteer: gz, Rules: book,
		Profiles: map[string]lexicon.Profile{"text": text}, DefaultProfile: "text",
	})
	if err != nil {
		log.Fatal(err)
	}

	return p, func() { _ = reg.Close() }
}
