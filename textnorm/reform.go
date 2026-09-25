package textnorm

import (
	"strings"
	"unicode/utf8"
)

// ReformVariants returns modern spellings of a form with a pre-reform adjective
// ending, nil when the ending is not pre-reform. Analyzers try the variants in
// order only when the form itself has no exact dictionary reading and take the
// first one the dictionary knows; the indexed form stays as in the text.
//
//   - -яго → -его («синяго» → «синего»);
//   - -аго → -ого; after a sibilant (ж, ш, ч, щ) -ого, then -его: the ending
//     depends on stress («большаго» → «большого», «хорошаго» → «хорошего»);
//   - -ыя → -ые, -ой («новыя»);
//   - -ия → -ие, then -ой after к/г/х («покровския» → «покровской») or -ей
//     otherwise («синия» → «синей»).
func ReformVariants(form string) []string {
	switch {
	case strings.HasSuffix(form, "яго"):
		return []string{strings.TrimSuffix(form, "яго") + "его"}
	case strings.HasSuffix(form, "аго"):
		stem := strings.TrimSuffix(form, "аго")
		if endsWithAny(stem, "жшчщ") {
			return []string{stem + "ого", stem + "его"}
		}

		return []string{stem + "ого"}
	case strings.HasSuffix(form, "ыя"):
		stem := strings.TrimSuffix(form, "ыя")

		return []string{stem + "ые", stem + "ой"}
	case strings.HasSuffix(form, "ия"):
		stem := strings.TrimSuffix(form, "ия")
		if endsWithAny(stem, "кгх") {
			return []string{stem + "ие", stem + "ой"}
		}

		return []string{stem + "ие", stem + "ей"}
	}

	return nil
}

// endsWithAny reports whether the last letter of s is one of letters.
func endsWithAny(s, letters string) bool {
	r, size := utf8.DecodeLastRuneInString(s)

	return size > 0 && strings.ContainsRune(letters, r)
}
