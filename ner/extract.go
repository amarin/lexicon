package ner

import (
	"context"
	"fmt"
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
	if _, ok := p.cfg.Profiles[name]; !ok {
		return Result{}, fmt.Errorf("%w: %q", ErrUnknownProfile, name)
	}
	snap := p.cfg.Gazetteer.Snapshot()
	return Result{Version: p.version(snap)}, nil
}
