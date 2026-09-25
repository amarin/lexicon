package lexicon

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/amarin/gomorphy/pkg/morphology"
)

// wordForm is one entry of a test dictionary: (word, lemma, tag).
type wordForm [3]string

var (
	baseForms = []wordForm{
		{"кот", "кот", "NOUN,anim,masc sing,nomn"},
		{"кота", "кот", "NOUN,anim,masc sing,gent"},
	}
	surnameForms = []wordForm{
		{"кузнецов", "кузнецов", "NOUN,anim,masc,Surn sing,nomn"},
	}
	abbrevTest = BuiltinDict{
		Name: "abbrev.test", Kind: KindAbbrev, Format: FormatTSV,
		Data: []byte("село\tс.\tNOUN\nсын\tс.\tNOUN\nдеревня\tдер.\tNOUN\n"),
	}
)

// datBytes builds a gomorphy dictionary from forms and returns it in GMOR
// format. Every call writes a new BuiltAt; ContentHash ignores it.
func datBytes(t testing.TB, forms []wordForm) []byte {
	t.Helper()

	b := morphology.NewBuilder(morphology.BuilderOptions{Language: "ru"})
	for _, f := range forms {
		if err := b.AddForm(f[0], f[1], f[2]); err != nil {
			t.Fatalf("AddForm: %v", err)
		}
	}

	d, err := b.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	path := filepath.Join(t.TempDir(), "d.dat")
	if err := d.SaveTo(path); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return data
}

// writeFile writes dir/name atomically (temporary file + rename): a
// memory-mapped dictionary file must never be overwritten in place.
func writeFile(t testing.TB, dir, name string, data []byte) string {
	t.Helper()

	path := filepath.Join(dir, name)
	tmp := path + ".tmp"

	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.Rename(tmp, path); err != nil {
		t.Fatal(err)
	}

	return path
}

// openTest opens a registry with the base built from baseForms, the abbrevTest
// built-in and directory dir; closed by t.Cleanup.
func openTest(t testing.TB, dir string, st StateStore) *Registry {
	t.Helper()

	r, err := Open(context.Background(), Options{Dir: dir, Base: datBytes(t, baseForms), Builtin: []BuiltinDict{abbrevTest}, State: st})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	t.Cleanup(func() { _ = r.Close() })

	return r
}

// exact keeps dictionary (not predicted) readings.
func exact(rs []Reading) []Reading {
	var out []Reading

	for _, r := range rs {
		if !r.Predicted {
			out = append(out, r)
		}
	}

	return out
}

// names renders "name:kind|…" of the registry list.
func names(r *Registry) string {
	var out []string
	for _, e := range r.List() {
		out = append(out, e.Name+":"+string(e.Kind))
	}

	return strings.Join(out, "|")
}
