package textnorm

// Rules is an orthography rule set. Forms derived by a rule set are stored by
// hosts (search index), so any change of a table requires a new Version.
//
// The package-level Modern and PreReform values must not be modified; build a
// new Rules value to customise.
type Rules struct {
	// Name identifies the rule set, e.g. "modern", "prereform".
	Name string
	// Version is part of every derived version hash (Analyzer.Version).
	Version string
	// Letters replaces letters after lower-casing and canonical decomposition
	// (ѣ→е, і→и, ѳ→ф …). A replacement may be several letters (ѯ→кс).
	Letters map[rune]string
	// KeepMarks lists precomposed letters kept intact although they decompose
	// into a letter and a combining mark (й).
	KeepMarks []rune
	// DropFinalHardSign removes a final ъ from every hyphen part of a word
	// («Ивановъ» → «иванов»); applied by NormalizeWord only.
	DropFinalHardSign bool
	// Homoglyphs maps Latin letters that look like Cyrillic ones to those
	// Cyrillic letters (OCR and typing errors: «Kот», «Iоаннъ»). A word part
	// (a maximal run of letters and combining marks, so hyphens separate
	// parts) that has a Cyrillic letter and no other letters than keys of
	// Homoglyphs is Cyrillic: the tokenizer keeps it one ScriptCyrillic word
	// and Orthography replaces its homoglyphs. Parts with other Latin letters
	// («Kотw») or without Cyrillic letters («XIX») are left alone.
	Homoglyphs map[rune]rune
}
