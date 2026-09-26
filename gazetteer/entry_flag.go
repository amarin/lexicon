package gazetteer

import (
	"fmt"
	"strings"
)

// EntryFlag modifies how an entry is matched and resolved.
type EntryFlag uint8

const (
	// RequiresContext keeps a match only when a hint, trigger or pattern supports it.
	RequiresContext EntryFlag = 1 << iota
	// SurfaceOnly compiles only the surface key (no lemma key).
	SurfaceOnly
	// CaseSensitive keeps a match only when every word has the alias's letter case.
	CaseSensitive
	// Blocked vetoes every match of the entry's type on the matched range.
	// The veto applies whatever the entry's other flags say: a blocked
	// CaseSensitive entry vetoes regardless of letter case, and a blocked
	// one-word lemma match vetoes even when the alias is shorter than
	// ner's MinLemmaMatchRunes.
	Blocked
)

var entryFlagNames = [...]struct {
	flag EntryFlag
	name string
}{
	{RequiresContext, "requires_context"},
	{SurfaceOnly, "surface_only"},
	{CaseSensitive, "case_sensitive"},
	{Blocked, "blocked"},
}

// Has reports whether every flag in x is set.
func (f EntryFlag) Has(x EntryFlag) bool { return f&x == x }

// String returns the comma-separated flag names used in TSV files.
func (f EntryFlag) String() string {
	var parts []string
	for _, n := range entryFlagNames {
		if f.Has(n.flag) {
			parts = append(parts, n.name)
		}
	}
	return strings.Join(parts, ",")
}

// ParseEntryFlags parses comma-separated flag names; empty items are ignored.
func ParseEntryFlags(s string) (EntryFlag, error) {
	var f EntryFlag
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		found := false
		for _, n := range entryFlagNames {
			if n.name == p {
				f |= n.flag
				found = true
				break
			}
		}
		if !found {
			return 0, fmt.Errorf("unknown entry flag %q", p)
		}
	}
	return f, nil
}
