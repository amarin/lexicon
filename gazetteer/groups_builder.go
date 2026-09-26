package gazetteer

import "slices"

// groupsBuilder collects variant groups: aliases of one source sharing a
// non-empty Ref form a group. Every alias key also indexes its canonical form.
type groupsBuilder struct {
	byRef     map[string]int
	keys      []map[string]bool
	groupsOf  map[string]map[int]bool
	canonical map[string]map[string]bool
}

func newGroupsBuilder() *groupsBuilder {
	return &groupsBuilder{
		byRef:     map[string]int{},
		groupsOf:  map[string]map[int]bool{},
		canonical: map[string]map[string]bool{},
	}
}

func (b *groupsBuilder) add(a *Alias) {
	for _, k := range append(slices.Clone(a.LemmaKeys), a.SurfaceKey) {
		addToSet(b.canonical, k, a.Entry.Canonical)
	}
	if a.Entry.Ref == "" || len(a.LemmaKeys) == 0 {
		return
	}
	g, ok := b.byRef[a.Entry.Ref]
	if !ok {
		g = len(b.keys)
		b.byRef[a.Entry.Ref] = g
		b.keys = append(b.keys, map[string]bool{})
	}
	for _, k := range a.LemmaKeys {
		b.keys[g][k] = true
		if b.groupsOf[k] == nil {
			b.groupsOf[k] = map[int]bool{}
		}
		b.groupsOf[k][g] = true
	}
}

func (b *groupsBuilder) freeze() *groups {
	g := &groups{
		canonical: make(map[string][]string, len(b.canonical)),
		groupsOf:  make(map[string][]int, len(b.groupsOf)),
		keys:      make([][]string, len(b.keys)),
	}
	for k, set := range b.canonical {
		g.canonical[k] = sortedKeys(set)
	}
	for k, set := range b.groupsOf {
		ids := make([]int, 0, len(set))
		for id := range set {
			ids = append(ids, id)
		}
		slices.Sort(ids)
		g.groupsOf[k] = ids
	}
	for i, set := range b.keys {
		g.keys[i] = sortedKeys(set)
	}
	return g
}

func addToSet(m map[string]map[string]bool, k, v string) {
	if m[k] == nil {
		m[k] = map[string]bool{}
	}
	m[k][v] = true
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}
