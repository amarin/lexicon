package lexicon

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestManifestRoundTrip(t *testing.T) {
	m := Manifest{
		Source: "OpenCorpora", License: "CC BY-SA 4.0", Version: "2.4", URL: "https://example.org/x",
		GeneratedFrom: "test", Extra: map[string]string{"note": "a: b"},
	}

	var buf bytes.Buffer
	if _, err := m.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	got, err := ParseManifest(&buf)
	if err != nil || !reflect.DeepEqual(got, m) {
		t.Fatalf("round trip = %+v, %v; want %+v", got, err, m)
	}
}

// TestParseManifestLenient: comments, blank and malformed lines are skipped;
// keys are case-insensitive; unknown keys go to Extra.
func TestParseManifestLenient(t *testing.T) {
	got, err := ParseManifest(strings.NewReader("# comment\n\nno colon line\nSource:  X \nfoo: bar\n"))
	if err != nil || got.Source != "X" || got.Extra["foo"] != "bar" || len(got.Extra) != 1 {
		t.Fatalf("ParseManifest = %+v, %v", got, err)
	}
}

func TestManifestString(t *testing.T) {
	m := Manifest{Source: "OpenCorpora", License: "CC BY-SA 4.0", Version: "2.4", URL: "https://example.org/x"}
	if got, want := m.String(), "OpenCorpora; 2.4; https://example.org/x; license CC BY-SA 4.0"; got != want {
		t.Fatalf("String = %q, want %q", got, want)
	}

	if !(Manifest{}).IsZero() || m.IsZero() || (Manifest{}).String() != "" {
		t.Fatal("IsZero/String of the zero manifest")
	}

	if ManifestPath("/d/surname.x.dat") != "/d/surname.x.dat.meta" {
		t.Fatal("ManifestPath")
	}
}

// TestMergeManifest: primary fields win, empty ones (and missing Extra keys)
// are filled from the fallback (Options.BaseManifest over gomorphy BuildInfo,
// owner decision 2026-09-25).
func TestMergeManifest(t *testing.T) {
	primary := Manifest{Source: "OpenCorpora", License: "CC BY-SA 4.0", Extra: map[string]string{"note": "host"}}
	fallback := Manifest{Source: "builder", Version: "2.4", URL: "https://example.org/x", Extra: map[string]string{"note": "info", "k": "v"}}

	want := Manifest{
		Source: "OpenCorpora", License: "CC BY-SA 4.0", Version: "2.4", URL: "https://example.org/x",
		Extra: map[string]string{"note": "host", "k": "v"},
	}
	if got := mergeManifest(primary, fallback); !reflect.DeepEqual(got, want) {
		t.Fatalf("mergeManifest = %+v, want %+v", got, want)
	}

	if got := mergeManifest(Manifest{}, fallback); !reflect.DeepEqual(got, fallback) {
		t.Fatalf("zero primary: %+v, want the fallback", got)
	}

	if got := mergeManifest(primary, Manifest{}); !reflect.DeepEqual(got, primary) {
		t.Fatalf("zero fallback: %+v, want the primary", got)
	}

	if primary.Extra["k"] != "" {
		t.Fatal("mergeManifest modified the primary's Extra")
	}
}

// TestMergeManifestExtraIndependent: the result's Extra is always an
// independent clone, even when the fallback's Extra is empty — mutating it
// must not change the primary's Extra (the primary's map must not be
// aliased into the result).
func TestMergeManifestExtraIndependent(t *testing.T) {
	primary := Manifest{Extra: map[string]string{"note": "host"}}

	got := mergeManifest(primary, Manifest{})
	got.Extra["note"] = "mutated"

	if primary.Extra["note"] != "host" {
		t.Fatalf("mergeManifest aliased the primary's Extra: primary.Extra = %+v", primary.Extra)
	}
}
