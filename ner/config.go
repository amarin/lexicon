package ner

import (
	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/rules"
)

// Config configures a Pipeline.
type Config struct {
	Analyzer  Analyzer
	Gazetteer Gazetteer
	Rules     *rules.Book // nil: no rules
	// Profiles maps Doc.Profile names to analyzer profiles.
	Profiles       map[string]lexicon.Profile
	DefaultProfile string
	// Nesting lists, per outer type, the inner types allowed inside it.
	Nesting map[string][]string
	// Weights: DefaultWeights when Surface, Lemma and Trigger are all zero.
	Weights Weights
	// MinLemmaMatchRunes drops one-word lemma-only matches of shorter aliases; 0 → 3.
	MinLemmaMatchRunes int
}
