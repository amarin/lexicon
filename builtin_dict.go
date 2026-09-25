package lexicon

// BuiltinDict is a dictionary supplied by the host (e.g. embedded with
// //go:embed): abbreviations, name lists. A file with the same Name in
// Options.Dir replaces it. Data of a FormatDat built-in is opened in place
// (gomorphy OpenBytes) and must not be modified while the registry lives.
type BuiltinDict struct {
	Name     string // "<kind>.<name>"
	Kind     Kind   // must equal the name prefix; "" = take it from the name
	Format   Format
	Data     []byte
	Manifest Manifest
}
