package nertest

// runConfig is what the Options of one Run set.
type runConfig struct {
	tags func(Case) []string // WithTags; nil = Case.Tags only
}
