package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/amarin/lexicon/ner"
)

// factJSON is one JSONL fact line of extract; Args are indexes among the
// document's span lines, in output order.
type factJSON struct {
	Doc  int            `json:"doc"`
	Fact string         `json:"fact"`
	Args map[string]int `json:"args"`
	Rule string         `json:"rule,omitempty"`
}

func newFactJSON(doc int, f ner.Fact) factJSON {
	return factJSON{Doc: doc, Fact: f.Kind, Args: f.Args, Rule: f.Rule}
}

// describeFact renders "role=«surface» …" with roles sorted.
func describeFact(f ner.Fact, spans []ner.Span) string {
	roles := make([]string, 0, len(f.Args))
	for r := range f.Args {
		roles = append(roles, r)
	}
	slices.Sort(roles)
	parts := make([]string, len(roles))
	for i, r := range roles {
		parts[i] = fmt.Sprintf("%s=«%s»", r, spans[f.Args[r]].Surface)
	}
	return strings.Join(parts, " ")
}
