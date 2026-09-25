package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/amarin/lexicon"
)

// runAnalyze prints the terms of the text; the dictionary summary and the
// analyzer version go to stderr.
func runAnalyze(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("analyze", flag.ContinueOnError)
	fs.SetOutput(stderr)

	dir := fs.String("dicts", defaultDictsDir(), "dictionary directory")
	rulesName := fs.String("ortho", "modern", "orthography rules: modern, prereform")
	profileSpec := fs.String("profile", "text", "profile NAME[:KIND[G1|G2],KIND...]; no kinds = all dictionaries")
	modeName := fs.String("mode", "index", "analysis mode: index, full")

	if err := parseFlags(fs, args); err != nil {
		return err
	}

	rules, err := parseRules(*rulesName)
	if err != nil {
		return err
	}

	p, err := parseProfile(*profileSpec)
	if err != nil {
		return err
	}

	m, err := parseMode(*modeName)
	if err != nil {
		return err
	}

	text := strings.Join(fs.Args(), " ")
	if text == "" {
		return usageError{"analyze: no text"}
	}

	// Kinds nil: the CLI has no host vocabulary and accepts every valid kind
	// (owner decision 7, D17).
	reg, err := lexicon.Open(ctx, lexicon.Options{Dir: *dir, Kinds: nil})
	if err != nil {
		return err
	}
	defer reg.Close()

	a := lexicon.NewAnalyzer(reg, rules, lexicon.AnalyzerOptions{})
	formatTerms(stdout, a.Analyze(text, p, m))
	fmt.Fprintf(stderr, "%s; version %s\n", reg.Summary(), a.Version())

	return nil
}
