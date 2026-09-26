package main

import "github.com/amarin/lexicon/ner"

// spanJSON is one JSONL output line of extract.
type spanJSON struct {
	Doc          int               `json:"doc"`
	Start        int               `json:"start"`
	End          int               `json:"end"`
	RuneStart    int               `json:"rune_start"`
	RuneEnd      int               `json:"rune_end"`
	Type         string            `json:"type"`
	Surface      string            `json:"surface"`
	Normal       []string          `json:"normal,omitempty"`
	Refs         []string          `json:"refs,omitempty"`
	Attrs        map[string]string `json:"attrs,omitempty"`
	Flags        []string          `json:"flags,omitempty"`
	Score        float32           `json:"score"`
	Evidence     []string          `json:"evidence,omitempty"`
	Alternatives []alternativeJSON `json:"alternatives,omitempty"`
}

func newSpanJSON(doc int, s ner.Span) spanJSON {
	out := spanJSON{
		Doc: doc, Start: s.Start, End: s.End, RuneStart: s.RuneStart, RuneEnd: s.RuneEnd,
		Type: s.Type, Surface: s.Surface, Normal: s.Normal, Refs: s.Refs, Attrs: s.Attrs,
		Flags: s.Flags.Names(), Score: s.Score, Evidence: s.Evidence,
	}
	for _, a := range s.Alternatives {
		out.Alternatives = append(out.Alternatives, alternativeJSON{
			Type: a.Type, Normal: a.Normal, Refs: a.Refs, Attrs: a.Attrs, Score: a.Score, Evidence: a.Evidence,
		})
	}
	return out
}
