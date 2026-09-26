package gazetteer

import "github.com/amarin/lexicon/textnorm"

// Alias is a compiled entry as seen by matching. Aliases are shared between
// matches and must be treated as read-only.
type Alias struct {
	Entry      Entry           // the source entry; Canonical is never empty
	Source     string          // name of the source that produced it
	Cases      []textnorm.Case // letter case of each word of Entry.Alias
	LemmaKeys  []string        // space-joined lemma keys (empty for SurfaceOnly)
	SurfaceKey string          // space-joined normalized forms
}
