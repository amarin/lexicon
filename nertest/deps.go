package nertest

import (
	"context"

	"github.com/amarin/lexicon/ner"
)

// Extractor is what Run evaluates. *ner.Pipeline implements it.
type Extractor interface {
	Extract(ctx context.Context, d ner.Doc, o ...ner.Option) (ner.Result, error)
}
