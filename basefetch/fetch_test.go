package basefetch

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/amarin/lexicon"
)

// TestFinalizeRenameFailureCleansUp: when the dictionary rename fails,
// finalize must leave neither the dictionary tmp file, the manifest tmp
// file, nor a final .meta behind.
func TestFinalizeRenameFailureCleansUp(t *testing.T) {
	dir := t.TempDir()

	datTmp := filepath.Join(dir, "base.dat.tmp")
	if err := os.WriteFile(datTmp, []byte("fake dict"), 0o644); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(dir, "base.dat")
	// dst already exists as a directory, so renaming the dat tmp onto it fails.
	if err := os.Mkdir(dst, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := finalize(datTmp, dst, "1.2.3"); err == nil {
		t.Fatal("finalize: want error when dst cannot be renamed onto")
	}

	if _, err := os.Stat(datTmp); !os.IsNotExist(err) {
		t.Fatalf("dat tmp left behind: %v", err)
	}

	metaTmp := lexicon.ManifestPath(dst) + ".tmp"
	if _, err := os.Stat(metaTmp); !os.IsNotExist(err) {
		t.Fatalf("meta tmp left behind: %v", err)
	}

	if _, err := os.Stat(lexicon.ManifestPath(dst)); !os.IsNotExist(err) {
		t.Fatalf("meta file created despite dat rename failure: %v", err)
	}
}
