package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/amarin/lexicon/nertest"
)

// runGolden implements `lexicon golden`: it scores extraction against
// golden JSONL cases. Case.Context is ignored (nertest.Run is called
// without options); a check below the thresholds is a plain error (exit 1).
// A pipeline/registry Close error is joined into the result so it is never
// silently dropped. Only the "text" profile exists (see pipelineFlags.build):
// a case with "profile": "name" aborts the run with ner.ErrUnknownProfile.
func runGolden(ctx context.Context, args []string, _ io.Reader, stdout, stderr io.Writer) (err error) {
	fs := flag.NewFlagSet("golden", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var pf pipelineFlags
	pf.register(fs)

	casesPath := fs.String("cases", "", "golden cases, JSON lines (required)")
	minP := fs.Float64("min-precision", 0, "minimum strict precision per type")
	minR := fs.Float64("min-recall", 0, "minimum strict recall per type")

	if err := parseFlags(fs, args); err != nil {
		return err
	}

	if *casesPath == "" {
		return usageError{"golden: --cases is required"}
	}

	cases, err := nertest.LoadCasesFile(*casesPath)
	if err != nil {
		return fmt.Errorf("golden: %w", err)
	}

	p, closeFn, err := pf.build(ctx, stderr)
	if err != nil {
		return fmt.Errorf("golden: %w", err)
	}

	defer func() {
		if cerr := closeFn(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("golden: close: %v", cerr))
		}
	}()

	rep, err := nertest.Run(ctx, p, cases) // no WithTags: Case.Context is ignored
	if err != nil {
		return fmt.Errorf("golden: %w", err)
	}

	if err := rep.Write(stdout); err != nil {
		return fmt.Errorf("golden: %w", err)
	}

	if v := rep.Check(*minP, *minR); len(v) > 0 {
		for _, line := range v {
			fmt.Fprintf(stderr, "golden: %s\n", line)
		}

		return fmt.Errorf("golden: below threshold")
	}

	return nil
}
