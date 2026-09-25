package lexicon

// OriginBuiltin is Entry.Origin of host-supplied built-in dictionaries.
const OriginBuiltin = "builtin"

// Entry describes a registry dictionary (List).
type Entry struct {
	Name     string   // "<kind>.<name>": file name without extension, or the built-in's name
	Kind     Kind     // name prefix
	Format   Format   // dat or tsv
	Origin   string   // OriginBuiltin or the file path
	Hash     string   // content hash: gomorphy ContentHash (dat) or sha256 of the text (tsv), hex
	Enabled  bool     // user state; a broken dictionary never takes part in parsing
	Manifest Manifest // provenance for attribution
	Error    string   // load error; "" when loaded
}
