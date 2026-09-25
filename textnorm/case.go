package textnorm

// Case is the letter case pattern of a word token.
type Case uint8

const (
	CaseNone  Case = iota // no cased letters (numbers, punctuation, caseless scripts)
	CaseLower             // all lower case
	CaseTitle             // first letter upper, rest lower — or every hyphen part so («Санкт-Петербург»)
	CaseUpper             // all upper case, at least two letters
	CaseMixed             // anything else («СПб»)
)

var caseNames = [...]string{"none", "lower", "title", "upper", "mixed"}

// String returns "none", "lower", "title", "upper" or "mixed".
func (c Case) String() string {
	if int(c) < len(caseNames) {
		return caseNames[c]
	}

	return "unknown"
}
