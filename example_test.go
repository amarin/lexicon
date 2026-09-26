package lexicon_test

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/textnorm"
)

// exampleBase stands in for the OpenCorpora base: lemma<TAB>wordform<TAB>tag.
const exampleBase = `кот	кот	NOUN,anim,masc sing,nomn
кот	кота	NOUN,anim,masc sing,gent
вера	вера	NOUN,inan,femn sing,nomn
вера	веры	NOUN,inan,femn sing,gent
вера	веры	NOUN,anim,femn,Name sing,gent
в	в	PREP
покровский	покровского	ADJF,Geox masc,sing,gent
`

// exampleRegistry opens a registry over exampleBase and an abbreviation
// dictionary, both built-ins, so the examples need no files.
func exampleRegistry() *lexicon.Registry {
	reg, err := lexicon.Open(context.Background(), lexicon.Options{
		Builtin: []lexicon.BuiltinDict{
			{Name: "base.example", Format: lexicon.FormatTSV, Data: []byte(exampleBase)},
			{Name: "abbrev.example", Format: lexicon.FormatTSV, Data: []byte("село\tс.\tNOUN\nсын\tс.\tNOUN\n")},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	return reg
}

// show prints form=lemma[flags]|… for each term.
func show(ts []lexicon.Term) string {
	var out []string
	for _, t := range ts {
		var ls []string
		for _, l := range t.Lemmas {
			s := l.Text
			if l.Flags != 0 {
				s += "[" + l.Flags.String() + "]"
			}
			ls = append(ls, s)
		}
		out = append(out, t.Form+"="+strings.Join(ls, "|"))
	}

	return strings.Join(out, " ")
}

// Open a registry from host built-ins; a real host adds Options.Dir with the
// base dictionary fetched by basefetch (or embeds it in Options.Base).
func ExampleOpen() {
	reg := exampleRegistry()
	defer func() { _ = reg.Close() }()

	for _, e := range reg.List() {
		fmt.Println(e.Name, e.Kind, e.Format, e.Origin, e.Enabled)
	}
	fmt.Println(reg.Summary())
	// Output:
	// base.example base tsv builtin true
	// abbrev.example abbrev tsv builtin true
	// dictionaries: 2 of 2 enabled; base: yes; broken: 0
}

// ModeIndex gives search-index terms: service words dropped, abbreviations
// expanded, pre-reform endings found through modern variants.
func ExampleAnalyzer_Analyze() {
	reg := exampleRegistry()
	defer func() { _ = reg.Close() }()

	a := lexicon.NewAnalyzer(reg, textnorm.PreReform, lexicon.AnalyzerOptions{})
	fmt.Println(show(a.Analyze("Кота въ с. Покровскаго", lexicon.Profile{Name: "text"}, lexicon.ModeIndex)))
	// Output: кота=кот с.=село[ambiguous,abbrev]|сын[ambiguous,abbrev] покровскаго=покровский[reform]
}

// ModeFull gives exactly one term per token, punctuation and service words
// included, for NER and markup.
func ExampleAnalyzer_Analyze_full() {
	reg := exampleRegistry()
	defer func() { _ = reg.Close() }()

	a := lexicon.NewAnalyzer(reg, textnorm.Modern, lexicon.AnalyzerOptions{})
	for _, t := range a.Analyze("Кота в дом!", lexicon.Profile{Name: "text"}, lexicon.ModeFull) {
		fmt.Printf("%d-%d %q %s\n", t.Token.RuneStart, t.Token.RuneEnd, t.Token.Raw, show([]lexicon.Term{t}))
	}
	// Output:
	// 0-4 "Кота" кота=кот
	// 5-6 "в" в=в[stop]
	// 7-10 "дом" дом=дом[unknown]
	// 10-11 "!" =
}

// A Profile restricts the dictionaries and grammemes of a field: in a name
// field «Веры» is only the given name.
func ExampleProfile() {
	reg := exampleRegistry()
	defer func() { _ = reg.Close() }()

	a := lexicon.NewAnalyzer(reg, textnorm.Modern, lexicon.AnalyzerOptions{})
	name := lexicon.Profile{
		Name:      "name",
		Kinds:     []lexicon.Kind{lexicon.KindBase},
		Grammemes: map[lexicon.Kind][]string{lexicon.KindBase: {"Name", "Surn", "Patr"}},
	}
	for _, p := range []lexicon.Profile{{Name: "text"}, name} {
		t := a.Analyze("Веры", p, lexicon.ModeIndex)[0]
		fmt.Println(p.Name, t.Lemmas[0].Text, t.Lemmas[0].Tag)
	}
	// Output:
	// text вера NOUN,inan,femn sing,gent
	// name вера NOUN,anim,femn,Name sing,gent
}

// The last word of a query, while still being typed, is Partial: search it
// by prefix and by its lemmas.
func ExampleAnalyzer_ParseQuery() {
	reg := exampleRegistry()
	defer func() { _ = reg.Close() }()

	a := lexicon.NewAnalyzer(reg, textnorm.Modern, lexicon.AnalyzerOptions{})
	q := a.ParseQuery("кота Покр", lexicon.Profile{Name: "text"})
	fmt.Println(show(q.Terms), "| partial:", show([]lexicon.Term{*q.Partial}))
	// Output: кота=кот | partial: покр=покр[unknown]
}

// Analyzer.Version changes when the dictionary set changes: store it with
// derived data (a search index) and rebuild on change.
func ExampleAnalyzer_Version() {
	reg := exampleRegistry()
	defer func() { _ = reg.Close() }()

	a := lexicon.NewAnalyzer(reg, textnorm.Modern, lexicon.AnalyzerOptions{})
	before := a.Version()
	if err := reg.SetEnabled(context.Background(), "abbrev.example", false); err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings.HasPrefix(before, "analyzer-2/modern-3/"), a.Version() != before)
	// Output: true true
}

// A provenance sidecar is "key: value" lines; hosts embed the one basefetch
// writes and pass it as Options.BaseManifest.
func ExampleParseManifest() {
	m, err := lexicon.ParseManifest(strings.NewReader("source: OpenCorpora\nlicense: CC BY-SA 4.0\nversion: 2.4\n"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(m)
	// Output: OpenCorpora; 2.4; license CC BY-SA 4.0
}
