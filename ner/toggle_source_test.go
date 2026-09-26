package ner

import (
	"context"
	"sync/atomic"

	"github.com/amarin/lexicon/gazetteer"
)

// toggleSource serves one of two entry sets; flip switches between them.
type toggleSource struct {
	b    atomic.Bool
	a, x []gazetteer.Entry // entries while b is false / true
}

func (s *toggleSource) Name() string { return "toggle" }

func (s *toggleSource) Version(context.Context) (string, error) {
	if s.b.Load() {
		return "b", nil
	}
	return "a", nil
}

func (s *toggleSource) Entries(ctx context.Context, yield func(gazetteer.Entry) error) error {
	es := s.a
	if s.b.Load() {
		es = s.x
	}
	return gazetteer.NewSliceSource(s.Name(), "", es).Entries(ctx, yield)
}

func (s *toggleSource) flip() { s.b.Store(!s.b.Load()) }
