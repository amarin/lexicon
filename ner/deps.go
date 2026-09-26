package ner

import (
	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/gazetteer"
)

// Analyzer analyzes document text. *lexicon.Analyzer implements it.
type Analyzer interface {
	Analyze(text string, p lexicon.Profile, m lexicon.Mode) []lexicon.Term
	Version() string
}

// Gazetteer provides the current compiled dictionaries. *gazetteer.Gazetteer implements it.
type Gazetteer interface {
	Snapshot() *gazetteer.Snapshot
}
