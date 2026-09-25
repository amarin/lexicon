package lexicon

import "slices"

// Host-defined kinds used by the tests (genodex names).
const (
	kindSurname    Kind = "surname"
	kindGiven      Kind = "given"
	kindPatronymic Kind = "patronymic"
	kindToponym    Kind = "toponym"
)

// Genodex-like profiles; hosts define their own.
var (
	profText = Profile{Name: "text"}
	profName = Profile{
		Name:      "name",
		Kinds:     []Kind{KindBase, kindSurname, kindGiven, kindPatronymic},
		Grammemes: map[Kind][]string{KindBase: {"Name", "Surn", "Patr"}},
	}
	profPlace = Profile{
		Name:      "place",
		Kinds:     []Kind{KindBase, kindToponym, KindAbbrev},
		Grammemes: map[Kind][]string{KindBase: {"Geox"}},
	}
)

// fakeDicts returns readings by word form; Reading.Kind is the dictionary kind.
// It may return predicted readings of any kind: the Analyzer must drop the
// non-base ones itself (decision D11).
type fakeDicts map[string][]Reading

func (f fakeDicts) Parse(word string, kinds []Kind) []Reading {
	var out []Reading

	for _, r := range f[word] {
		if len(kinds) == 0 || slices.Contains(kinds, r.Kind) {
			out = append(out, r)
		}
	}

	return out
}

func (f fakeDicts) Version() string { return "fake-1" }

func rBase(normal, tag string) Reading {
	return Reading{Normal: normal, Tag: tag, Kind: KindBase, Dict: "base.test"}
}

func rBasePredicted(normal, tag string) Reading {
	r := rBase(normal, tag)
	r.Predicted = true

	return r
}

func rAbbrev(normal string) Reading {
	return Reading{Normal: normal, Tag: "NOUN", Kind: KindAbbrev, Dict: "abbrev.test"}
}

// fake is the genodex P1 test vocabulary plus v0.1 additions (и, с, ц., г.).
var fake = fakeDicts{
	"кот":         {rBase("кот", "NOUN,anim,masc sing,nomn")},
	"кота":        {rBase("кот", "NOUN,anim,masc sing,gent")},
	"пса":         {rBase("пёс", "NOUN,anim,masc sing,gent")},
	"стали":       {rBase("сталь", "NOUN,inan,femn sing,gent"), rBase("стать", "VERB,perf,intr plur,past,indc")},
	"в":           {rBase("в", "PREP")},
	"и":           {rBase("и", "CONJ")},
	"с":           {rBase("с", "PREP"), rAbbrev("село")},
	"селе":        {rBase("село", "NOUN,inan,neut sing,loct")},
	"с.":          {rAbbrev("село"), rAbbrev("сын")},
	"ц.":          {rAbbrev("церковь")},
	"г.":          {rAbbrev("город"), rAbbrev("год")},
	"кр-нин":      {rAbbrev("крестьянин")},
	"петербург":   {rBase("петербург", "NOUN,inan,masc,Geox sing,nomn")},
	"губ":         {rBase("губа", "NOUN,inan,femn plur,gent"), rAbbrev("губерния")},
	"дорожкиной":  {rBasePredicted("дорожкина", "NOUN,anim,femn sing,gent")},
	"кузнецова":   {rBase("кузнецов", "NOUN,anim,masc,Surn sing,gent"), rBasePredicted("кузнецова", "ADJF")},
	"покровского": {rBase("покровский", "ADJF,Geox masc,sing,gent")},
	"ивана":       {rBase("иван", "NOUN,anim,masc,Name sing,gent")},
	"дер": {
		rAbbrev("деревня"),
		{Normal: "дерь", Tag: "NOUN", Kind: KindAbbrev, Dict: "abbrev.test", Predicted: true},
	},
	"мороза": {
		{Normal: "мороз", Tag: "NOUN,anim,masc,Surn sing,gent", Kind: kindSurname, Dict: "surname.test", Predicted: true},
	},
}
