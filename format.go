package lexicon

// Format is the storage format of a dictionary.
type Format uint8

const (
	FormatDat Format = iota + 1 // gomorphy binary (GMOR), as written by SaveTo
	FormatTSV                   // lemma<TAB>wordform[<TAB>tags], built at load
)

var formatNames = [...]string{"", "dat", "tsv"}

// String returns "dat" or "tsv" ("" for the zero value).
func (f Format) String() string {
	if int(f) < len(formatNames) {
		return formatNames[f]
	}

	return "unknown"
}
