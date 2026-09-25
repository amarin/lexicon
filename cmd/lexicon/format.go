package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/amarin/lexicon"
)

// formatTerms writes one line per term: "raw<TAB>form<TAB>lemma[flags] …";
// flags are omitted when empty.
func formatTerms(w io.Writer, terms []lexicon.Term) {
	for _, t := range terms {
		ls := make([]string, 0, len(t.Lemmas))

		for _, l := range t.Lemmas {
			if l.Flags == 0 {
				ls = append(ls, l.Text)

				continue
			}

			ls = append(ls, fmt.Sprintf("%s[%s]", l.Text, l.Flags))
		}

		fmt.Fprintf(w, "%s\t%s\t%s\n", t.Token.Raw, t.Form, strings.Join(ls, " "))
	}
}
