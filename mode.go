package lexicon

// Mode selects what Analyze emits.
type Mode int

const (
	// ModeIndex: genodex search P1 behaviour — Cyrillic words and numbers only,
	// stop words dropped, compounds as whole + parts.
	ModeIndex Mode = iota
	// ModeFull: exactly one term per token — stop words flagged, Latin words,
	// numbers and punctuation kept (for NER patterns). Any value other than
	// ModeIndex behaves as ModeFull.
	ModeFull
)
