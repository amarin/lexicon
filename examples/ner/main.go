// Command ner extracts entity spans from text with a gazetteer (aliases of
// host records, matched by lemmas or surface forms) and rules: a hint joins
// «деревни» and «уезда» to the place next to them, a trigger proposes «села
// Покровское» that no record knows (flag candidate), and a rule set switched
// on by the document tag period:pre1917 lets «Мороз» be a surname — its
// entry requires context, so without the tag it is dropped as "frost".
// ner.Explain fills the evidence printed under each span. The morphology is
// a tiny TSV registered in code and the gazetteer and rules are strings, so
// the example needs no download.
//
// Run: go run ./examples/ner
package main

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

// base is a stand-in for the OpenCorpora base: lemma<TAB>wordform<TAB>tag.
const base = `деревня	деревня	NOUN,inan,femn sing,nomn
деревня	деревни	NOUN,inan,femn sing,gent
село	село	NOUN,inan,neut sing,nomn
село	села	NOUN,inan,neut sing,gent
уезд	уезда	NOUN,inan,masc sing,gent
лягушкино	лягушкино	NOUN,inan,neut,Geox,Fixd sing,nomn
боровский	боровский	ADJF,Geox masc,sing,nomn
боровский	боровского	ADJF,Geox masc,sing,gent
крестьянин	крестьянин	NOUN,anim,masc sing,nomn
иван	иван	NOUN,anim,masc,Name sing,nomn
иван	ивана	NOUN,anim,masc,Name sing,gent
петров	петров	NOUN,anim,masc,Surn sing,nomn
мороз	мороз	NOUN,inan,masc sing,nomn
из	из	PREP
`

// places and people are gazetteer sources in TSV:
// type<TAB>ref<TAB>canonical<TAB>alias[<TAB>flags[<TAB>k=v;k=v]].
const places = `# source: lexicon example
division	d1	Лягушкино	Лягушкино		level=village
division	d2	Боровский	Боровский		level=uezd
`

// «Мороз» is also "frost": the surname needs a supporting rule.
const people = `given_name	g1	Иван	Иван
given_name	g1	Иван	Иоанн
surname	s1	Петров	Петров
surname	s2	Мороз	Мороз	requires_context
`

// ruleFile holds a hint (a keyword joining an adjacent span), a trigger (a
// keyword proposing a span the gazetteer does not know) and a rule set
// active only for documents tagged period:pre1917.
const ruleFile = `
sets:
  - name: places
    hints:
      - {lemma: деревня|село, type: division, window: 1, weight: 2, absorb: true}
      - {lemma: уезд, type: division, dir: left, window: 1, weight: 2, absorb: true}
    triggers:
      - lemma: деревня|село
        type: division
        window: 1..2
        shape: {case: title, script: cyrillic}
        absorb: true
  - name: pre1917
    when: ["period:pre1917"]
    hints:
      - {lemma: крестьянин, type: surname, window: 2}
`

func main() {
	ctx := context.Background()

	reg, err := lexicon.Open(ctx, lexicon.Options{
		Builtin: []lexicon.BuiltinDict{{Name: "base.demo", Format: lexicon.FormatTSV, Data: []byte(base)}},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = reg.Close() }()

	an := lexicon.NewAnalyzer(reg, textnorm.PreReform, lexicon.AnalyzerOptions{})
	text := lexicon.Profile{Name: "text"}

	gz, err := gazetteer.New(ctx, gazetteer.Config{
		Analyzer:       an,
		DefaultProfile: text,
		Sources: []gazetteer.Source{
			gazetteer.NewTSVSourceData("places", []byte(places)),
			gazetteer.NewTSVSourceData("people", []byte(people)),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	rf, err := rules.LoadNamed("example.yaml", strings.NewReader(ruleFile))
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

	for _, d := range []ner.Doc{
		{Text: "Иван Петров из деревни Лягушкино Боровского уезда"},
		{Text: "Иоанн из села Покровское"},
		{Text: "крестьянин Мороз"},
		{Text: "крестьянин Мороз", Tags: []string{"period:pre1917"}},
	} {
		res, err := p.Extract(ctx, d, ner.Explain())
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s [%s]\n", d.Text, strings.Join(d.Tags, ","))
		for _, s := range res.Spans {
			fmt.Printf("  %2d-%-2d %-10s %-18s normal=%s ref=%s %s\n", s.RuneStart, s.RuneEnd, s.Type, s.Surface,
				strings.Join(s.Normal, "|"), strings.Join(s.Refs, "|"), strings.Join(s.Flags.Names(), ","))
			for _, ev := range s.Evidence {
				fmt.Printf("        %s\n", ev)
			}
		}
	}

	// Output:
	// Иван Петров из деревни Лягушкино Боровского уезда []
	//    0-4  given_name Иван               normal=Иван ref=g1
	//         surface match «Иван» (people)
	//    5-11 surname    Петров             normal=Петров ref=s1
	//         surface match «Петров» (people)
	//   15-32 division   деревни Лягушкино  normal=Лягушкино ref=d1
	//         surface match «Лягушкино» (places)
	//         hint «деревни» → division +2
	//         trigger «деревни» → division +1
	//   33-49 division   Боровского уезда   normal=Боровский ref=d2
	//         lemma match «Боровский» (places)
	//         hint «уезда» → division +2
	//         trigger «деревни» → division +1
	// Иоанн из села Покровское []
	//    0-5  given_name Иоанн              normal=Иван ref=g1
	//         surface match «Иоанн» (people)
	//    9-24 division   села Покровское    normal=покровское ref= candidate
	//         trigger «села» → division candidate
	// крестьянин Мороз []
	// крестьянин Мороз [period:pre1917]
	//   11-16 surname    Мороз              normal=Мороз ref=s2
	//         surface match «Мороз» (people)
	//         hint «крестьянин» → surname +1
}
