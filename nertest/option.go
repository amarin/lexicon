package nertest

// Option configures Run.
type Option func(*runConfig)

// WithTags derives extra document tags from a case, typically from
// Case.Context; they are appended to Case.Tags for the extraction only (the
// case is not modified). Owner decision 2026-09-25.
func WithTags(f func(Case) []string) Option {
	return func(c *runConfig) { c.tags = f }
}
