package lexicon

// Lemma is a normal form of a word with flags and the reading that gave it.
type Lemma struct {
	Text  string // normal form, normalized like the word (textnorm.NormalizeWord)
	Flags Flag
	Tag   string // tag of the first reading that produced the lemma (grammemes for NER)
	Kind  Kind   // kind of that reading's dictionary ("" for unknown lemmas)
	Dict  string // name of that dictionary ("" for unknown lemmas)
}
