package nertest

import (
	"context"
	"fmt"
	"slices"

	"github.com/amarin/lexicon/ner"
)

// Run extracts every case and scores the result. Case.Context is used only
// through WithTags.
func Run(ctx context.Context, ex Extractor, cases []Case, opts ...Option) (*Report, error) {
	var cfg runConfig
	for _, o := range opts {
		o(&cfg)
	}
	r := &Report{Types: map[string]*TypeScore{}}
	for _, c := range cases {
		gold, err := c.goldBounds()
		if err != nil {
			return nil, fmt.Errorf("nertest: case %s: %w", c.ID, err)
		}
		tags := c.Tags
		if cfg.tags != nil {
			tags = append(slices.Clip(c.Tags), cfg.tags(c)...)
		}
		res, err := ex.Extract(ctx, ner.Doc{Text: c.Text, Profile: c.Profile, Tags: tags})
		if err != nil {
			return nil, fmt.Errorf("nertest: case %s: %w", c.ID, err)
		}
		pred := make([]bounds, len(res.Spans))
		for i, s := range res.Spans {
			pred[i] = bounds{start: s.Start, end: s.End, typ: s.Type}
		}
		r.Cases++
		r.score(c, gold, pred)
	}
	return r, nil
}
