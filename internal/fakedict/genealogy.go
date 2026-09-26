package fakedict

import "github.com/amarin/lexicon"

// Host-defined dictionary kinds used by the fixture.
const (
	KindGiven      lexicon.Kind = "given"
	KindSurname    lexicon.Kind = "surname"
	KindPatronymic lexicon.Kind = "patronymic"
	KindToponym    lexicon.Kind = "toponym"
)

// Genealogy returns the vocabulary shared by lexicon tests: service words,
// genealogy nouns, abbreviations, names, surnames and toponyms. Tags follow
// OpenCorpora conventions (Name, Surn, Patr, Geox).
func Genealogy() *Dict {
	const base, abbr = "base.fake", "abbrev.fake"
	b, a := lexicon.KindBase, lexicon.KindAbbrev
	d := New("fake-genealogy-1")

	d.Add(b, base, "деревня", "NOUN,inan,femn sing,nomn", "деревня", "деревни", "деревне", "деревню", "деревней")
	d.Add(b, base, "село", "NOUN,inan,neut sing,nomn", "село", "села", "селе")
	d.Add(b, base, "уезд", "NOUN,inan,masc sing,nomn", "уезд", "уезда", "уезде")
	d.Add(b, base, "улица", "NOUN,inan,femn sing,nomn", "улица", "улицы", "улице")
	d.Add(b, base, "мир", "NOUN,inan,masc sing,nomn", "мир", "мира", "миру")
	d.Add(b, base, "мороз", "NOUN,inan,masc sing,nomn", "мороз", "мороза")
	d.Add(b, base, "ус", "NOUN,inan,masc sing,nomn", "ус", "уса", "усы")
	d.Add(b, base, "вера", "NOUN,inan,femn sing,nomn", "вера", "веры")
	d.Add(b, base, "крестьянин", "NOUN,anim,masc sing,nomn", "крестьянин", "крестьянина", "крестьянину")
	d.Add(b, base, "сын", "NOUN,anim,masc sing,nomn", "сын", "сына")
	d.Add(b, base, "дочь", "NOUN,anim,femn sing,nomn", "дочь", "дочери")
	d.Add(b, base, "год", "NOUN,inan,masc sing,nomn", "год", "года", "лет")
	d.Add(b, base, "жить", "VERB,impf,intr", "жил", "жила")
	d.Add(b, base, "родиться", "VERB,perf,intr", "родился", "родилась")
	d.Add(b, base, "в", "PREP", "в")
	d.Add(b, base, "на", "PREP", "на")
	d.Add(b, base, "иван", "NOUN,anim,masc,Name sing,nomn", "иван", "ивана", "ивану")
	d.Add(b, base, "мария", "NOUN,anim,femn,Name sing,nomn", "мария", "марии")
	d.Add(b, base, "петров", "NOUN,anim,masc,Surn sing,nomn", "петров", "петрова", "петрову")
	d.Predict("сидоровке", "сидоровка", "NOUN,inan,femn,Geox sing,loct")

	d.Add(a, abbr, "деревня", "NOUN,inan,femn", "деревня", "дер", "дер.")
	d.Add(a, abbr, "улица", "NOUN,inan,femn", "улица", "ул", "ул.")
	d.Add(a, abbr, "крестьянин", "NOUN,anim,masc", "крестьянин", "крест", "крест.")
	d.Add(a, abbr, "село", "NOUN,inan,neut", "село", "с", "с.")
	d.Add(a, abbr, "сын", "NOUN,anim,masc", "сын", "с", "с.")
	d.Add(a, abbr, "погост", "NOUN,inan,masc", "погост", "пог", "пог.")

	d.Add(KindGiven, "given.fake", "иван", "Name,masc", "иван", "ивана", "ивану")
	d.Add(KindGiven, "given.fake", "иоанн", "Name,masc", "иоанн", "иоанна")
	d.Add(KindGiven, "given.fake", "мария", "Name,femn", "мария", "марии")
	d.Add(KindGiven, "given.fake", "вера", "Name,femn", "вера", "веры")
	d.Add(KindGiven, "given.fake", "евдокия", "Name,femn", "евдокия", "евдокии")
	d.Add(KindGiven, "given.fake", "авдотья", "Name,femn", "авдотья", "авдотьи")
	d.Add(KindSurname, "surname.fake", "петров", "Surn,masc", "петров", "петрова", "петрову")
	d.Add(KindSurname, "surname.fake", "мороз", "Surn", "мороз", "мороза")
	d.Add(KindPatronymic, "patronymic.fake", "петрович", "Patr,masc", "петрович", "петровича")
	d.Add(KindToponym, "toponym.fake", "лягушкино", "Geox,neut", "лягушкино")
	d.Add(KindToponym, "toponym.fake", "лягушкина", "Geox,femn", "лягушкина", "лягушкиной", "лягушкину")
	d.Add(KindToponym, "toponym.fake", "боровский", "Geox,masc", "боровский", "боровского", "боровском")
	return d
}
