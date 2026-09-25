package lexicon

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/amarin/gomorphy/pkg/morphology"
)

// yoReplacer: forms of text and queries come without ё (textnorm), so TSV
// dictionaries are stored with е (decision D28).
var yoReplacer = strings.NewReplacer("ё", "е", "Ё", "Е")

// openTSV builds the dictionary from TSV text; the hash is sha256 of the text.
func (x *dict) openTSV(data []byte) {
	sum := sha256.Sum256(data)
	x.entry.Hash = hex.EncodeToString(sum[:])

	d, err := morphology.ImportTSV(strings.NewReader(yoReplacer.Replace(string(data))),
		morphology.BuilderOptions{Language: "ru", Source: "tsv"})
	if err != nil {
		x.entry.Error = err.Error()

		return
	}

	x.d = d
}

// openDat maps a .dat file (morphology.Open, mmap).
func (x *dict) openDat(path string) {
	d, err := morphology.Open(path)
	if err != nil {
		x.entry.Error = err.Error()

		return
	}

	x.setDat(d)
}

// openBytes opens a built-in .dat in place (morphology.OpenBytes).
func (x *dict) openBytes(data []byte) {
	d, err := morphology.OpenBytes(data)
	if err != nil {
		x.entry.Error = err.Error()

		return
	}

	x.data = data
	x.setDat(d)
}

// setDat stores an opened binary dictionary with its content hash and
// BuildInfo provenance.
func (x *dict) setDat(d *morphology.Dictionary) {
	x.d = d
	x.entry.Hash = d.ContentHash()
	x.entry.Manifest = manifestFromInfo(d.Info())
}

// manifestFromInfo converts gomorphy BuildInfo (nil-safe).
func manifestFromInfo(bi *morphology.BuildInfo) Manifest {
	if bi == nil {
		return Manifest{}
	}

	return Manifest{Source: bi.Source, Version: bi.SourceVersion, URL: bi.SourceURL}
}
