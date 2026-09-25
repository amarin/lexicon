//go:build integration

package basefetch

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/amarin/gomorphy/pkg/morphology"
	"github.com/amarin/lexicon"
)

// TestFetch downloads and compiles the real dictionary (network, ~1 minute).
func TestFetch(t *testing.T) {
	if os.Getenv("LEXICON_FETCH") != "1" {
		t.Skip("set LEXICON_FETCH=1 to download pymorphy2-dicts-ru")
	}

	dst := filepath.Join(t.TempDir(), DefaultName)

	version, err := Fetch(dst)
	if err != nil || version == "" {
		t.Fatalf("Fetch = %q, %v", version, err)
	}

	d, err := morphology.Open(dst)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()

	if rs := d.Parse("кота"); len(rs) == 0 || rs[0].Normal != "кот" {
		t.Fatalf("Parse(кота) = %+v", rs)
	}

	if _, err := os.Stat(lexicon.ManifestPath(dst)); err != nil {
		t.Fatalf("no manifest: %v", err)
	}
}
