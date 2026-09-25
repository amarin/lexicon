package textnorm

import "unicode"

// caseOf classifies the letter case of raw word text. A single upper-case
// letter is Title («И»).
func caseOf(raw string) Case {
	var upper, lower int

	firstUpper, seen, partsTitle, partStart := false, false, true, true

	for _, r := range raw {
		switch {
		case isHyphen(r):
			partStart = true

			continue
		case unicode.IsUpper(r):
			if !seen {
				firstUpper = true
			}

			upper++

			if !partStart {
				partsTitle = false
			}
		case unicode.IsLower(r):
			lower++

			if partStart {
				partsTitle = false
			}
		default:
			continue
		}

		seen, partStart = true, false
	}

	switch {
	case upper+lower == 0:
		return CaseNone
	case upper == 0:
		return CaseLower
	case lower == 0 && upper > 1:
		return CaseUpper
	case firstUpper && upper == 1, partsTitle:
		return CaseTitle
	}

	return CaseMixed
}
