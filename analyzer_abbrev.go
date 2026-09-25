package lexicon

import "unicode/utf8"

// abbrevKinds is the kind list of abbreviation lookups. The lookup ignores the
// profile's kinds (genodex P1 parity, D16).
var abbrevKinds = []Kind{KindAbbrev}

// abbrev returns the lemmas of an abbreviation key («кр-нин», «с.»); nil when
// the key is not an abbreviation (cached from Task 10).
func (a *Analyzer) abbrev(key string) []Lemma {
	return a.lookupAbbrev(key)
}

// lookupAbbrev: exact readings of abbreviation dictionaries only.
func (a *Analyzer) lookupAbbrev(key string) []Lemma {
	exact, _ := a.parse(key, abbrevKinds)

	return a.lemmas(exact, 0)
}

// dottedAbbrev looks up form+"." as an abbreviation; a one-letter dotted
// abbreviation («с.», «ц.») is always ambiguous (D14).
func (a *Analyzer) dottedAbbrev(form string) []Lemma {
	ls := a.abbrev(form + ".")
	if ls != nil && utf8.RuneCountInString(form) == 1 {
		for i := range ls {
			ls[i].Flags |= FlagAmbiguous
		}
	}

	return ls
}
