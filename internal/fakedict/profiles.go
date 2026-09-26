package fakedict

import (
	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/textnorm"
)

// AllKinds lists every kind of the fixture in lookup order.
func AllKinds() []lexicon.Kind {
	return []lexicon.Kind{lexicon.KindBase, lexicon.KindAbbrev, KindGiven, KindSurname, KindPatronymic, KindToponym}
}

// Profiles returns the fixture profiles: "text" (all kinds), "name" (base
// filtered to Name|Surn|Patr + name kinds) and "place" (base, abbreviations,
// toponyms).
func Profiles() map[string]lexicon.Profile {
	return map[string]lexicon.Profile{
		"text": {Name: "text", Kinds: AllKinds()},
		"name": {
			Name:      "name",
			Kinds:     []lexicon.Kind{lexicon.KindBase, KindGiven, KindSurname, KindPatronymic},
			Grammemes: map[lexicon.Kind][]string{lexicon.KindBase: {"Name", "Surn", "Patr"}},
		},
		"place": {Name: "place", Kinds: []lexicon.Kind{lexicon.KindBase, lexicon.KindAbbrev, KindToponym}},
	}
}

// TypeProfiles maps gazetteer entry types to the profile their aliases are
// analyzed with; other types use the "text" profile.
func TypeProfiles() map[string]lexicon.Profile {
	p := Profiles()
	return map[string]lexicon.Profile{
		"given_name": p["name"],
		"surname":    p["name"],
		"patronymic": p["name"],
		"division":   p["place"],
		"street":     p["place"],
	}
}

// NewAnalyzer returns a v0.1 analyzer over Genealogy with pre-reform rules.
func NewAnalyzer() *lexicon.Analyzer {
	return lexicon.NewAnalyzer(Genealogy(), textnorm.PreReform, lexicon.AnalyzerOptions{})
}
