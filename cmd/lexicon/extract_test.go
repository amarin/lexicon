package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/internal/fakedict"
)

const (
	testTSV   = "../../nertest/testdata/genealogy.tsv"
	testRules = "../../nertest/testdata/genealogy.rules.yaml"
	testCases = "../../nertest/testdata/genealogy.jsonl"
)

func useFakeDictionaries(t *testing.T) {
	t.Helper()
	old := openDictionaries
	openDictionaries = func(context.Context, string) (lexicon.Dictionaries, []lexicon.Kind, func() error, error) {
		return fakedict.Genealogy(), fakedict.AllKinds(), func() error { return nil }, nil
	}
	t.Cleanup(func() { openDictionaries = old })
}

func runCmd(run func(context.Context, []string, io.Reader, io.Writer, io.Writer) int, stdin string, args ...string) (int, string, string) {
	var out, errb bytes.Buffer
	code := run(context.Background(), args, strings.NewReader(stdin), &out, &errb)
	return code, out.String(), errb.String()
}

func TestExtractJSONL(t *testing.T) {
	useFakeDictionaries(t)
	code, out, errs := runCmd(run, "",
		"extract", "--dicts", "unused", "--gazetteer", testTSV, "--rules", testRules, "--format", "jsonl", "--explain",
		"крестьянин деревни Лягушкиной Иван Петров")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errs)
	}
	var got []string
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		var s spanJSON
		if err := json.Unmarshal(sc.Bytes(), &s); err != nil {
			t.Fatalf("bad JSON line %q: %v", sc.Text(), err)
		}
		if s.Doc != 1 || len(s.Evidence) == 0 {
			t.Fatalf("span = %+v", s)
		}
		got = append(got, s.Type+":"+s.Surface)
	}
	want := []string{"estate:крестьянин", "division:деревни Лягушкиной", "given_name:Иван", "surname:Петров"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("got %q", got)
	}
}

func TestExtractTableFromStdin(t *testing.T) {
	useFakeDictionaries(t)
	code, out, errs := runCmd(run, "крестьянин Мороз\nжил на ул. Мира\n",
		"extract", "--dicts", "unused", "--gazetteer", testTSV, "--rules", testRules, "--tags", "period:pre1917", "-")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errs)
	}
	for _, want := range []string{"DOC", "surname", "Мороз", "street", "ул. Мира", "street:1"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output lacks %q:\n%s", want, out)
		}
	}
}

func TestExtractUsageErrors(t *testing.T) {
	useFakeDictionaries(t)
	if code, _, _ := runCmd(run, "", "extract", "--dicts", "x", "--gazetteer", testTSV); code != 2 {
		t.Fatalf("no text: exit %d", code)
	}
	if code, _, _ := runCmd(run, "", "extract", "--dicts", "x", "--ortho", "old", "Иван"); code != 2 {
		t.Fatalf("bad --ortho: exit %d", code)
	}
	if code, _, _ := runCmd(run, "", "extract", "--dicts", "x", "--rules", "missing.yaml", "Иван"); code != 1 {
		t.Fatalf("missing rules: exit %d", code)
	}
}
