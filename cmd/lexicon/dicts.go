package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/basefetch"
)

// fetch downloads the base dictionary; a seam so tests can fake it without
// touching the network.
var fetch = basefetch.Fetch

// runDicts: dicts list | dicts fetch. A Close error from the registry (list
// path) is joined into the result so it is never silently dropped.
func runDicts(ctx context.Context, args []string, stdout, stderr io.Writer) (err error) {
	if len(args) == 0 {
		return usageError{"dicts: expected list or fetch"}
	}

	sub := args[0]
	if sub != "list" && sub != "fetch" {
		return usageError{fmt.Sprintf("dicts: unknown subcommand %q", sub)}
	}

	fs := flag.NewFlagSet("dicts "+sub, flag.ContinueOnError)
	fs.SetOutput(stderr)

	dir := fs.String("dicts", defaultDictsDir(), "dictionary directory")

	var force *bool
	if sub == "fetch" {
		force = fs.Bool("force", false, "download again even if the file exists")
	}

	if err := parseFlags(fs, args[1:]); err != nil {
		return err
	}

	if sub == "fetch" {
		return fetchBase(stdout, *dir, *force)
	}

	// Kinds nil: the CLI has no host vocabulary and accepts every valid kind
	// (owner decision 7, D17).
	reg, err := lexicon.Open(ctx, lexicon.Options{Dir: *dir, Kinds: nil})
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, reg.Close())
	}()

	return printDicts(stdout, reg)
}

// fetchBase downloads the base dictionary into dir unless it is there.
func fetchBase(w io.Writer, dir string, force bool) error {
	dst := filepath.Join(dir, basefetch.DefaultName)
	if _, err := os.Stat(dst); err == nil && !force {
		fmt.Fprintf(w, "base dictionary already present: %s (use --force to download again)\n", dst)

		return nil
	}

	version, err := fetch(dst)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "base dictionary pymorphy2-dicts-ru %s -> %s\n", version, dst)

	return nil
}

// printDicts writes the dictionary table, the summary and the set version.
func printDicts(w io.Writer, reg *lexicon.Registry) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tKIND\tFORMAT\tENABLED\tORIGIN\tHASH\tSOURCE\tERROR")

	for _, e := range reg.List() {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%t\t%s\t%s\t%s\t%s\n",
			e.Name, e.Kind, e.Format, e.Enabled, e.Origin, shortHash(e.Hash), e.Manifest, e.Error)
	}

	if err := tw.Flush(); err != nil {
		return err
	}

	_, err := fmt.Fprintf(w, "%s\nversion: %s\n", reg.Summary(), reg.Version())

	return err
}

// shortHash is the first 12 hex digits of a hash.
func shortHash(h string) string {
	if len(h) > 12 {
		return h[:12]
	}

	return h
}
