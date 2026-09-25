package lexicon

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// TestOpenLoadsBuiltinsAndFiles: base, built-ins in order, then files by name;
// all enabled, hashed, with formats and origins.
func TestOpenLoadsBuiltinsAndFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "surname.test.dat", datBytes(t, surnameForms))
	tsv := writeFile(t, dir, "custom.x.tsv", []byte("село\tсельцо\tNOUN\n"))

	r := openTest(t, dir, nil)

	if got, want := names(r), "base.builtin:base|abbrev.test:abbrev|custom.x:custom|surname.test:surname"; got != want {
		t.Fatalf("List = %s, want %s", got, want)
	}

	for _, e := range r.List() {
		if !e.Enabled || e.Error != "" || e.Hash == "" {
			t.Errorf("%s: Enabled=%v Error=%q Hash=%q", e.Name, e.Enabled, e.Error, e.Hash)
		}
	}

	l := r.List()
	if l[0].Format != FormatDat || l[0].Origin != OriginBuiltin || l[2].Format != FormatTSV || l[2].Origin != tsv {
		t.Fatalf("formats/origins %+v", l)
	}
}

// TestParseByKinds (genodex P1): empty kinds = all; kinds narrow parsing; the
// reading names its dictionary.
func TestParseByKinds(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "surname.test.dat", datBytes(t, surnameForms))

	r := openTest(t, dir, nil)

	got := exact(r.Parse("кота", nil))
	if len(got) != 1 || got[0].Normal != "кот" || got[0].Kind != KindBase || got[0].Dict != BaseBuiltinName {
		t.Fatalf("Parse(кота, nil) = %+v", got)
	}

	for _, rd := range r.Parse("кота", []Kind{"surname"}) {
		if rd.Kind != "surname" {
			t.Fatalf("Parse(кота, surname) returned kind %q", rd.Kind)
		}
	}

	got = exact(r.Parse("кузнецов", []Kind{"surname"}))
	if len(got) != 1 || got[0].Kind != "surname" || got[0].Dict != "surname.test" {
		t.Fatalf("Parse(кузнецов, surname) = %+v", got)
	}
}

// TestBuiltinAbbrev (genodex P1): a TSV built-in with dotted, ambiguous forms.
func TestBuiltinAbbrev(t *testing.T) {
	r := openTest(t, "", nil)

	var normals []string
	for _, rd := range r.Parse("с.", []Kind{KindAbbrev}) {
		normals = append(normals, rd.Normal)
	}

	slices.Sort(normals)

	if strings.Join(normals, "|") != "село|сын" {
		t.Fatalf("Parse(с.) = %v, want [село сын]", normals)
	}
}

// TestFileReplacesBuiltinBase: any base.* file drops the built-in base (D18).
func TestFileReplacesBuiltinBase(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "base.other.dat", datBytes(t, surnameForms))

	r := openTest(t, dir, nil)

	var bases []Entry
	for _, e := range r.List() {
		if e.Kind == KindBase {
			bases = append(bases, e)
		}
	}

	if len(bases) != 1 || bases[0].Origin != path || bases[0].Name != "base.other" {
		t.Fatalf("base dictionaries = %+v, want one from %s", bases, path)
	}

	if len(exact(r.Parse("кота", nil))) != 0 {
		t.Fatal("built-in base not replaced: «кота» parses")
	}
}

// TestFileReplacesBuiltin: a file named like a built-in replaces it in place.
func TestFileReplacesBuiltin(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "abbrev.test.tsv", []byte("город\tг.\tNOUN\n"))

	r := openTest(t, dir, nil)

	if got := names(r); got != "base.builtin:base|abbrev.test:abbrev" {
		t.Fatalf("List = %s", got)
	}

	if r.List()[1].Origin != path || len(exact(r.Parse("с.", nil))) != 0 || len(exact(r.Parse("г.", nil))) != 1 {
		t.Fatalf("built-in not replaced: %+v", r.List())
	}
}

