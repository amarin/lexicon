package nertest

import (
	"fmt"
	"io"
	"slices"
	"text/tabwriter"
)

// Report aggregates a Run.
type Report struct {
	Cases    int
	Types    map[string]*TypeScore
	Failures []Failure
}

func (r *Report) typeScore(typ string) *TypeScore {
	if r.Types[typ] == nil {
		r.Types[typ] = &TypeScore{}
	}
	return r.Types[typ]
}

func (r *Report) typeNames() []string {
	names := make([]string, 0, len(r.Types))
	for t := range r.Types {
		names = append(names, t)
	}
	slices.Sort(names)
	return names
}

// Check lists every type whose strict precision or recall is below the minimum.
func (r *Report) Check(minPrecision, minRecall float64) []string {
	var out []string
	for _, t := range r.typeNames() {
		s := r.Types[t].Strict
		if p := s.Precision(); p < minPrecision {
			out = append(out, fmt.Sprintf("%s: strict precision %.3f < %.3f", t, p, minPrecision))
		}
		if rc := s.Recall(); rc < minRecall {
			out = append(out, fmt.Sprintf("%s: strict recall %.3f < %.3f", t, rc, minRecall))
		}
	}
	return out
}

// Write prints the per-type table and the failures.
func (r *Report) Write(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintf(tw, "TYPE\tP\tR\tF1\tP~\tR~\tF1~\tTP\tFP\tFN\n")
	for _, t := range r.typeNames() {
		s, p := r.Types[t].Strict, r.Types[t].Partial
		fmt.Fprintf(tw, "%s\t%.3f\t%.3f\t%.3f\t%.3f\t%.3f\t%.3f\t%d\t%d\t%d\n",
			t, s.Precision(), s.Recall(), s.F1(), p.Precision(), p.Recall(), p.F1(), s.TP, s.FP, s.FN)
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	fmt.Fprintf(w, "cases: %d, failures: %d\n", r.Cases, len(r.Failures))
	for _, f := range r.Failures {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t«%s»\n", f.Kind, f.Case, f.Type, f.Text); err != nil {
			return err
		}
	}
	return nil
}

// score adds one case: strict (exact bytes and type), then partial
// (same type, overlapping; each gold span matched at most once).
func (r *Report) score(c Case, gold, pred []bounds) {
	used := make([]bool, len(gold))
	for _, p := range pred {
		ts := r.typeScore(p.typ)
		if i := firstUnused(gold, used, func(g bounds) bool { return g == p }); i >= 0 {
			used[i] = true
			ts.Strict.TP++
		} else {
			ts.Strict.FP++
			r.Failures = append(r.Failures, Failure{Case: c.ID, Kind: "spurious", Type: p.typ, Text: c.Text[p.start:p.end]})
		}
	}
	for i, g := range gold {
		if !used[i] {
			r.typeScore(g.typ).Strict.FN++
			r.Failures = append(r.Failures, Failure{Case: c.ID, Kind: "missed", Type: g.typ, Text: c.Text[g.start:g.end]})
		}
	}
	used = make([]bool, len(gold))
	for _, p := range pred {
		ts := r.typeScore(p.typ)
		overlap := func(g bounds) bool { return g.typ == p.typ && g.start < p.end && p.start < g.end }
		if i := firstUnused(gold, used, overlap); i >= 0 {
			used[i] = true
			ts.Partial.TP++
		} else {
			ts.Partial.FP++
		}
	}
	for i, g := range gold {
		if !used[i] {
			r.typeScore(g.typ).Partial.FN++
		}
	}
}

func firstUnused(gold []bounds, used []bool, ok func(bounds) bool) int {
	for i, g := range gold {
		if !used[i] && ok(g) {
			return i
		}
	}
	return -1
}
