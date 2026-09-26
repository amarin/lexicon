package ner

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"maps"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/gazetteer"
	"github.com/amarin/lexicon/rules"
)

// extractorVersion identifies the extraction algorithm; it is part of
// Result.Version. Bump it on any change of extraction behaviour (candidate
// generation, filters, rule application, scoring, resolution, output flags)
// so hosts invalidate spans stored from an older build.
const extractorVersion = "ner-2"

// Pipeline extracts spans; it is safe for concurrent use.
type Pipeline struct {
	cfg      Config
	weights  Weights
	minRunes int
	book     *rules.Book
	nesting  map[string]map[string]bool
	cfgHash  string
}

// New validates cfg and applies defaults. It copies the host-owned maps
// (Profiles, Nesting, Weights.Types), so later edits to them do not affect
// the Pipeline.
func New(cfg Config) (*Pipeline, error) {
	switch {
	case cfg.Analyzer == nil:
		return nil, errors.New("ner: Config.Analyzer is nil")
	case cfg.Gazetteer == nil:
		return nil, errors.New("ner: Config.Gazetteer is nil")
	}
	cfg.Profiles = cloneProfiles(cfg.Profiles)
	cfg.Nesting = cloneNesting(cfg.Nesting)
	cfg.Weights.Types = maps.Clone(cfg.Weights.Types)
	if _, ok := cfg.Profiles[cfg.DefaultProfile]; !ok {
		return nil, fmt.Errorf("ner: default profile %q is not in Config.Profiles", cfg.DefaultProfile)
	}
	p := &Pipeline{cfg: cfg, weights: cfg.Weights, minRunes: cfg.MinLemmaMatchRunes, book: cfg.Rules}
	if w := p.weights; w.Surface == 0 && w.Lemma == 0 && w.Trigger == 0 {
		types := w.Types
		p.weights = DefaultWeights()
		p.weights.Types = types
	}
	if p.minRunes == 0 {
		p.minRunes = 3
	}
	if p.book == nil {
		b, err := rules.Compile()
		if err != nil {
			return nil, err
		}
		p.book = b
	}
	p.nesting = map[string]map[string]bool{}
	for outer, inners := range cfg.Nesting {
		m := map[string]bool{}
		for _, in := range inners {
			m[in] = true
		}
		p.nesting[outer] = m
	}
	// encoding/json writes map keys sorted, so the encoding is deterministic.
	data, err := json.Marshal(struct {
		W Weights
		M int
		N map[string][]string
		P map[string]lexicon.Profile
		D string
	}{p.weights, p.minRunes, cfg.Nesting, cfg.Profiles, cfg.DefaultProfile})
	if err != nil {
		return nil, fmt.Errorf("ner: %w", err)
	}
	sum := sha256.Sum256(data)
	p.cfgHash = hex.EncodeToString(sum[:8])
	return p, nil
}

// version combines every component version for Result.Version.
func (p *Pipeline) version(snap *gazetteer.Snapshot) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s\x00%s\x00%s\x00%s\x00%s", extractorVersion, p.cfg.Analyzer.Version(), snap.Version(), p.book.Version(), p.cfgHash)
	return hex.EncodeToString(h.Sum(nil))
}
