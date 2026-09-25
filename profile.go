package lexicon

// Profile selects the dictionaries a field is parsed with — the main defence
// against homonymy of names and common words («Вера», «Мороз»). Hosts define
// profiles (genodex: text = all; name = base filtered by Name|Surn|Patr plus
// surname/given/patronymic; place = base filtered by Geox plus toponym and
// abbrev).
type Profile struct {
	// Name identifies the profile; the Analyzer keys its lemma cache by it, so
	// names must be unique per Analyzer. "" disables caching.
	Name string
	// Kinds are the dictionaries consulted, in registry order; empty = all.
	Kinds []Kind
	// Grammemes keep an exact reading of the kind only if its tag has one of
	// the listed grammemes; an absent kind or an empty list keeps everything.
	Grammemes map[Kind][]string
}

// keep reports whether a reading passes the grammeme filter of its kind.
func (p Profile) keep(r Reading) bool {
	want := p.Grammemes[r.Kind]
	if len(want) == 0 {
		return true
	}

	for _, g := range want {
		if hasGrammeme(r.Tag, g) {
			return true
		}
	}

	return false
}

// filter returns the readings that pass keep.
func (p Profile) filter(rs []Reading) []Reading {
	if len(p.Grammemes) == 0 {
		return rs
	}

	var out []Reading

	for _, r := range rs {
		if p.keep(r) {
			out = append(out, r)
		}
	}

	return out
}
