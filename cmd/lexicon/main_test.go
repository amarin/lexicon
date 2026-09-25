package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// givenDir is a dictionary directory with one TSV given-name dictionary.
func givenDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	data := "иван\tивана\tNOUN,anim,masc,Name sing,gent\nиван\tиван\tNOUN,anim,masc,Name sing,nomn\n"

	if err := os.WriteFile(filepath.Join(dir, "given.test.tsv"), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	return dir
}

func TestRunAnalyzeFull(t *testing.T) {
	var out, errOut bytes.Buffer

	code := run(t.Context(), []string{"analyze", "--dicts", givenDir(t), "--mode", "full", "Ивана, сына"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("exit %d, stderr %s", code, errOut.String())
	}

	if want := "Ивана\tивана\tиван\n,\t\t\nсына\tсына\tсына[unknown]\n"; out.String() != want {
		t.Fatalf("stdout =\n%q\nwant\n%q", out.String(), want)
	}

	if !strings.Contains(errOut.String(), "base: no") || !strings.Contains(errOut.String(), "analyzer-1/modern-2/") {
		t.Fatalf("stderr %q", errOut.String())
	}
}

func TestRunDictsList(t *testing.T) {
	var out, errOut bytes.Buffer

	if code := run(t.Context(), []string{"dicts", "list", "--dicts", givenDir(t)}, &out, &errOut); code != 0 {
		t.Fatalf("exit %d, stderr %s", code, errOut.String())
	}

	for _, want := range []string{"given.test", "tsv", "dictionaries: 1 of 1 enabled", "version: "} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("stdout lacks %q:\n%s", want, out.String())
		}
	}
}

func TestRunUsage(t *testing.T) {
	cases := map[string][]string{
		"no command":      nil,
		"unknown command": {"nope"},
		"no text":         {"analyze", "--dicts", t.TempDir()},
		"bad mode":        {"analyze", "--mode", "x", "кот"},
		"bad flag":        {"analyze", "--nope", "кот"},
		"no subcommand":   {"dicts"},
		"bad subcommand":  {"dicts", "enable"},
	}
	for name, args := range cases {
		var out, errOut bytes.Buffer
		if code := run(t.Context(), args, &out, &errOut); code != 2 {
			t.Errorf("%s: exit %d, want 2", name, code)
		}
	}

	var out bytes.Buffer
	if code := run(t.Context(), []string{"help"}, &out, &out); code != 0 || !strings.Contains(out.String(), "usage:") {
		t.Errorf("help: exit %d, %q", code, out.String())
	}
}
