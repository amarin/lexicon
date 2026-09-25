package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/textnorm"
)

// parseFlags parses fs; parse errors become usage errors, -h stays flag.ErrHelp.
func parseFlags(fs *flag.FlagSet, args []string) error {
	err := fs.Parse(args)
	if err == nil || errors.Is(err, flag.ErrHelp) {
		return err
	}

	return usageError{err.Error()}
}

// parseRules maps --ortho to a rule set.
func parseRules(s string) (textnorm.Rules, error) {
	switch s {
	case "modern":
		return textnorm.Modern, nil
	case "prereform":
		return textnorm.PreReform, nil
	}

	return textnorm.Rules{}, usageError{fmt.Sprintf("unknown rules %q (modern, prereform)", s)}
}

// parseMode maps --mode to an analysis mode.
func parseMode(s string) (lexicon.Mode, error) {
	switch s {
	case "index":
		return lexicon.ModeIndex, nil
	case "full":
		return lexicon.ModeFull, nil
	}

	return 0, usageError{fmt.Sprintf("unknown mode %q (index, full)", s)}
}

// parseProfile reads NAME[:KIND[G1|G2],KIND...]; no kinds = all dictionaries.
func parseProfile(s string) (lexicon.Profile, error) {
	name, kinds, _ := strings.Cut(s, ":")
	if name == "" {
		return lexicon.Profile{}, usageError{fmt.Sprintf("profile %q: empty name", s)}
	}

	p := lexicon.Profile{Name: name}
	if kinds == "" {
		return p, nil
	}

	for _, item := range strings.Split(kinds, ",") {
		k, gs, hasGrammemes := strings.Cut(item, "[")

		kind := lexicon.Kind(k)
		if !kind.Valid() {
			return lexicon.Profile{}, usageError{fmt.Sprintf("profile %q: invalid kind %q", s, k)}
		}

		p.Kinds = append(p.Kinds, kind)

		if !hasGrammemes {
			continue
		}

		gs, closed := strings.CutSuffix(gs, "]")
		if !closed || gs == "" {
			return lexicon.Profile{}, usageError{fmt.Sprintf("profile %q: want %s[G1|G2]", s, k)}
		}

		if p.Grammemes == nil {
			p.Grammemes = map[lexicon.Kind][]string{}
		}

		p.Grammemes[kind] = strings.Split(gs, "|")
	}

	return p, nil
}

// defaultDictsDir: $LEXICON_DICTS, else $XDG_DATA_HOME/lexicon/dicts, else
// ~/.local/share/lexicon/dicts.
func defaultDictsDir() string {
	if d := os.Getenv("LEXICON_DICTS"); d != "" {
		return d
	}

	if d := os.Getenv("XDG_DATA_HOME"); d != "" {
		return filepath.Join(d, "lexicon", "dicts")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "dicts"
	}

	return filepath.Join(home, ".local", "share", "lexicon", "dicts")
}
