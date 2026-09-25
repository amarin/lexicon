package lexicon

import "github.com/amarin/lexicon/textnorm"

// Term is an analyzed token: the source token, the form looked up and its
// lemmas. For a compound part in ModeIndex, Token is the part token. For an
// abbreviation with a dot, Form includes the dot («с.»). Punctuation and
// symbol terms (ModeFull) have no lemmas.
type Term struct {
	Token  textnorm.Token
	Form   string
	Lemmas []Lemma
}
