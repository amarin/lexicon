package textnorm

// Token is a word, number, punctuation or symbol of the input. Offsets refer
// to the input string exactly as passed to Tokenize: input[Start:End] == Raw.
type Token struct {
	Raw         string    // input[Start:End]
	Form        string    // words: NormalizeWord(Raw); numbers: the digits; "" otherwise
	Start, End  int       // byte offsets in the input
	RuneStart   int       // code-point offset of Start
	RuneEnd     int       // code-point offset of End
	Kind        TokenKind // word, number, punct, symbol
	Script      Script    // words only; ScriptNone otherwise
	Case        Case      // words only; CaseNone otherwise
	Dotted      bool      // word immediately followed by '.' (abbreviation candidate)
	SentenceEnd bool      // '.', '!', '?', ';', '…' token, or last token before a blank line / U+2029
}
