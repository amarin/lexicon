package gazetteer

import "github.com/amarin/lexicon"

// Config configures a Gazetteer.
type Config struct {
	// Analyzer analyzes aliases (ModeFull). Its Version is part of every
	// compiled source; a change triggers recompilation on Refresh.
	Analyzer Analyzer
	// TypeProfiles selects the profile for aliases of an entry type.
	TypeProfiles map[string]lexicon.Profile
	// DefaultProfile is used for types missing from TypeProfiles.
	DefaultProfile lexicon.Profile
	// Sources are compiled in order; names must be unique and non-empty.
	Sources []Source
}
