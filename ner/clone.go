package ner

import (
	"maps"
	"slices"

	"github.com/amarin/lexicon"
)

// cloneProfiles deep-copies host-owned profiles so later host edits cannot
// change a running Pipeline.
func cloneProfiles(in map[string]lexicon.Profile) map[string]lexicon.Profile {
	if in == nil {
		return nil
	}
	out := make(map[string]lexicon.Profile, len(in))
	for name, p := range in {
		out[name] = cloneProfile(p)
	}
	return out
}

func cloneProfile(p lexicon.Profile) lexicon.Profile {
	p.Kinds = slices.Clone(p.Kinds)
	if p.Grammemes != nil {
		g := make(map[lexicon.Kind][]string, len(p.Grammemes))
		for k, v := range p.Grammemes {
			g[k] = slices.Clone(v)
		}
		p.Grammemes = g
	}
	return p
}

// cloneNesting deep-copies Config.Nesting.
func cloneNesting(in map[string][]string) map[string][]string {
	out := maps.Clone(in)
	for k, v := range out {
		out[k] = slices.Clone(v)
	}
	return out
}
