package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/amarin/lexicon/ner"
)

// runExtract implements `lexicon extract`: it finds entity spans (gazetteers
// and rules) in free text. A pipeline/registry Close error is joined into
// the result so it is never silently dropped.
func runExtract(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) (err error) {
	fs := flag.NewFlagSet("extract", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var pf pipelineFlags
	pf.register(fs)

	tags := fs.String("tags", "", "comma-separated document tags")
	types := fs.String("types", "", "comma-separated output types")
	explain := fs.Bool("explain", false, "print evidence")
	format := fs.String("format", "table", "output format: table, jsonl")

	if err := parseFlags(fs, args); err != nil {
		return err
	}

	texts := fs.Args()
	if len(texts) == 0 {
		return usageError{"extract: no text (pass TEXT... or - for stdin, one document per line)"}
	}

	if *format != "table" && *format != "jsonl" {
		return usageError{fmt.Sprintf("extract: unknown --format %q", *format)}
	}

	if len(texts) == 1 && texts[0] == "-" {
		texts = nil

		sc := bufio.NewScanner(stdin)
		sc.Buffer(make([]byte, 64*1024), 16*1024*1024)

		for sc.Scan() {
			// The line is the document exactly as read: span offsets refer
			// to it. Blank lines are skipped, nothing is trimmed.
			if line := sc.Text(); strings.TrimSpace(line) != "" {
				texts = append(texts, line)
			}
		}

		if err := sc.Err(); err != nil {
			return fmt.Errorf("extract: %w", err)
		}
	}

	p, closeFn, err := pf.build(ctx, stderr)
	if err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	defer func() {
		if cerr := closeFn(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("extract: close: %v", cerr))
		}
	}()

	var opts []ner.Option
	if *explain {
		opts = append(opts, ner.Explain())
	}

	tw := tabwriter.NewWriter(stdout, 0, 4, 2, ' ', 0)
	if *format == "table" {
		fmt.Fprintln(tw, "DOC\tSTART\tEND\tTYPE\tSURFACE\tNORMAL\tREFS\tFLAGS\tSCORE")
	}

	enc := json.NewEncoder(stdout)

	for i, text := range texts {
		res, err := p.Extract(ctx, ner.Doc{Text: text, Tags: splitList(*tags), Types: splitList(*types)}, opts...)
		if err != nil {
			return fmt.Errorf("extract: document %d: %w", i+1, err)
		}

		for _, s := range res.Spans {
			if *format == "jsonl" {
				if err := enc.Encode(newSpanJSON(i+1, s)); err != nil {
					return fmt.Errorf("extract: %w", err)
				}

				continue
			}

			fmt.Fprintf(tw, "%d\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t%.2f\n", i+1, s.Start, s.End, s.Type, s.Surface,
				strings.Join(s.Normal, "|"), strings.Join(s.Refs, "|"), strings.Join(s.Flags.Names(), ","), s.Score)

			for _, ev := range s.Evidence {
				fmt.Fprintf(tw, "\t\t\t\t  %s\n", ev)
			}

			for _, a := range s.Alternatives {
				fmt.Fprintf(tw, "\t\t\t\t  alt %s %s %s %.2f\n", a.Type,
					strings.Join(a.Normal, "|"), strings.Join(a.Refs, "|"), a.Score)
			}
		}
	}

	if err := tw.Flush(); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	return nil
}
