package nertest_test

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
	"github.com/amarin/lexicon/textnorm"
)

// Run scores a pipeline against golden cases: strict (exact range and type)
// and partial (overlapping range, same type) precision and recall per type.
// Check lists the types below the thresholds.
func ExampleRun() {
	ctx := context.Background()
	reg, err := lexicon.Open(ctx, lexicon.Options{
		Builtin: []lexicon.BuiltinDict{{Name: "base.example", Format: lexicon.FormatTSV,
			Data: []byte("из\tиз\tPREP\nуезд\tуезда\tNOUN,inan,masc sing,gent\n")}},
	})
	if err != nil {
		log.Fatal(err)
	}
	an := lexicon.NewAnalyzer(reg, textnorm.Modern, lexicon.AnalyzerOptions{})
	text := lexicon.Profile{Name: "text"}
	gz, err := gazetteer.New(ctx, gazetteer.Config{
		Analyzer: an, DefaultProfile: text,
		Sources: []gazetteer.Source{gazetteer.NewSliceSource("places", "1", []gazetteer.Entry{
			{Type: "division", Ref: "d1", Canonical: "Лягушкино", Alias: "Лягушкино"},
		})},
	})
	if err != nil {
		log.Fatal(err)
	}
	p, err := ner.New(ner.Config{Analyzer: an, Gazetteer: gz,
		Profiles: map[string]lexicon.Profile{"text": text}, DefaultProfile: "text"})
	if err != nil {
		log.Fatal(err)
	}

	cases, err := nertest.LoadCases(strings.NewReader(
		`{"id": "1", "text": "из Лягушкино", "spans": [{"text": "Лягушкино", "type": "division"}]}
{"id": "2", "text": "из Сидоровки Шуйского уезда", "spans": [{"text": "Сидоровки", "type": "division"}, {"text": "Шуйского уезда", "type": "division"}]}
`))
	if err != nil {
		log.Fatal(err)
	}
	rep, err := nertest.Run(ctx, p, cases)
	if err != nil {
		log.Fatal(err)
	}
	if err := rep.Write(os.Stdout); err != nil {
		log.Fatal(err)
	}
	fmt.Println(rep.Check(0.9, 0.9))
	// Output:
	// TYPE      P      R      F1     P~     R~     F1~    TP  FP  FN
	// division  1.000  0.333  0.500  1.000  0.333  0.500  1   0   2
	// cases: 2, failures: 2
	// missed	2	division	«Сидоровки»
	// missed	2	division	«Шуйского уезда»
	// [division: strict recall 0.333 < 0.900]
}
