// Command registry manages a dictionary directory the way a host does:
// <kind>.<name>.tsv|dat files with optional .meta provenance sidecars, a
// broken or undeclared file listed with its error instead of failing Open,
// a dictionary disabled and re-enabled through the StateStore, and a file
// replaced atomically and picked up by Reload. Analyzer.Version changes with
// every change of the enabled set, telling the host to reindex.
//
// Run: go run ./examples/registry
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/textnorm"
)

func main() {
	ctx := context.Background()

	dir, err := os.MkdirTemp("", "lexicon-registry-")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	write(dir, "base.demo.tsv", "кот\tкот\tNOUN,anim,masc sing,nomn\nкот\tкота\tNOUN,anim,masc sing,gent\n")
	write(dir, "surname.parish.tsv", "мороз\tмороз\tNOUN,anim,masc,Surn sing,nomn\nмороз\tмороза\tNOUN,anim,masc,Surn sing,gent\n")
	write(dir, "surname.parish.tsv.meta", "source: parish registers 1834\nlicense: CC0\n")
	write(dir, "surnme.typo.tsv", "x\tx\tX\n") // a typo in the kind: listed, not used

	reg, err := lexicon.Open(ctx, lexicon.Options{Dir: dir, Kinds: []lexicon.Kind{"surname"}})
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = reg.Close() }()

	list(reg)

	a := lexicon.NewAnalyzer(reg, textnorm.Modern, lexicon.AnalyzerOptions{})
	name := lexicon.Profile{Name: "name", Kinds: []lexicon.Kind{"surname"}}
	v := a.Version()
	show := func(step string) {
		changed := a.Version() != v
		v = a.Version()
		fmt.Printf("%-28s Мороза → %-18s version changed: %t\n", step, lemmas(a.Analyze("Мороза", name, lexicon.ModeIndex)), changed)
	}

	show("opened")

	if err := reg.SetEnabled(ctx, "surname.parish", false); err != nil {
		log.Fatal(err)
	}
	show("surname.parish disabled")

	if err := reg.SetEnabled(ctx, "surname.parish", true); err != nil {
		log.Fatal(err)
	}
	show("surname.parish enabled")

	// Replace the file atomically (temporary file + rename), then Reload.
	write(dir, "surname.parish.tsv", "морозов\tмороза\tNOUN,anim,masc,Surn sing,gent\n")
	if err := reg.Reload(ctx); err != nil {
		log.Fatal(err)
	}
	show("file replaced, Reload")

	fmt.Println(reg.Summary())

	// Output:
	// base.demo       base     tsv  -
	// surname.parish  surname  tsv  parish registers 1834; license CC0
	// surnme.typo     surnme   tsv  error: dictionary "surnme.typo": unknown kind "surnme"
	//
	// opened                       Мороза → мороз              version changed: false
	// surname.parish disabled      Мороза → мороза[unknown]    version changed: true
	// surname.parish enabled       Мороза → мороз              version changed: true
	// file replaced, Reload        Мороза → морозов            version changed: true
	// dictionaries: 2 of 3 enabled; base: yes; broken: 1
}

// write creates dir/name atomically: a temporary file renamed over the target.
func write(dir, name, data string) {
	tmp := filepath.Join(dir, name+".tmp")
	if err := os.WriteFile(tmp, []byte(data), 0o644); err != nil {
		log.Fatal(err)
	}
	if err := os.Rename(tmp, filepath.Join(dir, name)); err != nil {
		log.Fatal(err)
	}
}

// list prints the registry entries: name, kind, format, provenance or error.
func list(reg *lexicon.Registry) {
	for _, e := range reg.List() {
		state := e.Manifest.String()
		if state == "" {
			state = "-"
		}
		if e.Error != "" {
			state = "error: " + strings.TrimPrefix(e.Error, "lexicon: ")
		}
		fmt.Printf("%-15s %-8s %-4s %s\n", e.Name, e.Kind, e.Format, state)
	}
	fmt.Println()
}

// lemmas joins the lemma texts of all terms with their flags.
func lemmas(ts []lexicon.Term) string {
	var out []string
	for _, t := range ts {
		for _, l := range t.Lemmas {
			s := l.Text
			if l.Flags != 0 {
				s += "[" + l.Flags.String() + "]"
			}
			out = append(out, s)
		}
	}

	return strings.Join(out, " ")
}
