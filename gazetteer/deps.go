package gazetteer

import (
	"context"

	"github.com/amarin/lexicon"
)

// Analyzer turns alias text into terms. *lexicon.Analyzer implements it.
type Analyzer interface {
	Analyze(text string, p lexicon.Profile, m lexicon.Mode) []lexicon.Term
	Version() string
}

// Source yields dictionary entries. Hosts implement it over their storage;
// TSVSource and SliceSource are provided.
type Source interface {
	// Name identifies the source; it must be unique within a Gazetteer.
	Name() string
	// Version must be cheap; an unchanged version skips recompilation.
	Version(ctx context.Context) (string, error)
	// Entries calls yield for every entry and stops at the first yield error.
	Entries(ctx context.Context, yield func(Entry) error) error
}
