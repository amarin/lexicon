package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/amarin/lexicon/basefetch"
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

	if !strings.Contains(errOut.String(), "base: no") || !strings.Contains(errOut.String(), "analyzer-2/modern-2/") {
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

// TestRunDictsFetchExisting: dicts fetch on a directory that already has the
// base dictionary must not touch the network and must leave the file
// byte-identical.
func TestRunDictsFetchExisting(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, basefetch.DefaultName)
	want := []byte("stub-existing-bytes")

	if err := os.WriteFile(dst, want, 0o644); err != nil {
		t.Fatal(err)
	}

	orig := fetch
	fetch = func(string) (string, error) {
		t.Fatal("fetch must not be called without --force when the file already exists")

		return "", nil
	}
	t.Cleanup(func() { fetch = orig })

	var out, errOut bytes.Buffer
	if code := run(t.Context(), []string{"dicts", "fetch", "--dicts", dir}, &out, &errOut); code != 0 {
		t.Fatalf("exit %d, stderr %s", code, errOut.String())
	}

	if !strings.Contains(out.String(), "already present") || !strings.Contains(out.String(), "--force") {
		t.Errorf("stdout %q lacks the exists/--force message", out.String())
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, want) {
		t.Errorf("file changed: got %q, want %q", got, want)
	}
}

// TestRunDictsFetchForce: --force calls the fetcher exactly once with the
// expected destination, even though a file is already there; no real
// download happens (the fetch seam is faked).
func TestRunDictsFetchForce(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, basefetch.DefaultName)

	if err := os.WriteFile(dst, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	var calls int

	var gotDst string

	orig := fetch
	fetch = func(d string) (string, error) {
		calls++
		gotDst = d

		if err := os.WriteFile(d, []byte("stub-fetched-bytes"), 0o644); err != nil {
			return "", err
		}

		return "1.0.0-test", nil
	}
	t.Cleanup(func() { fetch = orig })

	var out, errOut bytes.Buffer
	if code := run(t.Context(), []string{"dicts", "fetch", "--dicts", dir, "--force"}, &out, &errOut); code != 0 {
		t.Fatalf("exit %d, stderr %s", code, errOut.String())
	}

	if calls != 1 {
		t.Fatalf("fetch called %d times, want 1", calls)
	}

	if gotDst != dst {
		t.Errorf("fetch dst = %q, want %q", gotDst, dst)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != "stub-fetched-bytes" {
		t.Errorf("file = %q, want the fake fetcher's stub bytes", got)
	}

	if !strings.Contains(out.String(), "1.0.0-test") {
		t.Errorf("stdout %q lacks the fetched version", out.String())
	}
}

func TestRunUsage(t *testing.T) {
	cases := map[string][]string{
		"no command":                 nil,
		"unknown command":            {"nope"},
		"no text":                    {"analyze", "--dicts", t.TempDir()},
		"bad mode":                   {"analyze", "--mode", "x", "кот"},
		"bad flag":                   {"analyze", "--nope", "кот"},
		"no subcommand":              {"dicts"},
		"bad subcommand":             {"dicts", "enable"},
		"dicts fetch bad flag":       {"dicts", "fetch", "--dicts", t.TempDir(), "--nope"},
		"dicts list rejects --force": {"dicts", "list", "--dicts", t.TempDir(), "--force"},
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