// TestBrokenInvalidAndDuplicateFiles: listed with an error, never fatal, not
// used (D17).
func TestBrokenInvalidAndDuplicateFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "custom.bad.dat", []byte("garbage"))
	writeFile(t, dir, "mine.dat", datBytes(t, surnameForms))
	writeFile(t, dir, "custom.x.dat", datBytes(t, surnameForms))
	writeFile(t, dir, "custom.x.tsv", []byte("село\tсельцо\tNOUN\n"))

	r := openTest(t, dir, nil)

	errs := map[string]int{}
	for _, e := range r.List() {
		if e.Error != "" {
			errs[e.Name]++
		}
	}

	if errs["custom.bad"] != 1 || errs["mine"] != 1 || errs["custom.x"] != 1 {
		t.Fatalf("errors %v in %+v", errs, r.List())
	}

	if got := exact(r.Parse("кота", nil)); len(got) != 1 {
		t.Fatalf("Parse(кота) = %+v", got)
	}

	if !strings.HasSuffix(r.Summary(), "broken: 3") {
		t.Fatalf("Summary = %q", r.Summary())
	}
}

// TestBuiltinKindMismatch: a built-in whose Kind differs from its name prefix
// is listed with an error.
func TestBuiltinKindMismatch(t *testing.T) {
	r, err := Open(t.Context(), Options{Builtin: []BuiltinDict{{Name: "abbrev.x", Kind: "surname", Format: FormatTSV, Data: []byte("a\tb\n")}}})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	if l := r.List(); len(l) != 1 || l[0].Error == "" {
		t.Fatalf("List = %+v", l)
	}
}

// TestDeclaredKinds (owner decision 7, D17): with Options.Kinds set, a file
// of an undeclared kind — here the typo «surnme» — is listed with an
// "unknown kind" error and does not take part in parsing; declared kinds,
// base and abbrev load as usual.
func TestDeclaredKinds(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "surname.test.dat", datBytes(t, surnameForms))
	writeFile(t, dir, "surnme.x.tsv", []byte("петров\tпетрова\tNOUN\n"))

	r, err := Open(t.Context(), Options{Dir: dir, Base: datBytes(t, baseForms), Builtin: []BuiltinDict{abbrevTest}, Kinds: []Kind{"surname"}})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	if got := names(r); got != "base.builtin:base|abbrev.test:abbrev|surname.test:surname|surnme.x:surnme" {
		t.Fatalf("List = %s", got)
	}

	for _, e := range r.List() {
		want := ""
		if e.Name == "surnme.x" {
			want = `lexicon: dictionary "surnme.x": unknown kind "surnme"`
		}

		if e.Error != want {
			t.Errorf("%s: Error = %q, want %q", e.Name, e.Error, want)
		}
	}

	if got := exact(r.Parse("петрова", nil)); len(got) != 0 {
		t.Fatalf("undeclared kind parsed: %+v", got)
	}

	if got := exact(r.Parse("кузнецов", []Kind{"surname"})); len(got) != 1 {
		t.Fatalf("declared kind not parsed: %+v", got)
	}

	if !strings.HasSuffix(r.Summary(), "broken: 1") {
		t.Fatalf("Summary = %q", r.Summary())
	}
}

// TestNilKindsAcceptAll: nil Options.Kinds accepts every valid kind (lexicon
// CLI); the same «surnme.x.tsv» loads and parses.
func TestNilKindsAcceptAll(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "surnme.x.tsv", []byte("петров\tпетрова\tNOUN\n"))

	r := openTest(t, dir, nil)

	for _, e := range r.List() {
		if e.Error != "" {
			t.Fatalf("%s: Error = %q", e.Name, e.Error)
		}
	}

	if got := exact(r.Parse("петрова", []Kind{"surnme"})); len(got) != 1 || got[0].Normal != "петров" {
		t.Fatalf("Parse(петрова) = %+v", got)
	}
}

