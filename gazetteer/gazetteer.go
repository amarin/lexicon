package gazetteer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/amarin/lexicon"
)

// Gazetteer owns compiled sources and publishes them as immutable snapshots.
// Readers call Snapshot (lock-free); Refresh and RefreshSource are
// serialized and swap the snapshot atomically.
type Gazetteer struct {
	b       builder
	sources []Source
	cfgHash string // profile configuration, part of every Snapshot.Version
	mu      sync.Mutex
	cur     atomic.Pointer[Snapshot]
}

// New validates cfg and compiles every source. Source failures do not fail
// New; they are visible in Snapshot().Reports(). New copies the profiles
// (TypeProfiles, DefaultProfile), so later host edits do not affect it.
func New(ctx context.Context, cfg Config) (*Gazetteer, error) {
	if cfg.Analyzer == nil {
		return nil, errors.New("gazetteer: Config.Analyzer is nil")
	}
	seen := map[string]bool{}
	for _, s := range cfg.Sources {
		n := s.Name()
		if n == "" || seen[n] {
			return nil, fmt.Errorf("gazetteer: empty or duplicate source name %q", n)
		}
		seen[n] = true
	}
	// encoding/json writes map keys sorted, so the encoding is deterministic.
	data, err := json.Marshal(struct {
		T map[string]lexicon.Profile
		D lexicon.Profile
	}{cfg.TypeProfiles, cfg.DefaultProfile})
	if err != nil {
		return nil, fmt.Errorf("gazetteer: %w", err)
	}
	sum := sha256.Sum256(data)
	g := &Gazetteer{
		b:       builder{analyzer: cfg.Analyzer, typeProfiles: cloneProfiles(cfg.TypeProfiles), defaultProfile: cloneProfile(cfg.DefaultProfile)},
		sources: slices.Clone(cfg.Sources),
		cfgHash: hex.EncodeToString(sum[:8]),
	}
	empty := &Snapshot{sources: make([]*compiledSource, len(g.sources))}
	for i, s := range g.sources {
		empty.sources[i] = &compiledSource{name: s.Name(), report: SourceReport{Source: s.Name()}}
	}
	empty.version = snapshotVersion(g.cfgHash, empty.sources)
	g.cur.Store(empty)
	if _, err := g.refresh(ctx, "", false); err != nil {
		return nil, err
	}
	return g, nil
}

// Snapshot returns the current snapshot. Hold on to it for the duration of
// one operation to see a consistent state.
func (g *Gazetteer) Snapshot() *Snapshot { return g.cur.Load() }

// Refresh asks every source for its version and recompiles those whose
// version (or the analyzer version) changed. It returns their reports.
func (g *Gazetteer) Refresh(ctx context.Context) ([]SourceReport, error) {
	return g.refresh(ctx, "", false)
}

// RefreshSource recompiles one source regardless of its version.
func (g *Gazetteer) RefreshSource(ctx context.Context, name string) (SourceReport, error) {
	reps, err := g.refresh(ctx, name, true)
	if err != nil {
		return SourceReport{}, err
	}
	return reps[0], nil
}

// Canonical delegates to the current snapshot.
func (g *Gazetteer) Canonical(key string) []string { return g.Snapshot().Canonical(key) }

// Expand delegates to the current snapshot.
func (g *Gazetteer) Expand(lemma string) []string { return g.Snapshot().Expand(lemma) }

func (g *Gazetteer) refresh(ctx context.Context, only string, force bool) ([]SourceReport, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	old := g.cur.Load()
	next := &Snapshot{sources: slices.Clone(old.sources)}
	av := g.b.analyzer.Version()
	var reps []SourceReport
	found := only == ""
	for i, src := range g.sources {
		if only != "" && src.Name() != only {
			continue
		}
		found = true
		if err := ctx.Err(); err != nil {
			return reps, err
		}
		prev := old.sources[i]
		ver, err := src.Version(ctx)
		if err != nil {
			kept := *prev
			kept.report.Err = fmt.Errorf("version: %w", err)
			next.sources[i] = &kept
			reps = append(reps, kept.report)
			continue
		}
		if !force && prev.built && prev.version == ver && prev.analyzerVersion == av {
			continue
		}
		next.sources[i] = g.b.build(ctx, src, ver, av, prev)
		reps = append(reps, next.sources[i].report)
	}
	if !found {
		return nil, fmt.Errorf("%w: %q", ErrUnknownSource, only)
	}
	if len(reps) > 0 {
		next.version = snapshotVersion(g.cfgHash, next.sources)
		g.cur.Store(next)
	}
	return reps, nil
}
