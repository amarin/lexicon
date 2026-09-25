package lexicon

import "fmt"

// kindSet is the set of host-declared kinds (Options.Kinds, owner decision 7,
// D17). A nil set accepts every valid kind; a non-empty set accepts
// KindBase, KindAbbrev and its members.
type kindSet map[Kind]bool

// newKindSet builds the set of declared kinds; nil/empty → nil (accept all).
func newKindSet(declared []Kind) (kindSet, error) {
	if len(declared) == 0 {
		return nil, nil
	}

	s := make(kindSet, len(declared))

	for _, k := range declared {
		if !k.Valid() {
			return nil, fmt.Errorf("lexicon: Options.Kinds: invalid kind %q", k)
		}

		s[k] = true
	}

	return s, nil
}

// accepts reports whether dictionaries of kind k take part in parsing.
func (s kindSet) accepts(k Kind) bool {
	return s == nil || k == KindBase || k == KindAbbrev || s[k]
}

// check returns an "unknown kind" error for a kind the set does not accept.
func (s kindSet) check(k Kind) error {
	if s.accepts(k) {
		return nil
	}

	return fmt.Errorf("unknown kind %q", k)
}
