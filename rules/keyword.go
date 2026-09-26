package rules

import (
	"fmt"
	"strings"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/textnorm"
)

// keyword is a compiled "a|b|c" lemma alternation. Alternatives are always
// normalized with NormalizeWord(PreReform, …), regardless of the rules an
// analyzer used to produce the Term being matched, so that a rule keyword
// written with modern spelling still matches pre-reform forms and vice
// versa.
type keyword struct {
	alts   []string
	dotted bool
}

func newKeyword(spec string, dotted bool) (keyword, error) {
	k := keyword{dotted: dotted}
	for _, p := range strings.Split(spec, "|") {
		p = strings.TrimSuffix(strings.TrimSpace(p), ".")
		if p == "" {
			continue
		}
		k.alts = append(k.alts, textnorm.NormalizeWord(textnorm.PreReform, p))
	}
	if len(k.alts) == 0 {
		return keyword{}, fmt.Errorf("empty lemma %q", spec)
	}
	return k, nil
}

// match reports whether t's form or one of its known lemmas is an
// alternative. In v0.1 (lexicon.Analyzer.fullTerm), a dotted word that
// resolves as an abbreviation gets Term.Form = tok.Form + "." (e.g. "ул.");
// alternatives never carry that dot, so it is trimmed before comparing.
func (k keyword) match(t *lexicon.Term) bool {
	if k.dotted && !t.Token.Dotted {
		return false
	}
	form := strings.TrimSuffix(t.Form, ".")
	for _, a := range k.alts {
		if form == a {
			return true
		}
		for i := range t.Lemmas {
			if l := &t.Lemmas[i]; l.Flags&lexicon.FlagUnknown == 0 && l.Text == a {
				return true
			}
		}
	}
	return false
}
