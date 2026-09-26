package ner

import (
	"context"
	"fmt"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/gazetteer"
)

// Extract finds the spans of d.Text. It pins one gazetteer snapshot.
func (p *Pipeline) Extract(ctx context.Context, d Doc, opts ...Option) (Result, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	name := d.Profile
	if name == "" {
		name = p.cfg.DefaultProfile
	}
	prof, ok := p.cfg.Profiles[name]
	if !ok {
		return Result{}, fmt.Errorf("%w: %q", ErrUnknownProfile, name)
	}
	snap := p.cfg.Gazetteer.Snapshot()
	res := Result{Version: p.version(snap)}
	if d.Text == "" {
		return res, nil
	}
	terms := p.cfg.Analyzer.Analyze(d.Text, prof, lexicon.ModeFull)
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	st := &state{p: p, tx: gazetteer.Prepare(terms), terms: terms, explain: o.explain}
	st.fromMatches(snap.Match(st.tx, nil))
	st.filterEarly()
	st.filterContext()
	st.scoreAll()
	chosen, nested := st.resolve()
	res.Spans, _ = st.output(d.Text, chosen, nested, d.Types)
	return res, nil
}
