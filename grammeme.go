package lexicon

import "strings"

// stopPOS are service parts of speech: a word whose exact readings all have
// one of them is a stop word.
var stopPOS = map[string]bool{"PREP": true, "CONJ": true, "PRCL": true, "INTJ": true}

// tagSeparators split native gomorphy tags: OpenCorpora/pymorphy2
// "NOUN,anim,masc,Surn sing,ablt" and UniMorph "N;GEN;SG".
const tagSeparators = ", ;"

// hasGrammeme reports whether tag contains g as a whole grammeme. Local
// stand-in for gomorphy 1.3.0 HasGrammeme (decision D27).
func hasGrammeme(tag, g string) bool {
	for tag != "" {
		i := strings.IndexAny(tag, tagSeparators)
		if i < 0 {
			return tag == g
		}

		if tag[:i] == g {
			return true
		}

		tag = tag[i+1:]
	}

	return false
}

// pos is the first grammeme of tag (the part of speech).
func pos(tag string) string {
	if i := strings.IndexAny(tag, tagSeparators); i >= 0 {
		return tag[:i]
	}

	return tag
}

// allStop reports whether there are readings and all are service words.
// Callers pass readings without Abbr ones (see serviceCandidates).
func allStop(rs []Reading) bool {
	for _, r := range rs {
		if !stopPOS[pos(r.Tag)] {
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
		if !hasGrammeme(r.Tag, "Abbr") {
			out = append(out, r)
		}
	}

	return out
}
