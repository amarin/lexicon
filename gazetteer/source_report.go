package gazetteer

import "time"

// SourceReport describes the last compilation attempt of a source.
type SourceReport struct {
	Source      string
	Version     string        // version of the compiled data in use
	Entries     int           // entries yielded by the source
	Aliases     int           // entries compiled (blocked ones included)
	LemmaKeys   int           // lemma keys inserted
	SurfaceKeys int           // surface keys inserted
	Capped      int           // aliases whose lemma expansion hit MaxLemmaKeys
	Blocked     int           // entries flagged Blocked
	Skipped     int           // aliases without words
	Duration    time.Duration // compilation time
	Errors      []string      // non-fatal problems: bad lines, capped or empty aliases
	Err         error         // fatal: the source kept its previous compiled data
}
