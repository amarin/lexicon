package basefetch

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/amarin/lexicon"
)

// TestWriteManifest: the sidecar reads back through lexicon.ParseManifest.
func TestWriteManifest(t *testing.T) {
	path := lexicon.ManifestPath(filepath.Join(t.TempDir(), DefaultName))
	if err := writeManifest(path, "2.4.417127.4579844"); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	m, err := lexicon.ParseManifest(f)
	if err != nil {
		t.Fatal(err)
	}

	if m.Version != "2.4.417127.4579844" || m.License == "" || m.Source == "" || m.URL == "" {
		t.Fatalf("manifest %+v", m)
	}
}
