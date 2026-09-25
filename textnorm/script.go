package textnorm

// Script is the letter script of a word token.
type Script uint8

const (
	ScriptNone     Script = iota // not a word
	ScriptCyrillic               // Cyrillic letters (Latin Rules.Homoglyphs of a Cyrillic word part included)
	ScriptLatin                  // Latin letters only
	ScriptMixed                  // other scripts, alone or mixed (Greek, Latin+Greek …)
)

var scriptNames = [...]string{"none", "cyrillic", "latin", "mixed"}

// String returns "none", "cyrillic", "latin" or "mixed".
func (s Script) String() string {
	if int(s) < len(scriptNames) {
		return scriptNames[s]
	}

	return "unknown"
}
