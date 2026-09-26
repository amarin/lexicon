package rules

import (
	"errors"
	"fmt"

	"github.com/amarin/lexicon"
)

// MaxWindow bounds hint and trigger windows.
const MaxWindow = 8

// HintRule is a validated Hint with defaults applied.
type HintRule struct {
	Hint
	Set string // name of the rule set
	key keyword
}

func compileHint(set string, h Hint) (*HintRule, error) {
	if h.Type == "" {
		return nil, errors.New("empty type")
	}
	key, err := newKeyword(h.Lemma, h.Dotted)
	if err != nil {
		return nil, err
	}
	if h.Window == 0 {
		h.Window = 1
	}
	if h.Window < 1 || h.Window > MaxWindow {
		return nil, fmt.Errorf("window %d out of 1..%d", h.Window, MaxWindow)
	}
	if h.Weight == 0 {
		h.Weight = 1
	}
	return &HintRule{Hint: h, Set: set, key: key}, nil
}

// Keyword reports whether t is this hint's keyword.
func (h *HintRule) Keyword(t *lexicon.Term) bool { return h.key.match(t) }
