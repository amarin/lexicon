package examples_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestExamplesOutput runs every example whose main.go has an "// Output:"
// block — the ones that need no downloaded data — and compares its stdout
// with that block, like go test checks ExampleXxx functions. Leading spaces
// of expected lines are kept; trailing spaces are ignored on both sides.
func TestExamplesOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and runs every example")
	}

	mains, err := filepath.Glob("*/main.go")
	if err != nil {
		t.Fatal(err)
	}

	for _, mainGo := range mains {
		want, ok := expectedOutput(t, mainGo)
		if !ok {
			continue
		}

		name := filepath.Dir(mainGo)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cmd := exec.Command("go", "run", "./examples/"+name)
			cmd.Dir = ".." // the module root, where ./examples/<name> resolves

			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			got, err := cmd.Output()
			if err != nil {
				t.Fatalf("go run ./examples/%s: %v\n%s", name, err, stderr.String())
			}

			if g := trimLines(string(got)); g != want {
				t.Errorf("output mismatch\n got:\n%s\nwant:\n%s", g, want)
			}
		})
	}
}

// expectedOutput returns the lines of main.go's "// Output:" comment block.
func expectedOutput(t *testing.T, mainGo string) (string, bool) {
	t.Helper()

	src, err := os.ReadFile(mainGo)
	if err != nil {
		t.Fatal(err)
	}

	_, block, ok := strings.Cut(string(src), "// Output:")
	if !ok {
		return "", false
	}

	_, rest, _ := strings.Cut(block, "\n")

	var lines []string

	for line := range strings.Lines(rest) {
		text, isComment := strings.CutPrefix(strings.TrimSpace(line), "//")
		if !isComment {
			break
		}

		lines = append(lines, strings.TrimPrefix(text, " "))
	}

	return trimLines(strings.Join(lines, "\n")), true
}

// trimLines drops trailing spaces of every line and surrounding blank lines.
func trimLines(s string) string {
	var out []string
	for line := range strings.Lines(s) {
		out = append(out, strings.TrimRight(line, " \t\r\n"))
	}

	return strings.Trim(strings.Join(out, "\n"), "\n")
}
