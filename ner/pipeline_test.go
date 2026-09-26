package ner

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/internal/fakedict"
)

func TestNewValidates(t *testing.T) {
	p, gz := newPipeline(t, testEntries(), "")
	an := fakedict.NewAnalyzer()
	for name, cfg := range map[string]Config{
		"no analyzer":     {Gazetteer: gz, Profiles: fakedict.Profiles(), DefaultProfile: "text"},
		"no gazetteer":    {Analyzer: an, Profiles: fakedict.Profiles(), DefaultProfile: "text"},
		"missing default": {Analyzer: an, Gazetteer: gz, Profiles: fakedict.Profiles(), DefaultProfile: "nope"},
	} {
		if _, err := New(cfg); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
	if w := p.weights; w.Surface != 3 || w.Lemma != 2 || w.Trigger != 1 || w.LengthBonus != 0.5 || w.AmbiguityPenalty != 0.25 || p.minRunes != 3 {
		t.Fatalf("defaults not applied: %+v %d", p.weights, p.minRunes)
	}
}

func TestExtractVersionAndProfile(t *testing.T) {
	p, gz := newPipeline(t, testEntries(), placesRules)
	r1 := extract(t, p, Doc{Text: ""})
	r2 := extract(t, p, Doc{Text: "Иван", Profile: "name"})
	if r1.Version == "" || r1.Version != r2.Version || len(r1.Spans) != 0 {
		t.Fatalf("versions: %q %q", r1.Version, r2.Version)
	}
	if _, err := gz.RefreshSource(context.Background(), "test"); err != nil {
		t.Fatal(err)
	}
	if r3 := extract(t, p, Doc{}); r3.Version != r1.Version {
		t.Fatal("rebuilding identical source content must keep the version")
	}
	p2, _ := newPipeline(t, testEntries(), placesRules, func(c *Config) { c.Nesting = map[string][]string{"division": {"division"}} })
	if extract(t, p2, Doc{}).Version == r1.Version {
		t.Fatal("configuration must be part of the version")
	}
	p3, _ := newPipeline(t, testEntries(), "")
	if extract(t, p3, Doc{}).Version == r1.Version {
		t.Fatal("rules must be part of the version")
	}
	if _, err := p.Extract(context.Background(), Doc{Text: "Иван", Profile: "nope"}); !errors.Is(err, ErrUnknownProfile) {
		t.Fatalf("unknown profile: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := p.Extract(ctx, Doc{Text: "Иван"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled context: %v", err)
	}
}

// Result.Version covers the document profiles: another dictionary set for
// a profile changes the output, so it must change the version.
func TestExtractVersionCoversProfiles(t *testing.T) {
	p1, _ := newPipeline(t, testEntries(), placesRules)
	p2, _ := newPipeline(t, testEntries(), placesRules, func(c *Config) {
		c.Profiles = fakedict.Profiles()
		name := c.Profiles["name"]
		name.Kinds = []lexicon.Kind{lexicon.KindBase}
		c.Profiles["name"] = name
	})
	v1, v2 := extract(t, p1, Doc{}).Version, extract(t, p2, Doc{}).Version
	if v1 == v2 {
		t.Fatal("changing a profile's kinds must change Result.Version")
	}
	p3, _ := newPipeline(t, testEntries(), placesRules)
	if extract(t, p3, Doc{}).Version != v1 {
		t.Fatal("identical configuration must give the same version")
	}
}

// New copies the host-owned maps: editing them afterwards changes neither
// the output nor the version.
func TestNewCopiesHostMaps(t *testing.T) {
	profiles := fakedict.Profiles()
	types := map[string]float32{"division": 1}
	nesting := map[string][]string{"division": {"division"}}
	p, _ := newPipeline(t, testEntries(), placesRules, func(c *Config) {
		c.Profiles, c.Weights.Types, c.Nesting = profiles, types, nesting
	})
	const text = "Лягушкино Боровского уезда, Иван Петров"
	before := extract(t, p, Doc{Text: text}, Explain())
	text0 := profiles["text"]
	text0.Kinds[0] = "nope"
	profiles["text"] = lexicon.Profile{Name: "text", Kinds: []lexicon.Kind{"nope"}}
	types["division"] = -100
	nesting["division"][0] = "street"
	after := extract(t, p, Doc{Text: text}, Explain())
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("host map edits leaked into the pipeline:\n before %+v\n after  %+v", before, after)
	}
}
