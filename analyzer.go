package lexicon

import "github.com/amarin/lexicon/textnorm"

// analyzerVersion changes whenever analysis rules of this package change the
// produced terms (hosts reindex).
const analyzerVersion = "1"

// Analyzer turns text into terms with lemmas: textnorm tokens, abbreviations,
// lemmas from Dictionaries filtered by a Profile, stop words. One Analyzer
// serves index building, query parsing and NER, so they cannot diverge. Safe
// for concurrent use.
type Analyzer struct {
	dicts Dictionaries
	rules textnorm.Rules
}

// NewAnalyzer returns an analyzer over d with orthography rules r.
func NewAnalyzer(d Dictionaries, r textnorm.Rules, opts AnalyzerOptions) *Analyzer {
	return &Analyzer{dicts: d, rules: r}
}

// Analyze returns the terms of text under profile p: in ModeIndex the index
// terms (genodex P1), otherwise exactly one term per token.
func (a *Analyzer) Analyze(text string, p Profile, m Mode) []Term {
	toks := textnorm.Tokenize(a.rules, text)

	if m != ModeIndex {
		out := make([]Term, len(toks))
		for i, tok := range toks {
			out[i] = a.fullTerm(tok, p)
		}

		return out
	}

	var out []Term

	for _, tok := range toks {
		out = append(out, a.indexTerms(tok, p)...)
	}

	return out
}

// Version identifies everything that determines terms: analyzer rules,
// orthography rule set and the dictionary set. Hosts store it with derived
// data and recompute on change.
func (a *Analyzer) Version() string {
	return "analyzer-" + analyzerVersion + "/" + a.rules.Name + "-" + a.rules.Version + "/" + a.dicts.Version()
}
