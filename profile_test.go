package lexicon

import "testing"

// TestProfileFilter: grammeme lists filter readings of their kind only; an
// empty list or an absent kind keeps everything.
func TestProfileFilter(t *testing.T) {
	p := Profile{Name: "name", Grammemes: map[Kind][]string{KindBase: {"Name", "Surn", "Patr"}, "toponym": {}}}
	rs := []Reading{
		{Normal: "кузнецов", Tag: "NOUN,anim,masc,Surn sing,nomn", Kind: KindBase},
		{Normal: "сталь", Tag: "NOUN,inan,femn sing,gent", Kind: KindBase},
		{Normal: "петров", Tag: "NOUN", Kind: "surname"},
		{Normal: "лягушкино", Tag: "NOUN", Kind: "toponym"},
	}

	var got []string
	for _, r := range p.filter(rs) {
		got = append(got, r.Normal)
	}

	if len(got) != 3 || got[0] != "кузнецов" || got[1] != "петров" || got[2] != "лягушкино" {
		t.Fatalf("filter = %v", got)
	}

	if len((Profile{}).filter(rs)) != len(rs) {
		t.Fatal("a profile without grammemes filtered readings")
	}
}
