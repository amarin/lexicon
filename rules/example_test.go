package rules_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/amarin/lexicon/rules"
)

// A rule file holds rule sets of hints and triggers; a set with when: is
// active only for documents carrying all of its tags.
func ExampleCompile() {
	const yaml = `meta: {source: example, license: CC0-1.0}
sets:
  - name: places
    hints:
      - {lemma: уезд, type: division, dir: left, window: 1, weight: 2, absorb: true}
    triggers:
      - {lemma: деревня|село, type: division, window: 1..2, shape: {case: title, script: cyrillic}, absorb: true}
  - name: pre1917
    when: ["period:pre1917"]
    hints:
      - {lemma: крестьянин, type: surname, window: 2}
`
	f, err := rules.LoadNamed("places.yaml", strings.NewReader(yaml))
	if err != nil {
		log.Fatal(err)
	}
	book, err := rules.Compile(f)
	if err != nil {
		log.Fatal(err)
	}
	for _, tags := range [][]string{nil, {"period:pre1917"}} {
		a := book.Active(tags)
		fmt.Println(tags, a.Sets, len(a.Hints), len(a.Triggers))
	}
	// Output:
	// [] [places] 1 1
	// [period:pre1917] [places pre1917] 2 1
}

// Errors name the file and line.
func ExampleLoadNamed() {
	_, err := rules.LoadNamed("bad.yaml", strings.NewReader("sets:\n  - name: s\n    hints:\n      - {lemma: уезд, type: division, dir: up}\n"))
	fmt.Println(err)
	// Output: bad.yaml:4: unknown direction "up" (want left, right or both)
}
