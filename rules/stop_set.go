package rules

import (
	"fmt"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/textnorm"
)

// stopSet is a compiled Trigger.StopAt. "punct" is accepted and always in
// effect: punctuation breaks the walk (gazetteer.Text breaks).
type stopSet struct{ stop, number, latin bool }

func compileStops(names []string) (stopSet, error) {
	var s stopSet
	for _, n := range names {
		switch n {
		case "punct":
		case "stop":
			s.stop = true
		case "number":
			s.number = true
		case "latin":
			s.latin = true
		default:
			return stopSet{}, fmt.Errorf("unknown stop_at class %q (want punct, stop, number or latin)", n)
		}
	}
	return s, nil
}

func (s stopSet) stops(t *lexicon.Term) bool {
	if s.number && t.Token.Kind == textnorm.TokenNumber {
		return true
	}
	if s.latin && t.Token.Script == textnorm.ScriptLatin {
		return true
	}
	if s.stop {
		for _, l := range t.Lemmas {
			if l.Flags&lexicon.FlagStop != 0 {
				return true
			}
		}
	}
	return false
}
