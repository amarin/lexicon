package basefetch

import (
	"bytes"

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