// TestBuiltinUndeclaredKind: a built-in of an undeclared kind is a programming
// error of the host — Open fails; an invalid Options.Kinds entry fails too.
func TestBuiltinUndeclaredKind(t *testing.T) {
	b := BuiltinDict{Name: "given.x", Kind: "given", Format: FormatTSV, Data: []byte("иван\tивана\tNOUN\n")}
	if _, err := Open(t.Context(), Options{Builtin: []BuiltinDict{b}, Kinds: []Kind{"surname"}}); err == nil ||
		err.Error() != `lexicon: built-in "given.x": unknown kind "given"` {
		t.Fatalf("Open = %v", err)
	}

	if _, err := Open(t.Context(), Options{Builtin: []BuiltinDict{abbrevTest}, Kinds: []Kind{"surname"}}); err != nil {
		t.Fatalf("abbrev built-in with declared kinds: %v", err)
	}

	if _, err := Open(t.Context(), Options{Kinds: []Kind{"Surname"}}); err == nil {
		t.Fatal("Open accepted an invalid declared kind")
	}
}

// TestNoBaseNoDir: an empty registry works; a missing directory is empty and
// not created (D19).
func TestNoBaseNoDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing")

	r, err := Open(t.Context(), Options{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	if len(r.List()) != 0 || r.Parse("кот", nil) != nil || r.Version() == "" {
		t.Fatalf("empty registry: %+v, version %q", r.List(), r.Version())
	}

	if got := r.Summary(); got != "dictionaries: 0 of 0 enabled; base: no; broken: 0" {
		t.Fatalf("Summary = %q", got)
	}

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("directory created: %v", err)
	}
}

// TestPredictedOnlyFromBase: user dictionaries never predict, abbreviations
// never predict, exact readings exclude predictions (gomorphy A).
func TestPredictedOnlyFromBase(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "surname.test.dat", datBytes(t, surnameForms))

	r := openTest(t, dir, nil)

	if got := r.Parse("петрова", []Kind{"surname"}); len(got) != 0 {
		t.Fatalf("surname dictionary predicted: %+v", got)
	}

	if got := r.Parse("дер", []Kind{KindAbbrev}); len(got) != 0 {
		t.Fatalf("abbreviation dictionary predicted: %+v", got)
	}

	for _, rd := range r.Parse("петрова", nil) {
		if rd.Kind != KindBase || !rd.Predicted {
			t.Fatalf("prediction from %s: %+v", rd.Dict, rd)
		}
	}

	if got := r.Parse("кузнецов", nil); len(exact(got)) != len(got) || len(got) == 0 {
		t.Fatalf("Parse(кузнецов) mixes predictions: %+v", got)
	}
}

// TestVersionStableAcrossRebuild: rebuilding identical content keeps the
// version (gomorphy E); a content change changes it.
func TestVersionStableAcrossRebuild(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "surname.test.dat", datBytes(t, surnameForms))

	v1 := openTest(t, dir, nil).Version()

	writeFile(t, dir, "surname.test.dat", datBytes(t, surnameForms)) // new BuiltAt

	if v2 := openTest(t, dir, nil).Version(); v2 != v1 {
		t.Fatalf("rebuild changed version: %s → %s", v1, v2)
	}

	more := append(slices.Clone(surnameForms), wordForm{"петров", "петров", "NOUN,anim,masc,Surn sing,nomn"})
	writeFile(t, dir, "surname.test.dat", datBytes(t, more))

	if v3 := openTest(t, dir, nil).Version(); v3 == v1 {
		t.Fatal("content change kept the version")
	}
}

// TestTSVYoAndCase: ё in TSV is replaced with е at load (P1); mixed-case TSV
// forms are found by lower-case words (gomorphy C).
func TestTSVYoAndCase(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "surname.yo.tsv", []byte("королёв\tкоролёв\tNOUN,Surn\n"))
	writeFile(t, dir, "toponym.case.tsv", []byte("Москва\tМосква\tNOUN,Geox\n"))

	r := openTest(t, dir, nil)

	if got := exact(r.Parse("королев", []Kind{"surname"})); len(got) != 1 || got[0].Normal != "королев" {
		t.Fatalf("Parse(королев) = %+v", got)
	}

	if got := exact(r.Parse("москва", []Kind{"toponym"})); len(got) != 1 {
		t.Fatalf("Parse(москва) = %+v", got)
	}
}

