package gazetteer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func tsvLine(fields ...string) string { return strings.Join(fields, "\t") }

func writeTSV(t *testing.T, lines ...string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "places.tsv")
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestTSVSource(t *testing.T) {
	path := writeTSV(t,
		"# source: test gazetteer",
		"# license: CC0-1.0",
		"# a comment without a key",
		tsvLine("division", "division:1", "Лягушкино", "Лягушкино", "", "division_type=деревня; level=settlement"),
		tsvLine("division", "division:1", "Лягушкино", "дер. Лягушкина"),
		tsvLine("surname", "surname:2", "Мороз", "Мороз", "requires_context"),
		"broken line",
		tsvLine("surname", "surname:3", "Ус", "Ус", "shiny"),
		"# license: ignored after the first entry",
	)
	src := NewTSVSource("places", path)
	var got []Entry
	err := src.Entries(context.Background(), func(e Entry) error {
		got = append(got, e)
		return nil
	})
	var pe *ParseErrors
	if !errors.As(err, &pe) {
		t.Fatalf("want *ParseErrors, got %v", err)
	}
	if len(pe.Lines) != 2 || pe.Lines[0].Line != 7 || pe.Lines[1].Line != 8 {
		t.Fatalf("bad lines = %+v", pe.Lines)
	}
	if len(got) != 3 {
		t.Fatalf("entries = %+v", got)
	}
	if e := got[0]; e.Type != "division" || e.Ref != "division:1" || e.Canonical != "Лягушкино" ||
		e.Attrs["division_type"] != "деревня" || e.Attrs["level"] != "settlement" || e.Flags != 0 {
		t.Fatalf("entry 0 = %+v", e)
	}
	if e := got[1]; e.Alias != "дер. Лягушкина" || e.Attrs != nil {
		t.Fatalf("entry 1 = %+v", e)
	}
	if e := got[2]; !e.Flags.Has(RequiresContext) {
		t.Fatalf("entry 2 = %+v", e)
	}
	m := src.Manifest()
	if len(m) != 2 || m["source"] != "test gazetteer" || m["license"] != "CC0-1.0" {
		t.Fatalf("manifest = %v", m)
	}
}

func TestTSVSourceVersionFollowsContent(t *testing.T) {
	path := writeTSV(t, tsvLine("surname", "surname:1", "Петров", "Петров"))
	src := NewTSVSource("s", path)
	v1, err := src.Version(context.Background())
	if err != nil || v1 == "" {
		t.Fatalf("v1 = %q, %v", v1, err)
	}
	if v, _ := src.Version(context.Background()); v != v1 {
		t.Fatal("version not stable")
	}
	if err := os.WriteFile(path, []byte(tsvLine("surname", "surname:1", "Петров", "Петровъ")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if v2, _ := src.Version(context.Background()); v2 == v1 {
		t.Fatal("version did not change with content")
	}
	if _, err := NewTSVSource("x", filepath.Join(t.TempDir(), "missing.tsv")).Version(context.Background()); err == nil {
		t.Fatal("missing file: no error")
	}
}

// TestTSVSourceData: in-memory content (a host's //go:embed file) is parsed
// exactly like a file; the version is the sha256 of the data.
func TestTSVSourceData(t *testing.T) {
	data := []byte(strings.Join([]string{
		"# source: embedded forms",
		tsvLine("given_name", "grp:1", "Иоанн", "Иоанн", "", "form=church"),
		tsvLine("given_name", "grp:1", "Иоанн", "Иван", "", "form=folk"),
		"broken line",
	}, "\n") + "\n")
	src := NewTSVSourceData("forms", data)
	if src.Name() != "forms" {
		t.Fatalf("Name() = %q", src.Name())
	}
	v, err := src.Version(context.Background())
	sum := sha256.Sum256(data)
	if err != nil || v != hex.EncodeToString(sum[:]) {
		t.Fatalf("Version() = %q, %v", v, err)
	}
	var got []Entry
	err = src.Entries(context.Background(), func(e Entry) error {
		got = append(got, e)
		return nil
	})
	var pe *ParseErrors
	if !errors.As(err, &pe) || len(pe.Lines) != 1 || pe.Lines[0].Line != 4 {
		t.Fatalf("want one bad line 4, got %v", err)
	}
	if len(got) != 2 || got[1].Alias != "Иван" || got[1].Ref != "grp:1" || got[1].Attrs["form"] != "folk" {
		t.Fatalf("entries = %+v", got)
	}
	if m := src.Manifest(); m["source"] != "embedded forms" {
		t.Fatalf("manifest = %v", m)
	}
	other := NewTSVSourceData("forms", []byte(tsvLine("given_name", "grp:1", "Иоанн", "Ваня")+"\n"))
	if v2, _ := other.Version(context.Background()); v2 == v {
		t.Fatal("version does not follow the data")
	}
}
