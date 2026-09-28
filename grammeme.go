package lexicon

import "github.com/amarin/gomorphy/pkg/morphology"

// stopPOS are service parts of speech: a word whose exact readings all have
// one of them is a stop word.
var stopPOS = map[string]bool{"PREP": true, "CONJ": true, "PRCL": true, "INTJ": true}

// allStop reports whether there are readings and all are service words.
// Callers pass readings without Abbr ones (see serviceCandidates).
func allStop(rs []Reading) bool {
	for _, r := range rs {
		if !stopPOS[morphology.POS(r.Tag)] {
			return false
		}
	}

	return len(rs) > 0
}

// serviceCandidates are the readings that decide whether a word is a stop
// word: readings tagged Abbr are ignored — OpenCorpora gives one-letter
// service words («в», «с», «и») letter-name abbreviation readings, and
// undotted one-letter forms never take abbreviation readings (D14).
func serviceCandidates(rs []Reading) []Reading {
	var out []Reading

	for _, r := range rs {
		if !morphology.HasGrammeme(r.Tag, "Abbr") {
			out = append(out, r)
		}
	}

	return out
}
