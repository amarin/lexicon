package ner

import (
	"slices"

	"github.com/amarin/lexicon/gazetteer"
)

// candidate is a possible span over content positions [start, end).
type candidate struct {
	start, end int
	typ        string
	origin     origin
	hits       []hit    // gazetteer aliases; empty for trigger candidates
	normals    []string // normal forms of candidates without hits
	bonus      float64  // hint/trigger/pattern evidence
	context    bool     // supported by a hint, trigger or pattern
	flags      SpanFlag
	evidence   []string
	absorbed   []int // keyword positions a rule absorbed into the range
	removed    bool
	score      float64
	alts       []*candidate // other readings of this range (D14): resolution losers, v0.3 relabel sources
}

func (c *candidate) words() int { return c.end - c.start }

func (c *candidate) addHit(a *gazetteer.Alias, k gazetteer.MatchKind) {
	for i := range c.hits {
		if c.hits[i].alias == a {
			c.hits[i].surface = c.hits[i].surface || k == gazetteer.BySurface
			return
		}
	}
	c.hits = append(c.hits, hit{alias: a, surface: k == gazetteer.BySurface})
}

// refs returns the sorted distinct non-empty refs of the hits.
func (c *candidate) refs() []string {
	var out []string
	for _, h := range c.hits {
		if r := h.alias.Entry.Ref; r != "" && !slices.Contains(out, r) {
			out = append(out, r)
		}
	}
	slices.Sort(out)
	return out
}

// normalForms returns the sorted distinct canonical forms of the hits, or
// the candidate's own normal forms when it has no hits.
func (c *candidate) normalForms() []string {
	if len(c.hits) == 0 {
		return slices.Clone(c.normals)
	}
	var out []string
	for _, h := range c.hits {
		if n := h.alias.Entry.Canonical; n != "" && !slices.Contains(out, n) {
			out = append(out, n)
		}
	}
	slices.Sort(out)
	return out
}

// attrs merges the hits' attributes; keys with conflicting values are dropped.
func (c *candidate) attrs() map[string]string {
	out := map[string]string{}
	conflict := map[string]bool{}
	for _, h := range c.hits {
		for k, v := range h.alias.Entry.Attrs {
			if conflict[k] {
				continue
			}
			if old, ok := out[k]; ok && old != v {
				delete(out, k)
				conflict[k] = true
				continue
			}
			out[k] = v
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// readings is the number of competing refs or normal forms (at least 1).
func (c *candidate) readings() int {
	return max(len(c.refs()), len(c.normalForms()), 1)
}

// recomputeOrigin sets origin from the surviving hits: originSurface if any
// of them matched by surface, originLemma otherwise. A no-op for candidates
// without hits (trigger candidates keep the origin they were created with).
func (c *candidate) recomputeOrigin() {
	if len(c.hits) == 0 {
		return
	}
	c.origin = originLemma
	for _, h := range c.hits {
		if h.surface {
			c.origin = originSurface
			return
		}
	}
}
