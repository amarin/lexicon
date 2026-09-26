package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoldenPasses(t *testing.T) {
	useFakeDictionaries(t)
	code, out, errs := runCmd(run, "",
		"golden", "--cases", testCases, "--dicts", "unused", "--gazetteer", testTSV, "--rules", testRules,
		"--min-precision", "1", "--min-recall", "1")
	if code != 0 {
		t.Fatalf("exit %d: %s\n%s", code, errs, out)
	}
	if !strings.Contains(out, "division") || !strings.Contains(out, "cases: 6") {
		t.Fatalf("report:\n%s", out)
	}
}

func TestGoldenBelowThreshold(t *testing.T) {
	useFakeDictionaries(t)
	cases := filepath.Join(t.TempDir(), "cases.jsonl")
	line := `{"id": "missing", "text": "видел Уса", "spans": [{"text": "Уса", "type": "surname"}]}` + "\n"
	if err := os.WriteFile(cases, []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, errs := runCmd(run, "",
		"golden", "--cases", cases, "--dicts", "unused", "--gazetteer", testTSV, "--min-recall", "1")
	if code != 1 || !strings.Contains(errs, "surname: strict recall") || !strings.Contains(out, "missed") {
		t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errs)
	}
	if code, _, _ := runCmd(run, "", "golden", "--dicts", "unused"); code != 2 {
		t.Fatalf("no --cases: exit %d", code)
	}
}
