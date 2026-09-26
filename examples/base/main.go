// Command base analyzes text with the real OpenCorpora base dictionary. It
// needs the dictionary on disk first:
//
//	go run ./cmd/lexicon dicts fetch            # ~15 MB download, CC BY-SA
//	go run ./examples/base -dicts ~/.local/share/lexicon/dicts
//
// It prints the dictionary provenance (hosts that distribute the base must
// attribute it), then the ModeIndex terms of a pre-reform sentence under a
// "text" and a "name" profile.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/textnorm"
)

func main() {
	home, _ := os.UserHomeDir()
	dir := flag.String("dicts", filepath.Join(home, ".local", "share", "lexicon", "dicts"), "dictionary directory with base.opencorpora.dat")
	flag.Parse()

	reg, err := lexicon.Open(context.Background(), lexicon.Options{Dir: *dir})
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = reg.Close() }()

	for _, e := range reg.List() {
		fmt.Printf("%s: %s%s\n", e.Name, e.Manifest, e.Error)
	}
	fmt.Println(reg.Summary())

	a := lexicon.NewAnalyzer(reg, textnorm.PreReform, lexicon.AnalyzerOptions{})
	text := "Вѣра Морозъ изъ села Покровскаго, 1834 г."
	profiles := []lexicon.Profile{
		{Name: "text"},
		{Name: "name", Kinds: []lexicon.Kind{lexicon.KindBase}, Grammemes: map[lexicon.Kind][]string{lexicon.KindBase: {"Name", "Surn", "Patr"}}},
	}

	for _, p := range profiles {
		fmt.Printf("\nprofile %q\n", p.Name)
		for _, t := range a.Analyze(text, p, lexicon.ModeIndex) {
			fmt.Printf("  %-12s %s\n", t.Token.Raw, lemmas(t))
		}
	}
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

	return strings.Join(out, " ")
}
