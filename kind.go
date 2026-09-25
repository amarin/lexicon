package lexicon

import (
	"fmt"
	"strings"
)

// Kind is the kind of a dictionary: the prefix of its name "<kind>.<name>".
// The library defines KindBase and KindAbbrev; every other kind matching
// [a-z][a-z0-9_]* is host-defined (surname, given, toponym …) and selected
// by Profiles.
type Kind string

const (
	KindBase   Kind = "base"   // general morphology (OpenCorpora); the only kind whose predictions are used
	KindAbbrev Kind = "abbrev" // abbreviations: form (with its dot or hyphen) → full word lemma
)

// Valid reports whether k matches [a-z][a-z0-9_]*.
func (k Kind) Valid() bool {
	if k == "" {
		return false
	}

	for i, r := range k {
		switch {
		case r >= 'a' && r <= 'z':
		case i > 0 && (r >= '0' && r <= '9' || r == '_'):
		default:
			return false
		}
	}

	return true
}

// kindOf returns the kind of a dictionary named "<kind>.<name>".
func kindOf(name string) (Kind, error) {
	prefix, rest, ok := strings.Cut(name, ".")
	if k := Kind(prefix); ok && rest != "" && k.Valid() {
		return k, nil
	}

	return "", fmt.Errorf("lexicon: dictionary name %q: want <kind>.<name> with kind [a-z][a-z0-9_]*", name)
}
