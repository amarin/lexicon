package gazetteer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
)

// Snapshot is an immutable set of compiled sources. It is safe for
// concurrent use; Gazetteer publishes a new one on every successful refresh.
type Snapshot struct {
	sources []*compiledSource
	version string
}

// Match appends to out every match of every built source in tx and returns
// the extended slice. It does not allocate beyond growing out.
func (s *Snapshot) Match(tx *Text, out []Match) []Match {
	for _, src := range s.sources {
		if !src.built {
			continue
		}
		for p := 0; p < tx.Len(); p++ {
			out = walkLemma(src.lemma, tx, p, p, 0, out)
			out = walkSurface(src.surface, tx, p, out)
		}
	}
	return out
}

// walkLemma extends the match that started at start and reached trie node
// node, consuming position p through every lemma alternative of p.
func walkLemma(t *trie, tx *Text, start, p int, node uint32, out []Match) []Match {
	if p >= len(tx.pos) || p-start >= t.depth {
		return out
	}
	if p > start && tx.brk[p-1] {
		return out
	}
	for _, id := range tx.lemIDs[tx.lemOff[p]:tx.lemOff[p+1]] {
		c, ok := t.child(node, id)
		if !ok {
			continue
		}
		if as := t.terminal(c); len(as) > 0 {
			out = append(out, Match{Start: start, End: p + 1, Kind: ByLemma, Aliases: as})
		}
		out = walkLemma(t, tx, start, p+1, c, out)
	}
	return out
}

// Canonical returns the sorted canonical forms of every alias whose lemma
// key or surface key (space-joined) equals key.
func (s *Snapshot) Canonical(key string) []string {
	return s.unionSorted(func(g *groups) []string { return g.canonicalOf(key) })
}

// Expand returns the sorted lemma keys of every variant group (aliases of
// one source sharing a Ref) that contains the lemma key lemma, or nil.
// Search uses it to expand queries; NER uses the same groups via Canonical.
func (s *Snapshot) Expand(lemma string) []string {
	return s.unionSorted(func(g *groups) []string { return g.expand(lemma) })
}

func (s *Snapshot) unionSorted(get func(*groups) []string) []string {
	var out []string
	for _, src := range s.sources {
		if !src.built {
			continue
		}
		for _, v := range get(src.groups) {
			if !slices.Contains(out, v) {
				out = append(out, v)
			}
		}
	}
	slices.Sort(out)
	return out
}

// Version identifies the compiled content: it changes when any source is
// recompiled from a different version or with a different analyzer.
func (s *Snapshot) Version() string { return s.version }

// Reports returns the last report of every source, in configuration order.
func (s *Snapshot) Reports() []SourceReport {
	out := make([]SourceReport, len(s.sources))
	for i, src := range s.sources {
		out[i] = src.report
	}
	return out
}

func snapshotVersion(srcs []*compiledSource) string {
	h := sha256.New()
	for _, s := range srcs {
		if s.built {
			fmt.Fprintf(h, "%s\x00%s\x00%s\n", s.name, s.version, s.analyzerVersion)
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}

func walkSurface(t *trie, tx *Text, start int, out []Match) []Match {
	node := uint32(0)
	for p := start; p < len(tx.pos) && p-start < t.depth; p++ {
		if p > start && tx.brk[p-1] {
			break
		}
		c, ok := t.child(node, tx.formIDs[p])
		if !ok {
			break
		}
		node = c
		if as := t.terminal(c); len(as) > 0 {
			out = append(out, Match{Start: start, End: p + 1, Kind: BySurface, Aliases: as})
		}
	}
	return out
}
