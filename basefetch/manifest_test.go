package basefetch

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/amarin/lexicon"
)

// TestManifestBytes: the sidecar content reads back through
// lexicon.ParseManifest.
func TestManifestBytes(t *testing.T) {
	data, err := manifestBytes("2.4.417127.4579844")
	if err != nil {
		t.Fatal(err)
	}

	m, err := lexicon.ParseManifest(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	if m.Version != "2.4.417127.4579844" || m.License == "" || m.Source == "" || m.URL == "" {
		t.Fatalf("manifest %+v", m)
	}
}

// readMeta parses the final sidecar of dst.
func readMeta(t *testing.T, dst string) lexicon.Manifest {
	t.Helper()

	f, err := os.Open(lexicon.ManifestPath(dst))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	m, err := lexicon.ParseManifest(f)
	if err != nil {
		t.Fatal(err)
	}

	return m
}

// TestFinalizeReplaces: finalize places the dictionary and a sidecar for the
// new version over an old pair, leaving no temp files.
func TestFinalizeReplaces(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, DefaultName)

	for _, v := range []string{"1.0", "2.0"} {
		datTmp := dst + ".tmp"
		if err := os.WriteFile(datTmp, []byte("dict "+v), 0o644); err != nil {
			t.Fatal(err)
		}

		if err := finalize(datTmp, dst, v); err != nil {
			t.Fatal(err)
		}

		if data, _ := os.ReadFile(dst); string(data) != "dict "+v {
			t.Fatalf("dictionary %q, want version %s", data, v)
		}

		if m := readMeta(t, dst); m.Version != v {
			t.Fatalf("manifest version %q, want %q", m.Version, v)
		}
	}

	for _, p := range []string{dst + ".tmp", lexicon.ManifestPath(dst) + ".tmp"} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Fatalf("%s left behind: %v", p, err)
		}
	}
}
