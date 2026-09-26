package gazetteer

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/textnorm"
)

// builder compiles sources with one analyzer and profile configuration.
type builder struct {
	analyzer       Analyzer
	typeProfiles   map[string]lexicon.Profile
	defaultProfile lexicon.Profile
}

// build compiles src. On a fatal error from Entries it returns a copy of
// prev with report.Err set, so matching keeps using the previous data.
func (b *builder) build(ctx context.Context, src Source, version, analyzerVersion string, prev *compiledSource) *compiledSource {
	start := time.Now()
	rep := SourceReport{Source: src.Name(), Version: version}
	lb, sb, gb := newTrieBuilder(), newTrieBuilder(), newGroupsBuilder()
	var aliases []*Alias
	err := src.Entries(ctx, func(e Entry) error {
		rep.Entries++
		if a := b.compile(e, src.Name(), &rep, lb, sb); a != nil {
			gb.add(a)
			aliases = append(aliases, a)
		}
		return nil
	})
	var pe *ParseErrors
	if err != nil && !errors.As(err, &pe) {
		kept := *prev
		kept.report = rep
		kept.report.Version = prev.version
		kept.report.Err = err
		kept.report.Duration = time.Since(start)
		return &kept
	}
	if pe != nil {
		for _, l := range pe.Lines {
			rep.Errors = append(rep.Errors, fmt.Sprintf("line %d: %s", l.Line, l.Msg))
		}
	}
	rep.Duration = time.Since(start)
	return &compiledSource{
		name:            src.Name(),
		version:         version,
		analyzerVersion: analyzerVersion,
		built:           true,
		lemma:           lb.freeze(),
		surface:         sb.freeze(),
		groups:          gb.freeze(),
		aliases:         aliases,
		report:          rep,
	}
}

// compile analyzes one entry and inserts its keys; nil when it has no words.
func (b *builder) compile(e Entry, source string, rep *SourceReport, lb, sb *trieBuilder) *Alias {
	if e.Flags.Has(Blocked) {
		rep.Blocked++
	}
	if e.Canonical == "" {
		e.Canonical = e.Alias
	}
	prof, ok := b.typeProfiles[e.Type]
	if !ok {
		prof = b.defaultProfile
	}
	terms := b.analyzer.Analyze(e.Alias, prof, lexicon.ModeFull)
	var alts [][]string
	var forms []string
	var cases []textnorm.Case
	for i := range terms {
		t := &terms[i]
		if !IsContent(t) {
			continue
		}
		alts = append(alts, Alternatives(t))
		forms = append(forms, t.Form)
		cases = append(cases, t.Token.Case)
	}
	if len(forms) == 0 {
		rep.Skipped++
		rep.Errors = append(rep.Errors, fmt.Sprintf("alias %q (%s): no words", e.Alias, e.Type))
		return nil
	}
	a := &Alias{Entry: e, Source: source, Cases: cases, SurfaceKey: strings.Join(forms, " ")}
	sb.insert(internAll(forms), a)
	rep.SurfaceKeys++
	if !e.Flags.Has(SurfaceOnly) {
		combos, capped := Combine(alts, MaxLemmaKeys)
		if capped {
			rep.Capped++
			rep.Errors = append(rep.Errors, fmt.Sprintf("alias %q (%s): lemma expansion capped at %d", e.Alias, e.Type, MaxLemmaKeys))
		}
		for _, c := range combos {
			lb.insert(internAll(c), a)
			a.LemmaKeys = append(a.LemmaKeys, strings.Join(c, " "))
		}
		rep.LemmaKeys += len(combos)
	}
	rep.Aliases++
	return a
}

func internAll(ss []string) []uint32 {
	ids := make([]uint32, len(ss))
	for i, s := range ss {
		ids[i] = global.intern(s)
	}
	return ids
}
