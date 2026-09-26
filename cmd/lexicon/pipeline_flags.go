package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strings"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/gazetteer"
	"github.com/amarin/lexicon/ner"
	"github.com/amarin/lexicon/rules"
)

// openDictionaries opens the morphology registry and lists its enabled
// kinds. Tests replace it with the fakedict fixture.
var openDictionaries = func(ctx context.Context, dir string) (lexicon.Dictionaries, []lexicon.Kind, func() error, error) {
	reg, err := lexicon.Open(ctx, lexicon.Options{Dir: dir})
	if err != nil {
		return nil, nil, nil, err
	}

	var kinds []lexicon.Kind
	for _, e := range reg.List() {
		if e.Enabled && !slices.Contains(kinds, e.Kind) {
			kinds = append(kinds, e.Kind)
		}
	}

	return reg, kinds, reg.Close, nil
}

// pipelineFlags are the flags shared by extract and golden.
type pipelineFlags struct {
	dicts      string
	ortho      string
	gazetteers stringList
	rules      stringList
	nesting    stringList
}

func (f *pipelineFlags) register(fs *flag.FlagSet) {
	fs.StringVar(&f.dicts, "dicts", defaultDictsDir(), "morphology dictionary directory")
	fs.StringVar(&f.ortho, "ortho", "modern", "orthography rules: modern, prereform")
	fs.Var(&f.gazetteers, "gazetteer", "gazetteer TSV file (repeatable)")
	fs.Var(&f.rules, "rules", "rule file, YAML (.yaml/.yml) or JSON (repeatable)")
	fs.Var(&f.nesting, "nest", "allowed nesting outer>inner (repeatable)")
}

// build assembles a pipeline; non-fatal gazetteer problems go to warn.
func (f *pipelineFlags) build(ctx context.Context, warn io.Writer) (*ner.Pipeline, func() error, error) {
	ortho, err := parseRules(f.ortho)
	if err != nil {
		return nil, nil, err
	}

	var files []rules.File

	for _, path := range f.rules {
		rf, err := rules.LoadFile(path)
		if err != nil {
			return nil, nil, err
		}

		files = append(files, rf)
	}

	book, err := rules.Compile(files...)
	if err != nil {
		return nil, nil, err
	}

	nesting := map[string][]string{}
	for _, n := range f.nesting {
		outer, inner, ok := strings.Cut(n, ">")
		if !ok || outer == "" || inner == "" {
			return nil, nil, usageError{fmt.Sprintf("bad --nest %q, want outer>inner", n)}
		}

		nesting[outer] = append(nesting[outer], inner)
	}

	dicts, kinds, closeFn, err := openDictionaries(ctx, f.dicts)
	if err != nil {
		return nil, nil, err
	}

	fail := func(err error) (*ner.Pipeline, func() error, error) {
		_ = closeFn()
		return nil, nil, err
	}

	an := lexicon.NewAnalyzer(dicts, ortho, lexicon.AnalyzerOptions{})
	text := lexicon.Profile{Name: "text", Kinds: kinds}

	var srcs []gazetteer.Source
	for _, path := range f.gazetteers {
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		srcs = append(srcs, gazetteer.NewTSVSource(name, path))
	}

	gz, err := gazetteer.New(ctx, gazetteer.Config{Analyzer: an, DefaultProfile: text, Sources: srcs})
	if err != nil {
		return fail(err)
	}

	for _, rep := range gz.Snapshot().Reports() {
		if rep.Err != nil {
			return fail(fmt.Errorf("gazetteer %s: %w", rep.Source, rep.Err))
		}

		for _, e := range rep.Errors {
			fmt.Fprintf(warn, "warning: gazetteer %s: %s\n", rep.Source, e)
		}
	}

	p, err := ner.New(ner.Config{
		Analyzer: an, Gazetteer: gz, Rules: book,
		Profiles: map[string]lexicon.Profile{"text": text}, DefaultProfile: "text",
		Nesting: nesting,
	})
	if err != nil {
		return fail(err)
	}

	return p, closeFn, nil
}

// splitList splits a comma-separated flag value, dropping empty parts.
func splitList(s string) []string {
	var out []string

	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}

	return out
}
