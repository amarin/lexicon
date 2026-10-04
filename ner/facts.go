package ner

import (
	"fmt"
	"slices"
	"strings"
)

// buildFacts keeps facts whose argument candidates were output (index maps
// candidates to Result.Spans positions) and removes duplicates (decision P8).
func (s *state) buildFacts(index map[*candidate]int) []Fact {
	var out []Fact
	seen := map[string]bool{}
	for _, f := range s.facts {
		args := make(map[string]int, len(f.args))
		ok := true
		for role, c := range f.args {
			i, found := index[c]
			if !found {
				ok = false
				break
			}
			args[role] = i
		}
		if !ok {
			continue
		}
		if key := factKey(f.kind, args); !seen[key] {
			seen[key] = true
			out = append(out, Fact{Kind: f.kind, Args: args, Rule: f.rule})
		}
	}
	return out
}

func factKey(kind string, args map[string]int) string {
	roles := make([]string, 0, len(args))
	for r := range args {
		roles = append(roles, r)
	}
	slices.Sort(roles)
	var b strings.Builder
	b.WriteString(kind)
	for _, r := range roles {
		fmt.Fprintf(&b, "|%s=%d", r, args[r])
	}
	return b.String()
}
