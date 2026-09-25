package basefetch

import (
	"bytes"
	"os"

	"github.com/amarin/lexicon"
)

// manifest is the provenance of the fetched dictionary; OpenCorpora data is
// distributed under CC BY-SA and needs attribution.
func manifest(version string) lexicon.Manifest {
	return lexicon.Manifest{
		Source:        "OpenCorpora via pymorphy2-dicts-ru",
		License:       "CC BY-SA 4.0",
		Version:       version,
		URL:           "https://pypi.org/project/pymorphy2-dicts-ru/",
		GeneratedFrom: "lexicon basefetch",
	}
}

// manifestBytes renders the sidecar content for version.
func manifestBytes(version string) ([]byte, error) {
	var b bytes.Buffer
	if _, err := manifest(version).WriteTo(&b); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// writeManifest writes the sidecar atomically (own temp file, then rename to
// path).
func writeManifest(path, version string) error {
	data, err := manifestBytes(version)
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}
