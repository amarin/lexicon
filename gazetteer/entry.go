package gazetteer

// Entry is one alias of a host dictionary record.
type Entry struct {
	Alias     string            // surface text of the alias, e.g. «дер. Лягушкино», «СПб»
	Type      string            // host entity type: "surname", "division", "estate" …
	Ref       string            // opaque host key of the dictionary record; may be ""
	Canonical string            // normal form of the record; Alias when empty
	Attrs     map[string]string // passthrough (since/until of a rename, level, …)
	Flags     EntryFlag
}
