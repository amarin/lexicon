package gazetteer

// groups is the frozen variant-group index of one source.
type groups struct {
	canonical map[string][]string // lemma or surface key → sorted canonical forms
	groupsOf  map[string][]int    // lemma key → group IDs
	keys      [][]string          // group ID → sorted lemma keys
}

func (g *groups) canonicalOf(key string) []string { return g.canonical[key] }

func (g *groups) expand(lemma string) []string {
	var out []string
	for _, id := range g.groupsOf[lemma] {
		out = append(out, g.keys[id]...)
	}
	return out
}
