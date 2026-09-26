package rules

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Load decodes one rule file read from r; errors call it "<input>". Use
// LoadNamed or LoadFile to give the file a name.
func Load(r io.Reader) (File, error) { return LoadNamed("", r) }

// LoadNamed decodes one rule file: exactly one YAML document (a JSON
// document is valid YAML and loads too). Unknown fields are errors. name (a
// path or any label, "" = "<input>") prefixes every error as
// "<name>:<line>: <message>" and stays with the file, so Compile errors
// point at the same place.
func LoadNamed(name string, r io.Reader) (File, error) {
	label := fileLabel(name)
	data, err := io.ReadAll(r)
	if err != nil {
		return File{}, fmt.Errorf("%s: %w", label, err)
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var f File
	if err := dec.Decode(&f); err != nil {
		if errors.Is(err, io.EOF) {
			return File{}, fmt.Errorf("%s: empty rule file", label)
		}
		return File{}, located(label, err)
	}
	var extra yaml.Node
	switch err := dec.Decode(&extra); {
	case errors.Is(err, io.EOF):
	case err != nil:
		return File{}, located(label, err)
	default:
		return File{}, fmt.Errorf("%s:%d: a rule file holds one YAML document", label, extra.Line)
	}
	// Second pass for source lines (decision D15): the strict decode above
	// cannot record them without losing KnownFields.
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return File{}, located(label, err)
	}
	f.name = name
	setPositions(&f, &root)
	return f, nil
}

// LoadFile reads and decodes the rule file at path (.yaml, .yml or .json —
// the extension is not checked); errors and Compile errors name path.
func LoadFile(path string) (File, error) {
	fh, err := os.Open(path)
	if err != nil {
		return File{}, fmt.Errorf("rules: %w", err)
	}
	defer fh.Close()
	return LoadNamed(path, fh)
}

// linePrefix matches the "line N: " that yaml.v3 (also as "yaml: line N: "
// for syntax errors) and this package's UnmarshalYAML methods put first.
var linePrefix = regexp.MustCompile(`^(?:yaml: )?line (\d+): `)

// located rewrites a decode error as "<label>:<line>: <message>" — one line
// per message of a *yaml.TypeError — or "<label>: <message>" when the
// message has no line.
func located(label string, err error) error {
	msgs := []string{err.Error()}
	var te *yaml.TypeError
	if errors.As(err, &te) {
		msgs = te.Errors
	}
	out := make([]string, len(msgs))
	for i, m := range msgs {
		if sub := linePrefix.FindStringSubmatch(m); sub != nil {
			out[i] = label + ":" + sub[1] + ": " + m[len(sub[0]):]
		} else {
			out[i] = label + ": " + m
		}
	}
	return errors.New(strings.Join(out, "\n"))
}
