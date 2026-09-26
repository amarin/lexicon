// Command gazetteer compiles host records into a gazetteer and keeps it
// current: a TSV source with a bad line (listed in the build report, not
// fatal), a host source whose records change at run time (Refresh
// recompiles only the sources whose Version changed and swaps a new
// snapshot in atomically), raw matches of a snapshot, and the variant
// groups a search box can expand a query with. The morphology is a tiny
// TSV registered in code, so the example needs no download; «Сидорова» is
// not in it, and its lemma «сидоров» is predicted from the ending.
//
// Run: go run ./examples/gazetteer
package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/gazetteer"
	"github.com/amarin/lexicon/textnorm"
)

// base is a stand-in for the OpenCorpora base: lemma<TAB>wordform<TAB>tag.
const base = `иван	иван	NOUN,anim,masc,Name sing,nomn
иван	ивана	NOUN,anim,masc,Name sing,gent
иоанн	иоанн	NOUN,anim,masc,Name sing,nomn
иоанн	иоанна	NOUN,anim,masc,Name sing,gent
сын	сын	NOUN,anim,masc sing,nomn
`

// names is a TSV source: type<TAB>ref<TAB>canonical<TAB>alias[<TAB>flags].
// «Иван» and «Иоанн» share the ref g1: one record, two variants.
const names = `# source: lexicon example
# license: CC0-1.0
given_name	g1	Иван	Иван
given_name	g1	Иван	Иоанн
given_name	g2
`

// surnames is a host source over live data: its Version changes whenever
// the records do, so Refresh knows to recompile it.
type surnames struct {
	rev     int
	entries []gazetteer.Entry
}

func (s *surnames) Name() string { return "surnames" }

func (s *surnames) Version(context.Context) (string, error) { return strconv.Itoa(s.rev), nil }

func (s *surnames) Entries(_ context.Context, yield func(gazetteer.Entry) error) error {
	for _, e := range s.entries {
		if err := yield(e); err != nil {
			return err
		}
	}
	return nil
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

	an := lexicon.NewAnalyzer(reg, textnorm.Modern, lexicon.AnalyzerOptions{})
	text := lexicon.Profile{Name: "text"}
	host := &surnames{rev: 1, entries: []gazetteer.Entry{{Alias: "Петров", Type: "surname", Ref: "s1"}}}

	gz, err := gazetteer.New(ctx, gazetteer.Config{
		Analyzer:       an,
		DefaultProfile: text,
		Sources:        []gazetteer.Source{gazetteer.NewTSVSourceData("names", []byte(names)), host},
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, r := range gz.Snapshot().Reports() {
		fmt.Printf("%s: %d aliases, errors %q\n", r.Source, r.Aliases, r.Errors)
	}

	doc := "Иоанна Сидорова сын Иван Петров"
	show := func() {
		terms := an.Analyze(doc, text, lexicon.ModeFull)
		tx := gazetteer.Prepare(terms)
		for _, m := range gz.Snapshot().Match(tx, nil) {
			from := terms[tx.TermIndex(m.Start)].Token
			to := terms[tx.TermIndex(m.End-1)].Token
			for _, a := range m.Aliases {
				fmt.Printf("  %-8s %-10s %s %s\n", doc[from.Start:to.End], a.Entry.Type, a.Entry.Ref, kind(m.Kind))
			}
		}
	}
	fmt.Println("matches:")
	show()

	// The host adds a record; only its source is recompiled. Readers that
	// hold the old snapshot keep a consistent view until they are done.
	host.rev++
	host.entries = append(host.entries, gazetteer.Entry{Alias: "Сидоров", Type: "surname", Ref: "s2"})
	reps, err := gz.Refresh(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range reps {
		fmt.Printf("refreshed %s: version %s, %d aliases\n", r.Source, r.Version, r.Aliases)
	}
	fmt.Println("matches:")
	show()

	// Variant groups: a query for «Иван» can also search «Иоанн».
	fmt.Println("expand иван:", strings.Join(gz.Expand("иван"), ", "))
	fmt.Println("canonical иоанн:", strings.Join(gz.Canonical("иоанн"), ", "))

	// Output:
	// names: 2 aliases, errors ["line 5: want 4 to 6 tab-separated fields, got 2"]
	// surnames: 1 aliases, errors []
	// matches:
	//   Иоанна   given_name g1 by lemma
	//   Иван     given_name g1 by lemma
	//   Иван     given_name g1 by surface
	//   Петров   surname    s1 by lemma
	//   Петров   surname    s1 by surface
	// refreshed surnames: version 2, 2 aliases
	// matches:
	//   Иоанна   given_name g1 by lemma
	//   Иван     given_name g1 by lemma
	//   Иван     given_name g1 by surface
	//   Сидорова surname    s2 by lemma
	//   Петров   surname    s1 by lemma
	//   Петров   surname    s1 by surface
	// expand иван: иван, иоанн
	// canonical иоанн: Иван
}

// kind names how a match was found.
func kind(k gazetteer.MatchKind) string {
	if k == gazetteer.ByLemma {
		return "by lemma"
	}
	return "by surface"
}
