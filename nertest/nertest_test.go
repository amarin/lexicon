package nertest

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/amarin/lexicon/gazetteer"
	"github.com/amarin/lexicon/internal/fakedict"
	"github.com/amarin/lexicon/ner"
	"github.com/amarin/lexicon/rules"
)

// stubExtractor returns predefined spans (located by surface) per text.
type stubExtractor map[string][][2]string // text → [surface, type]

func (s stubExtractor) Extract(_ context.Context, d ner.Doc, _ ...ner.Option) (ner.Result, error) {
	var res ner.Result
	for _, st := range s[d.Text] {
		i := strings.Index(d.Text, st[0])
		res.Spans = append(res.Spans, ner.Span{Start: i, End: i + len(st[0]), Surface: st[0], Type: st[1]})
	}
	return res, nil
}

func TestRunMetrics(t *testing.T) {
	cases, err := LoadCases(strings.NewReader(`
{"id": "a", "text": "Иван Петров", "spans": [{"text": "Иван", "type": "given_name"}, {"text": "Петров", "type": "surname"}]}
`))
	if err != nil {
		t.Fatal(err)
	}
	ex := stubExtractor{"Иван Петров": {{"Иван", "given_name"}, {"Иван Петров", "surname"}}}
	rep, err := Run(context.Background(), ex, cases)
	if err != nil {
		t.Fatal(err)
	}
	g, s := rep.Types["given_name"], rep.Types["surname"]
	if g.Strict != (Counts{TP: 1}) || s.Strict != (Counts{FP: 1, FN: 1}) || s.Partial != (Counts{TP: 1}) {
		t.Fatalf("counts: given %+v surname %+v", g, s)
	}
	if s.Strict.Precision() != 0 || s.Partial.Recall() != 1 || g.Strict.F1() != 1 {
		t.Fatal("ratios are wrong")
	}
	if len(rep.Failures) != 2 || rep.Failures[0].Kind != "spurious" || rep.Failures[1].Kind != "missed" || rep.Failures[1].Text != "Петров" {
		t.Fatalf("failures = %+v", rep.Failures)
	}
	if v := rep.Check(0.9, 0.9); len(v) != 2 || !strings.HasPrefix(v[0], "surname: strict precision") {
		t.Fatalf("Check = %q", v)
	}
	var buf bytes.Buffer
	if err := rep.Write(&buf); err != nil || !strings.Contains(buf.String(), "surname") || !strings.Contains(buf.String(), "missed") {
		t.Fatalf("Write: %v\n%s", err, buf.String())
	}
}

func TestLoadCasesErrors(t *testing.T) {
	for name, js := range map[string]string{
		"not found":     `{"text": "Иван", "spans": [{"text": "Петров", "type": "surname"}]}`,
		"no type":       `{"text": "Иван", "spans": [{"text": "Иван"}]}`,
		"unknown field": `{"text": "Иван", "spans": [], "colour": 1}`,
		"occurrence":    `{"text": "Иван Иван", "spans": [{"text": "Иван", "type": "given_name", "occurrence": 3}]}`,
	} {
		if _, err := LoadCases(strings.NewReader(js)); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
	cases, err := LoadCases(strings.NewReader(`{"text": "Иван Иван", "spans": [{"text": "Иван", "type": "given_name", "occurrence": 2}]}`))
	if err != nil || cases[0].ID != "line1" {
		t.Fatalf("second occurrence: %+v %v", cases, err)
	}
	b, _ := cases[0].goldBounds()
	if b[0].start != len("Иван ") {
		t.Fatalf("occurrence 2 at %d", b[0].start)
	}
}

// genealogyPipeline builds a pipeline over the testdata gazetteer and rules,
// plus extra rule files.
func genealogyPipeline(t *testing.T, extra ...rules.File) *ner.Pipeline {
	t.Helper()
	ctx := context.Background()
	an := fakedict.NewAnalyzer()
	gz, err := gazetteer.New(ctx, gazetteer.Config{
		Analyzer: an, TypeProfiles: fakedict.TypeProfiles(), DefaultProfile: fakedict.Profiles()["text"],
		Sources: []gazetteer.Source{gazetteer.NewTSVSource("genealogy", "testdata/genealogy.tsv")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if rep := gz.Snapshot().Reports()[0]; rep.Err != nil || len(rep.Errors) != 0 || rep.Aliases != 11 {
		t.Fatalf("gazetteer report = %+v", rep)
	}
	rf, err := rules.LoadFile("testdata/genealogy.rules.yaml")
	if err != nil {
		t.Fatal(err)
	}
	book, err := rules.Compile(append([]rules.File{rf}, extra...)...)
	if err != nil {
		t.Fatal(err)
	}
	p, err := ner.New(ner.Config{Analyzer: an, Gazetteer: gz, Rules: book, Profiles: fakedict.Profiles(), DefaultProfile: "text"})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestGenealogyGolden(t *testing.T) {
	ctx := context.Background()
	p := genealogyPipeline(t)
	cases, err := LoadCasesFile("testdata/genealogy.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	rep, err := Run(ctx, p, cases)
	if err != nil {
		t.Fatal(err)
	}
	if v := rep.Check(1, 1); len(v) != 0 {
		var buf bytes.Buffer
		_ = rep.Write(&buf)
		t.Fatalf("golden set regressed: %q\n%s", v, buf.String())
	}
	if rep.Cases != 6 {
		t.Fatalf("cases = %d", rep.Cases)
	}
}

// contextRules gates the «крестьянин X» surname hint by a tag that only
// WithTags derives from Case.Context.
const contextRules = `
sets:
  - name: peasant
    when: ["estate:peasant"]
    hints:
      - {lemma: крестьянин, type: surname, dir: right, window: 2, weight: 1}
`

// TestContextTags: Case.Context is loaded but ignored by Run itself;
// WithTags turns it into document tags appended to Case.Tags, which activate
// a rule set (owner decision 2026-09-25).
func TestContextTags(t *testing.T) {
	ctx := context.Background()
	extra, err := rules.LoadNamed("context.yaml", strings.NewReader(contextRules))
	if err != nil {
		t.Fatal(err)
	}
	p := genealogyPipeline(t, extra)
	cases, err := LoadCases(strings.NewReader(`{"id": "ctx", "text": "крестьянин Мороз", "context": {"estate": "peasant"}, "spans": [{"text": "крестьянин", "type": "estate"}, {"text": "Мороз", "type": "surname"}]}`))
	if err != nil || cases[0].Context["estate"] != "peasant" || len(cases[0].Tags) != 0 {
		t.Fatalf("LoadCases = %+v, %v", cases, err)
	}
	rep, err := Run(ctx, p, cases)
	if err != nil {
		t.Fatal(err)
	}
	if s := rep.Types["surname"].Strict; s != (Counts{FN: 1}) {
		t.Fatalf("without WithTags the peasant set must stay inactive: surname %+v", s)
	}
	byContext := WithTags(func(c Case) []string {
		var tags []string
		for k, v := range c.Context {
			tags = append(tags, k+":"+v)
		}
		return tags
	})
	rep, err = Run(ctx, p, cases, byContext)
	if err != nil {
		t.Fatal(err)
	}
	if v := rep.Check(1, 1); len(v) != 0 {
		var buf bytes.Buffer
		_ = rep.Write(&buf)
		t.Fatalf("estate:peasant not applied: %q\n%s", v, buf.String())
	}
	if len(cases[0].Tags) != 0 {
		t.Fatalf("Run modified Case.Tags: %q", cases[0].Tags)
	}
}
