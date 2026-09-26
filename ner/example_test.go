package ner_test

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/gazetteer"
	"github.com/amarin/lexicon/ner"
	"github.com/amarin/lexicon/rules"
	"github.com/amarin/lexicon/textnorm"
)

// examplePipeline builds a pipeline over a tiny TSV stand-in for the
// OpenCorpora base, a places gazetteer and place rules, so the examples need
// no download.
func examplePipeline() *ner.Pipeline {
	ctx := context.Background()
	const base = "деревня\tдеревни\tNOUN,inan,femn sing,gent\n" +
		"село\tсела\tNOUN,inan,neut sing,gent\n" +
		"уезд\tуезда\tNOUN,inan,masc sing,gent\n" +
		"боровский\tборовский\tADJF,Geox masc,sing,nomn\n" +
		"боровский\tборовского\tADJF,Geox masc,sing,gent\n" +
		"из\tиз\tPREP\n"
	reg, err := lexicon.Open(ctx, lexicon.Options{
		Builtin: []lexicon.BuiltinDict{{Name: "base.example", Format: lexicon.FormatTSV, Data: []byte(base)}},
	})
	if err != nil {
		log.Fatal(err)
	}
	an := lexicon.NewAnalyzer(reg, textnorm.Modern, lexicon.AnalyzerOptions{})
	text := lexicon.Profile{Name: "text"}

	const places = "division\td1\tЛягушкино\tЛягушкино\t\tlevel=village\n" +
		"division\td2\tБоровский\tБоровский\t\tlevel=uezd\n"
	gz, err := gazetteer.New(ctx, gazetteer.Config{
		Analyzer: an, DefaultProfile: text,
		Sources: []gazetteer.Source{gazetteer.NewTSVSourceData("places", []byte(places))},
	})
	if err != nil {
		log.Fatal(err)
	}

	const yaml = `sets:
  - name: places
    hints:
      - {lemma: деревня, type: division, window: 1, weight: 2, absorb: true}
      - {lemma: уезд, type: division, dir: left, window: 1, weight: 2, absorb: true}
    triggers:
      - {lemma: село, type: division, shape: {case: title, script: cyrillic}, absorb: true}
`
	f, err := rules.Load(strings.NewReader(yaml))
	if err != nil {
		log.Fatal(err)
	}
	book, err := rules.Compile(f)
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
	return p
}

// Extract marks up spans with offsets into the text, the host's refs, the
// canonical forms and flags; «села Покровское» is a trigger candidate no
// record knows.
func ExamplePipeline_Extract() {
	p := examplePipeline()
	text := "из деревни Лягушкино Боровского уезда и села Покровское"
	res, err := p.Extract(context.Background(), ner.Doc{Text: text})
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range res.Spans {
		fmt.Printf("%d-%d %q %s %v %v %v %.2f\n", s.RuneStart, s.RuneEnd, s.Surface, s.Type, s.Refs, s.Normal, s.Flags.Names(), s.Score)
	}
	// Output:
	// 3-20 "деревни Лягушкино" division [d1] [Лягушкино] [] 8.50
	// 21-37 "Боровского уезда" division [d2] [Боровский] [] 6.50
	// 40-55 "села Покровское" division [] [покровское] [candidate] 3.50
}

// Explain records why each span was produced.
func ExampleExplain() {
	p := examplePipeline()
	res, err := p.Extract(context.Background(), ner.Doc{Text: "из Боровского уезда"}, ner.Explain())
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range res.Spans {
		fmt.Println(s.Surface)
		for _, e := range s.Evidence {
			fmt.Println("-", e)
		}
	}
	// Output:
	// Боровского уезда
	// - lemma match «Боровский» (places)
	// - hint «уезда» → division +2
}
