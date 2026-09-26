package lexicon

import "strings"

// Flag marks properties of a lemma.
type Flag uint16

const (
	FlagAmbiguous Flag = 1 << iota // the form has more than one lemma (or is a one-letter abbreviation)
	FlagPredicted                  // not in dictionaries; predicted by the base dictionary from the ending
	FlagUnknown                    // no reading and no prediction: the lemma is the form itself
	FlagAbbrev                     // the form abbreviates the lemma (an abbreviation dictionary expands it)
	FlagStop                       // service word (ModeFull only; ModeIndex drops stop words)
	FlagReform                     // found through a pre-reform ending variant (textnorm.ReformVariants)
)

var flagNames = [...]string{"ambiguous", "predicted", "unknown", "abbrev", "stop", "reform"}

// String lists the flags separated by commas (CLI and tests).
func (f Flag) String() string {
	var out []string

	for i, name := range flagNames {
		if f&(1<<i) != 0 {
			out = append(out, name)
		}
	}

	return strings.Join(out, ",")
}
