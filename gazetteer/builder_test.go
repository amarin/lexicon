package gazetteer

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestBuildKeysAndReport(t *testing.T) {
	b := &builder{analyzer: stubAnalyzer{version: "s1"}}
	src := NewSliceSource("places", "v1", []Entry{
		{Alias: "Дер/деревня . Лягушкино/лягушкино", Type: "division", Ref: "division:1", Canonical: "Лягушкино"},
		{Alias: "С/село/сын Иваново", Type: "division", Ref: "division:2"},
		{Alias: "СПб", Type: "division", Ref: "division:3", Flags: SurfaceOnly},
		{Alias: ".", Type: "division"},
		{Alias: "a/1/2 b/1/2 c/1/2 d/1/2", Type: "t"},
		{Alias: "Мороз", Type: "surname", Flags: Blocked},
	})
	cs := b.build(context.Background(), src, "v1", "s1", &compiledSource{name: "places"})
	r := cs.report
	if r.Err != nil || !cs.built || cs.version != "v1" || cs.analyzerVersion != "s1" {
		t.Fatalf("build state: %+v", cs)
	}
	if r.Entries != 6 || r.Aliases != 5 || r.Skipped != 1 || r.Blocked != 1 || r.Capped != 1 ||
		r.SurfaceKeys != 5 || r.LemmaKeys != 12 {
		t.Fatalf("report = %+v", r)
	}
	if len(r.Errors) != 2 || !strings.Contains(r.Errors[0], "no words") || !strings.Contains(r.Errors[1], "capped at 8") {
		t.Fatalf("report errors = %q", r.Errors)
	}
	der := aliasByText(cs, "Дер/деревня . Лягушкино/лягушкино")
	if !reflect.DeepEqual(der.LemmaKeys, []string{"деревня лягушкино"}) || der.SurfaceKey != "дер лягушкино" {
		t.Fatalf("der keys: %q %q", der.LemmaKeys, der.SurfaceKey)
	}
	s := aliasByText(cs, "С/село/сын Иваново")
	keys := slices.Clone(s.LemmaKeys)
	slices.Sort(keys)
	if !reflect.DeepEqual(keys, []string{"село иваново", "сын иваново"}) || s.Entry.Canonical != "С/село/сын Иваново" {
		t.Fatalf("ambiguous alias: %q canonical %q", keys, s.Entry.Canonical)
	}
	if spb := aliasByText(cs, "СПб"); len(spb.LemmaKeys) != 0 || spb.SurfaceKey != "спб" {
		t.Fatalf("surface-only alias: %+v", spb)
	}
}

func TestBuildKeepsPreviousOnFatalError(t *testing.T) {
	b := &builder{analyzer: stubAnalyzer{version: "s1"}}
	prev := b.build(context.Background(), NewSliceSource("broken", "v1", []Entry{{Alias: "Петров", Type: "surname"}}), "v1", "s1", &compiledSource{name: "broken"})
	boom := errors.New("database is locked")
	cs := b.build(context.Background(), failingSource{err: boom}, "v2", "s1", prev)
	if !errors.Is(cs.report.Err, boom) || cs.version != "v1" || cs.lemma != prev.lemma || !cs.built {
		t.Fatalf("previous data not kept: %+v", cs)
	}
	pe := &ParseErrors{Source: "broken", Lines: []LineError{{Line: 3, Msg: "empty alias"}}}
	cs = b.build(context.Background(), failingSource{err: pe}, "v2", "s1", prev)
	if cs.report.Err != nil || cs.version != "v2" || len(cs.report.Errors) != 1 || cs.report.Errors[0] != "line 3: empty alias" {
		t.Fatalf("parse errors must be non-fatal: %+v", cs.report)
	}
}
