package lexicon

import "context"

// StateStore persists the enabled/disabled state of dictionaries by name for
// the host. A dictionary without a record is enabled.
//
// Tests use hand-written fakes (no mockgen, decision D24).
type StateStore interface {
	Enabled(ctx context.Context) (map[string]bool, error)
	SetEnabled(ctx context.Context, name string, on bool) error
}

// Dictionaries parses words for the Analyzer; *Registry implements it. Parse
// returns the readings of word from dictionaries of the given kinds (empty =
// all), in registry order. Version identifies the current dictionary set; the
// Analyzer drops its lemma cache when it changes.
type Dictionaries interface {
	Parse(word string, kinds []Kind) []Reading
	Version() string
}
