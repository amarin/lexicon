// Package fakedict is an in-memory stand-in for lexicon.Dictionaries with a
// small Russian genealogy vocabulary. Tests of gazetteer, ner, nertest and
// the CLI use it so that they never need the OpenCorpora base dictionary.
package fakedict

import (
	"slices"
	"strings"

	"github.com/amarin/lexicon"
)

// Dict maps lower-cased word forms to readings. It is safe for concurrent
// reads once built.
type Dict struct {
	version string
	exact   map[string][]lexicon.Reading
	guessed map[string][]lexicon.Reading
}

// New returns an empty dictionary reporting version.
func New(version string) *Dict {
	return &Dict{
		version: version,
		exact:   map[string][]lexicon.Reading{},
		guessed: map[string][]lexicon.Reading{},
	}
}

// Add registers lemma for every listed form (list the lemma itself when it
// is a form) in dictionary dict of the given kind.
func (d *Dict) Add(kind lexicon.Kind, dict, lemma, tag string, forms ...string) *Dict {
	for _, f := range forms {
		k := strings.ToLower(f)
		d.exact[k] = append(d.exact[k], lexicon.Reading{Normal: lemma, Tag: tag, Kind: kind, Dict: dict})
	}
	return d
}

// Predict registers a predicted base-dictionary reading for form. Parse
// returns it only when no exact reading exists and the base kind is asked
// for, mirroring the Registry rule of the spec.
func (d *Dict) Predict(form, lemma, tag string) *Dict {
	k := strings.ToLower(form)
	d.guessed[k] = append(d.guessed[k], lexicon.Reading{
		Normal: lemma, Tag: tag, Kind: lexicon.KindBase, Dict: "base.fake", Predicted: true,
	})
	return d
}

// Parse implements lexicon.Dictionaries.
func (d *Dict) Parse(word string, kinds []lexicon.Kind) []lexicon.Reading {
	k := strings.ToLower(word)
	if out := filterKinds(d.exact[k], kinds); len(out) > 0 {
		return out
	}
	if !hasKind(kinds, lexicon.KindBase) {
		return nil
	}
	return filterKinds(d.guessed[k], kinds)
}

// Version implements lexicon.Dictionaries.
func (d *Dict) Version() string { return d.version }

func filterKinds(rs []lexicon.Reading, kinds []lexicon.Kind) []lexicon.Reading {
	var out []lexicon.Reading
	for _, r := range rs {
		if hasKind(kinds, r.Kind) {
			out = append(out, r)
		}
	}
	return out
}

// hasKind reports whether k should be consulted for kinds: an empty kinds
// list means "all kinds", like the real Registry (v0.1 D20).
func hasKind(kinds []lexicon.Kind, k lexicon.Kind) bool {
	return len(kinds) == 0 || slices.Contains(kinds, k)
}
