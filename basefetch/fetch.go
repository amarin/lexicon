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

	version, _ := loader.LocalVersion()

	if err := writeManifest(lexicon.ManifestPath(dst), version); err != nil {
		_ = os.Remove(tmp)

		return "", err
	}

	if err := os.Rename(tmp, dst); err != nil {
		return "", err
	}

	return version, nil
}
