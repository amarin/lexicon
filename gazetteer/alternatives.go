package gazetteer

import (
	"slices"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/textnorm"
)

// IsContent reports whether t takes part in keys and matching (words and
// numbers; punctuation and symbols only separate them).
func IsContent(t *lexicon.Term) bool {
	return t.Token.Kind == textnorm.TokenWord || t.Token.Kind == textnorm.TokenNumber
}

// Alternatives returns the lemma texts t can match by: every lemma without
// the Unknown flag, de-duplicated in order, or the normalized form when there
// is none. It returns nil for non-content terms.
func Alternatives(t *lexicon.Term) []string {
	if !IsContent(t) {
		return nil
	}
	var out []string
	for _, l := range t.Lemmas {
		if l.Flags&lexicon.FlagUnknown != 0 || l.Text == "" {
			continue
		}
		if !slices.Contains(out, l.Text) {
			out = append(out, l.Text)
		}
	}
	if len(out) == 0 && t.Form != "" {
		out = append(out, t.Form)
	}
	return out
}
