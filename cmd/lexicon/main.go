// Command lexicon analyzes text with lexicon dictionaries, extracts and
// scores entity spans, and manages the dictionary directory:
//
//	lexicon analyze [--dicts DIR] [--ortho modern|prereform] [--profile SPEC] [--mode index|full] TEXT...
//	lexicon extract --dicts DIR [--ortho modern|prereform] --gazetteer FILE... [--rules FILE...]
//	                [--nest outer>inner...] [--tags a,b] [--types a,b] [--explain] [--format table|jsonl] TEXT...|-
//	lexicon golden --cases FILE.jsonl --dicts DIR [--ortho modern|prereform] --gazetteer FILE... [--rules FILE...]
//	               [--nest outer>inner...] [--min-precision X] [--min-recall Y]
//	lexicon dicts list [--dicts DIR]
//	lexicon dicts fetch [--dicts DIR] [--force]
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

const usage = `usage:
  lexicon analyze [--dicts DIR] [--ortho modern|prereform] [--profile NAME[:KIND[G1|G2],KIND...]] [--mode index|full] TEXT...
  lexicon extract --dicts DIR [--ortho modern|prereform] --gazetteer FILE... [--rules FILE...]
                  [--nest outer>inner...] [--tags a,b] [--types a,b] [--explain] [--format table|jsonl] TEXT...|-
                  find entity spans (gazetteers + rules)
  lexicon golden --cases FILE.jsonl --dicts DIR [--ortho modern|prereform] --gazetteer FILE... [--rules FILE...]
                 [--nest outer>inner...] [--min-precision X] [--min-recall Y]
                 score extraction against golden JSONL cases
  lexicon dicts list [--dicts DIR]
  lexicon dicts fetch [--dicts DIR] [--force]
Flags go before the text. DIR defaults to $LEXICON_DICTS, $XDG_DATA_HOME/lexicon/dicts or ~/.local/share/lexicon/dicts.
`

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// run executes a command and returns the exit code: 0 ok, 1 failure, 2 usage.
func run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)

		return 2
	}

	var err error

	switch args[0] {
	case "analyze":
		err = runAnalyze(ctx, args[1:], stdout, stderr)
	case "extract":
		err = runExtract(ctx, args[1:], stdin, stdout, stderr)
	case "golden":
		err = runGolden(ctx, args[1:], stdin, stdout, stderr)
	case "dicts":
		err = runDicts(ctx, args[1:], stdout, stderr)
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)

		return 0
	default:
		err = usageError{fmt.Sprintf("unknown command %q", args[0])}
	}

	var ue usageError

	switch {
	case err == nil, errors.Is(err, flag.ErrHelp):
		return 0
	case errors.As(err, &ue):
		fmt.Fprintf(stderr, "lexicon: %v\n%s", err, usage)

		return 2
	default:
		fmt.Fprintf(stderr, "lexicon: %v\n", err)

		return 1
	}
}
