package gazetteer_test

import (
	"context"
	"fmt"
	"log"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/gazetteer"
	"github.com/amarin/lexicon/textnorm"
)

// exampleAnalyzer analyzes text over a tiny TSV stand-in for the OpenCorpora
// base (lemma<TAB>wordform<TAB>tag), so the examples need no download.
func exampleAnalyzer() *lexicon.Analyzer {
	const base = "боровский\tборовский\tADJF,Geox masc,sing,nomn\n" +
		"боровский\tборовского\tADJF,Geox masc,sing,gent\n" +
		"уезд\tуезда\tNOUN,inan,masc sing,gent\n" +
		"иван\tиван\tNOUN,anim,masc,Name sing,nomn\n"
	reg, err := lexicon.Open(context.Background(), lexicon.Options{
		Builtin: []lexicon.BuiltinDict{{Name: "base.example", Format: lexicon.FormatTSV, Data: []byte(base)}},
	})
	if err != nil {
		log.Fatal(err)
	}
	return lexicon.NewAnalyzer(reg, textnorm.Modern, lexicon.AnalyzerOptions{})
}

// A gazetteer compiles host aliases into lemma and surface keys; matching an
// analyzed text finds «Боровского» by the lemma of the alias «Боровский».
// Match positions count content words (words and numbers) of the text.
func ExampleNew() {
	ctx := context.Background()
	an := exampleAnalyzer()
	const tsv = "# source: example\n" +
		"division\td2\tБоровский\tБоровский\t\tlevel=uezd\n" +
		"given_name\tg1\tИван\tИван\n" +
		"given_name\tg1\tИван\tИоанн\n"
	gz, err := gazetteer.New(ctx, gazetteer.Config{
		Analyzer:       an,
		DefaultProfile: lexicon.Profile{Name: "text"},
		Sources:        []gazetteer.Source{gazetteer.NewTSVSourceData("places", []byte(tsv))},
	})
	if err != nil {
		log.Fatal(err)
	}

	tx := gazetteer.Prepare(an.Analyze("Иоанн из Боровского уезда", lexicon.Profile{Name: "text"}, lexicon.ModeFull))
	for _, m := range gz.Snapshot().Match(tx, nil) {
		for _, a := range m.Aliases {
			kind := "surface"
			if m.Kind == gazetteer.ByLemma {
				kind = "lemma"
			}
			fmt.Println(m.Start, m.End, kind, a.Entry.Type, a.Entry.Ref, a.Entry.Canonical)
		}
	}
	r := gz.Snapshot().Reports()[0]
	fmt.Println(r.Source, r.Entries, r.Aliases, len(r.Errors))
	// Output:
	// 0 1 lemma given_name g1 Иван
	// 0 1 surface given_name g1 Иван
	// 2 3 lemma division d2 Боровский
	// places 3 3 0
}

// Aliases sharing a Ref are variants of one record: Canonical maps any of
// them to the record's canonical form, Expand gives all variants.
func ExampleGazetteer_Expand() {
	an := exampleAnalyzer()
	gz, err := gazetteer.New(context.Background(), gazetteer.Config{
		Analyzer:       an,
		DefaultProfile: lexicon.Profile{Name: "text"},
		Sources: []gazetteer.Source{gazetteer.NewSliceSource("people", "1", []gazetteer.Entry{
			{Type: "given_name", Ref: "g1", Canonical: "Иван", Alias: "Иван"},
			{Type: "given_name", Ref: "g1", Canonical: "Иван", Alias: "Иоанн"},
		})},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(gz.Canonical("иоанн"), gz.Expand("иоанн"))
	// Output: [Иван] [иван иоанн]
}
