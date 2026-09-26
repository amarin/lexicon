package gazetteer

import (
	"slices"

	"github.com/amarin/lexicon"
)

// cloneProfiles deep-copies host-owned profiles so later host edits cannot
// change a running Gazetteer.
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
