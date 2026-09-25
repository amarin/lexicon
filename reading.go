package lexicon

// Reading is one parse of a word by one dictionary.
type Reading struct {
	Normal    string // lemma as stored in the dictionary (may contain ё)
	Tag       string // gomorphy tag, the first grammeme is the part of speech
	Kind      Kind   // kind of the dictionary that produced the reading
	Dict      string // name of that dictionary
	Predicted bool   // predicted from the word ending: the word is not in the dictionary
}
