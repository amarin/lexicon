package rules

import (
	"fmt"
	"strings"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/textnorm"
)

// shapeCheck is a compiled Shape.
type shapeCheck struct {
	hasCase, hasScript bool
	c                  textnorm.Case
	s                  textnorm.Script
}

func compileShape(sh Shape) (shapeCheck, error) {
	var sc shapeCheck
	switch strings.ToLower(sh.Case) {
	case "":
	case "lower":
		sc.hasCase, sc.c = true, textnorm.CaseLower
	case "title":
		sc.hasCase, sc.c = true, textnorm.CaseTitle
	case "upper":
		sc.hasCase, sc.c = true, textnorm.CaseUpper
	default:
		return shapeCheck{}, fmt.Errorf("unknown shape case %q (want lower, title or upper)", sh.Case)
	}
	switch strings.ToLower(sh.Script) {
	case "":
	case "cyrillic":
		sc.hasScript, sc.s = true, textnorm.ScriptCyrillic
	case "latin":
		sc.hasScript, sc.s = true, textnorm.ScriptLatin
	default:
		return shapeCheck{}, fmt.Errorf("unknown shape script %q (want cyrillic or latin)", sh.Script)
	}
	return sc, nil
}

func (sc shapeCheck) ok(t *lexicon.Term) bool {
	if sc.hasCase && t.Token.Case != sc.c {
		return false
	}
	if sc.hasScript && t.Token.Script != sc.s {
		return false
	}
	return true
}
