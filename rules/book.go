package rules

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"go.yaml.in/yaml/v3"
)

// Book is an immutable, validated collection of rule sets.
type Book struct {
	sets    []*compiledSet
	version string
}

// Compile validates files and returns a Book; set names must be unique
// across files. Compile() with no files returns an empty book. Validation
// errors read "<file>:<line>: <set>/<rule>: <message>" (decision D15).
func Compile(files ...File) (*Book, error) {
	b := &Book{}
	first := map[string]string{} // set name → where it is defined first
	for fi := range files {
		f := &files[fi]
		for si := range f.Sets {
			rs := &f.Sets[si]
			if rs.Name == "" {
				return nil, &ruleError{file: f.name, line: rs.line, set: fmt.Sprintf("#%d", si), msg: "empty set name"}
			}
			if at, dup := first[rs.Name]; dup {
				return nil, &ruleError{file: f.name, line: rs.line, set: rs.Name, msg: "duplicate set name (first defined at " + at + ")"}
			}
			first[rs.Name] = where(f.name, rs.line)
			cs, err := compileSet(f.name, rs)
			if err != nil {
				return nil, err
			}
			b.sets = append(b.sets, cs)
		}
	}
	data, err := yaml.Marshal(files)
	if err != nil {
		return nil, fmt.Errorf("rules: %w", err)
	}
	sum := sha256.Sum256(data)
	b.version = hex.EncodeToString(sum[:])
	return b, nil
}

// Version changes whenever any rule or meta value changes.
func (b *Book) Version() string { return b.version }

// Sets returns the names of all sets in order.
func (b *Book) Sets() []string {
	out := make([]string, len(b.sets))
	for i, s := range b.sets {
		out[i] = s.name
	}
	return out
}

// Active returns the rules of the sets whose When tags are all in tags.
func (b *Book) Active(tags []string) Active {
	var a Active
	for _, s := range b.sets {
		if !s.activeFor(tags) {
			continue
		}
		a.Sets = append(a.Sets, s.name)
		a.Hints = append(a.Hints, s.hints...)
		a.Triggers = append(a.Triggers, s.triggers...)
	}
	return a
}
