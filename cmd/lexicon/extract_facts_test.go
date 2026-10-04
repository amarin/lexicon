package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const kinshipRules = `# «X, дочь/сын [сословие] Y» → child_of(child: X, parent: Y).
sets:
  - name: kinship
    patterns:
      - name: child-of
        elements:
          - {type: given_name, role: child}
          - {token: punct, repeat: "?"}
          - lemma: сын|дочь
          - {type: estate, repeat: "?"}
          - {type: given_name, role: parent}
        actions:
          - emit: {kind: child_of, args: {child: child, parent: parent}}
`

func writeRules(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "kinship.yaml")
	if err := os.WriteFile(path, []byte(kinshipRules), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExtractFactsJSONL(t *testing.T) {
	useFakeDictionaries(t)
	code, out, errs := runCmd(run, "",
		"extract", "--dicts", "unused", "--gazetteer", testTSV, "--rules", writeRules(t),
		"--format", "jsonl", "Мария, дочь крестьянина Ивана")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errs)
	}
	var fact factJSON
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.Contains(line, `"fact"`) {
			if err := json.Unmarshal([]byte(line), &fact); err != nil {
				t.Fatal(err)
			}
		}
	}
	if fact.Doc != 1 || fact.Fact != "child_of" || fact.Args["child"] != 0 || fact.Args["parent"] != 2 || fact.Rule != "child-of" {
		t.Fatalf("fact = %+v\n%s", fact, out)
	}
}

func TestExtractFactsTable(t *testing.T) {
	useFakeDictionaries(t)
	code, out, errs := runCmd(run, "",
		"extract", "--dicts", "unused", "--gazetteer", testTSV, "--rules", writeRules(t),
		"Мария, дочь крестьянина Ивана")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errs)
	}
	for _, want := range []string{"fact", "child_of", "child=«Мария»", "parent=«Ивана»"} {
		if !strings.Contains(out, want) {
			t.Fatalf("table lacks %q:\n%s", want, out)
		}
	}
}
