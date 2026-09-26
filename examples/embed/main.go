// Command embed ships a host dictionary inside the binary: abbrev.records.tsv
// and its provenance sidecar are embedded with //go:embed and registered as a
// built-in (Options.Builtin), so the program needs no data files. A file of
// the same name in Options.Dir replaces the built-in — users can override
// shipped data without a rebuild. The base dictionary is embedded the same
// way through Options.Base and Options.BaseManifest (see
// docs/en/scenarios.md); it is not committed here.
//
// Run: go run ./examples/embed
package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/textnorm"
)

var (
	//go:embed abbrev.records.tsv
	abbrevData []byte
	//go:embed abbrev.records.tsv.meta
	abbrevMeta []byte
)

func main() {
	ctx := context.Background()

	meta, err := lexicon.ParseManifest(bytes.NewReader(abbrevMeta))
	if err != nil {
		log.Fatal(err)
	}
	builtin := []lexicon.BuiltinDict{{
		Name: "abbrev.records", Format: lexicon.FormatTSV, Data: abbrevData, Manifest: meta,
	}}

	run(ctx, "built-in only", lexicon.Options{Builtin: builtin})

	// A user directory with its own abbrev.records.tsv replaces the built-in.
	dir, err := os.MkdirTemp("", "lexicon-embed-")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()
	if err := os.WriteFile(filepath.Join(dir, "abbrev.records.tsv"), []byte("сельцо\tс.\tNOUN\n"), 0o644); err != nil {
		log.Fatal(err)
	}

	run(ctx, "overridden by a file", lexicon.Options{Dir: dir, Builtin: builtin})

	// Output:
	// built-in only: abbrev.records (builtin; lexicon examples; 1; license CC0): с. → село | сын
	// overridden by a file: abbrev.records (file abbrev.records.tsv; no sidecar): с. → сельцо
}

// run opens a registry, prints its abbrev.records entry and expands «с.».
func run(ctx context.Context, title string, o lexicon.Options) {
	reg, err := lexicon.Open(ctx, o)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = reg.Close() }()

	e := reg.List()[0]
	origin := e.Origin
	if origin != lexicon.OriginBuiltin {
		origin = "file " + filepath.Base(origin)
	}
	manifest := e.Manifest.String()
	if manifest == "" {
		manifest = "no sidecar"
	}

	a := lexicon.NewAnalyzer(reg, textnorm.Modern, lexicon.AnalyzerOptions{})
	var ls []string
	for _, t := range a.Analyze("с.", lexicon.Profile{Name: "text"}, lexicon.ModeIndex) {
		for _, l := range t.Lemmas {
			ls = append(ls, l.Text)
		}
	}

	fmt.Printf("%s: %s (%s; %s): с. → %s\n", title, e.Name, origin, manifest, strings.Join(ls, " | "))
}
