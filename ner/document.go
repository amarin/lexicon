package ner

// Doc is one text to analyze.
type Doc struct {
	Text    string
	Profile string   // lexicon profile for the text; Config.DefaultProfile when empty
	Tags    []string // document tags selecting rule sets, e.g. "period:pre1917"
	Types   []string // optional output filter, applied after resolution
}
