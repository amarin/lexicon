package nertest

import (
	"errors"
	"fmt"
	"strings"
)

// Case is one golden document.
type Case struct {
	ID      string   `json:"id,omitempty"`
	Text    string   `json:"text"`
	Profile string   `json:"profile,omitempty"`
	Tags    []string `json:"tags,omitempty"`
	Spans   []Gold   `json:"spans"`

	// Context is host data about the document (e.g. {"estate": "peasant"}).
	// lexicon ignores it; a WithTags option may derive tags from it.
	Context map[string]string `json:"context,omitempty"`
}

// goldBounds locates every gold span in the text.
func (c Case) goldBounds() ([]bounds, error) {
	out := make([]bounds, 0, len(c.Spans))
	for _, g := range c.Spans {
		if g.Text == "" || g.Type == "" {
			return nil, errors.New("gold span needs text and type")
		}
		n := max(g.Occurrence, 1)
		from, at := 0, -1
		for i := 0; i < n; i++ {
			k := strings.Index(c.Text[from:], g.Text)
			if k < 0 {
				return nil, fmt.Errorf("gold span %q (occurrence %d) not found in text", g.Text, n)
			}
			at = from + k
			from = at + len(g.Text)
		}
		out = append(out, bounds{start: at, end: at + len(g.Text), typ: g.Type})
	}
	return out, nil
}
