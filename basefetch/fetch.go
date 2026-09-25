// Package basefetch downloads the pymorphy2-dicts-ru package (OpenCorpora
// data) from PyPI and compiles it into a gomorphy dictionary file with a
// provenance sidecar. It needs network access and imports gomorphy's pymorphy
// loader (which pulls amarin/logging and zap), so it lives apart from package
// lexicon; hosts that ship their own base dictionary never compile it.
package basefetch

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/amarin/gomorphy/pkg/morphology"
	"github.com/amarin/gomorphy/pkg/pymorphy"
	"github.com/amarin/lexicon"
)

// DefaultName is the conventional file name of the base dictionary in a
// dictionary directory.
const DefaultName = "base.opencorpora.dat"

// Fetch downloads pymorphy2-dicts-ru, compiles it and saves it to dst
// atomically (temporary file + rename), together with the sidecar
// lexicon.ManifestPath(dst). It returns the dictionary package version.
func Fetch(dst string) (string, error) {
	if err := ensureLogging(); err != nil {
		return "", err
	}

	work, err := os.MkdirTemp("", "lexicon-basedict-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(work)

	loader := pymorphy.NewLoader(work)
	if err := loader.Sync(false); err != nil {
		return "", fmt.Errorf("basefetch: download pymorphy2-dicts-ru: %w", err)
	}

	d, err := morphology.OpenPyMorphyDense(loader.UnpackedDirPath())
	if err != nil {
		return "", fmt.Errorf("basefetch: compile: %w", err)
	}
	defer d.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", err
	}

	tmp := dst + ".tmp"
	if err := d.SaveTo(tmp); err != nil {
		_ = os.Remove(tmp)

		return "", fmt.Errorf("basefetch: save: %w", err)
	}

	version, _ := loader.LocalVersion() // found-bool ignored: an empty version is allowed

	if err := finalize(tmp, dst, version); err != nil {
		return "", err
	}

	return version, nil
}

// finalize places the compiled dictionary and its manifest sidecar next to
// dst, given the already-saved dictionary temp file datTmp. It writes the
// manifest to its own temp file first, then renames the dictionary tmp into
// place before the manifest tmp (owner ruling): a resulting .dat without a
// .meta is acceptable, since the registry falls back to gomorphy BuildInfo,
// but a final .meta must never exist without its .dat. On any error it
// removes whichever temp files still exist.
func finalize(datTmp, dst, version string) error {
	metaPath := lexicon.ManifestPath(dst)
	metaTmp := metaPath + ".tmp"

	data, err := manifestBytes(version)
	if err != nil {
		_ = os.Remove(datTmp)

		return err
	}

	if err := os.WriteFile(metaTmp, data, 0o644); err != nil {
		_ = os.Remove(datTmp)

		return err
	}

	if err := os.Rename(datTmp, dst); err != nil {
		_ = os.Remove(datTmp)
		_ = os.Remove(metaTmp)

		return fmt.Errorf("basefetch: save: %w", err)
	}

	if err := os.Rename(metaTmp, metaPath); err != nil {
		_ = os.Remove(metaTmp)

		return err
	}

	return nil
}