// TestManifests: a sidecar wins; without one, .dat provenance comes from
// gomorphy BuildInfo.
func TestManifests(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "surname.test.dat", datBytes(t, surnameForms))
	writeFile(t, dir, filepath.Base(ManifestPath(p)), []byte("source: genodex surnames export\nlicense: CC0-1.0\n"))
	writeFile(t, dir, "toponym.x.dat", datBytes(t, baseForms))

	r := openTest(t, dir, nil)

	got := map[string]Manifest{}
	for _, e := range r.List() {
		got[e.Name] = e.Manifest
	}

	if m := got["surname.test"]; m.Source != "genodex surnames export" || m.License != "CC0-1.0" {
		t.Fatalf("sidecar manifest %+v", m)
	}

	if m := got["toponym.x"]; m.Source != "builder" {
		t.Fatalf("BuildInfo manifest %+v", m)
	}
}

// TestBaseManifest: Options.BaseManifest (the embedded basefetch sidecar)
// describes the built-in base; its fields win and empty ones come from gomorphy
// BuildInfo (owner decision 2026-09-25). Without it the base carries BuildInfo
// only; a base.* file replacing the built-in base does not inherit it.
func TestBaseManifest(t *testing.T) {
	baseEntry := func(r *Registry) Entry {
		t.Helper()

		for _, e := range r.List() {
			if e.Kind == KindBase {
				return e
			}
		}

		t.Fatalf("no base entry in %s", names(r))

		return Entry{}
	}

	r, err := Open(t.Context(), Options{Base: datBytes(t, baseForms), BaseManifest: Manifest{License: "CC BY-SA 4.0"}})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = r.Close() })

	e := baseEntry(r)
	if e.Name != BaseBuiltinName || e.Manifest.License != "CC BY-SA 4.0" || e.Manifest.Source != "builder" {
		t.Fatalf("base manifest %+v, want license from BaseManifest and source from BuildInfo", e)
	}

	if !strings.Contains(e.Manifest.String(), "license CC BY-SA 4.0") {
		t.Fatalf("attribution %q", e.Manifest.String())
	}

	if m := baseEntry(openTest(t, "", nil)).Manifest; m.License != "" || m.Source != "builder" {
		t.Fatalf("without BaseManifest: %+v, want BuildInfo only", m)
	}

	dir := t.TempDir()
	writeFile(t, dir, "base.other.dat", datBytes(t, surnameForms))

	r2, err := Open(t.Context(), Options{Dir: dir, Base: datBytes(t, baseForms), BaseManifest: Manifest{License: "CC BY-SA 4.0"}})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = r2.Close() })

	if e := baseEntry(r2); e.Name != "base.other" || e.Manifest.License != "" {
		t.Fatalf("file base %+v must not take BaseManifest", e)
	}
}

// TestSymlinkedFiles: a symbolic link to a dictionary file is followed; a
// dangling link is listed with an error and does not take part in parsing.
func TestSymlinkedFiles(t *testing.T) {
	src := t.TempDir()
	target := writeFile(t, src, "real.dat", datBytes(t, surnameForms))

	dir := t.TempDir()
	if err := os.Symlink(target, filepath.Join(dir, "surname.linked.dat")); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}

	if err := os.Symlink(filepath.Join(src, "missing.dat"), filepath.Join(dir, "surname.dangling.dat")); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(src, filepath.Join(dir, "surname.dir.dat")); err != nil {
		t.Fatal(err)
	}

	r := openTest(t, dir, nil)

	byName := map[string]Entry{}
	for _, e := range r.List() {
		byName[e.Name] = e
	}

	if e, ok := byName["surname.linked"]; !ok || e.Error != "" {
		t.Fatalf("linked file: %+v", r.List())
	}

	if e, ok := byName["surname.dangling"]; !ok || e.Error == "" {
		t.Fatalf("dangling link: %+v", r.List())
	}

	if _, ok := byName["surname.dir"]; ok {
		t.Fatalf("link to a directory listed: %+v", r.List())
	}

	if got := exact(r.Parse("кузнецов", []Kind{"surname"})); len(got) != 1 {
		t.Fatalf("Parse(кузнецов) = %+v", got)
	}
}
