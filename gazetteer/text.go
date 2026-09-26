package gazetteer

import (
	"unicode/utf8"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/textnorm"
)

// Text is analyzed text prepared for matching: content positions (words
// and numbers) with interned lemma and form IDs, match breaks and sentence
// segments. Build it with Prepare; it is immutable afterwards.
type Text struct {
	terms   []lexicon.Term
	pos     []int    // term index of each content position
	lemOff  []int    // lemma IDs of position p are lemIDs[lemOff[p]:lemOff[p+1]]
	lemIDs  []uint32 // 0-free: unknown lemmas are dropped
	formIDs []uint32 // 0 when the form was never interned
	brk     []bool   // brk[p]: position p+1 is not reachable from p
	sents   [][2]int // term index ranges
}

// Prepare builds a Text over terms produced by Analyzer.Analyze(ModeFull).
func Prepare(terms []lexicon.Term) *Text {
	tx := &Text{terms: terms}
	for i := range terms {
		if IsContent(&terms[i]) {
			tx.pos = append(tx.pos, i)
		}
	}
	tx.lemOff = make([]int, 0, len(tx.pos)+1)
	tx.formIDs = make([]uint32, 0, len(tx.pos))
	for _, i := range tx.pos {
		tx.lemOff = append(tx.lemOff, len(tx.lemIDs))
		for _, s := range Alternatives(&terms[i]) {
			if id := global.lookup(s); id != 0 {
				tx.lemIDs = append(tx.lemIDs, id)
			}
		}
		tx.formIDs = append(tx.formIDs, global.lookup(terms[i].Form))
	}
	tx.lemOff = append(tx.lemOff, len(tx.lemIDs))
	tx.brk = make([]bool, len(tx.pos))
	for p, i := range tx.pos {
		next := len(terms)
		if p+1 < len(tx.pos) {
			next = tx.pos[p+1]
		}
		tx.brk[p] = !joins(terms, i, next)
	}
	start := 0
	for i := range terms {
		if endsSentence(terms, i) {
			tx.sents = append(tx.sents, [2]int{start, i + 1})
			start = i + 1
		}
	}
	if start < len(terms) {
		tx.sents = append(tx.sents, [2]int{start, len(terms)})
	}
	return tx
}

// Terms returns the analyzed terms (shared, read-only).
func (t *Text) Terms() []lexicon.Term { return t.terms }

// Len returns the number of content positions.
func (t *Text) Len() int { return len(t.pos) }

// TermIndex returns the term index of content position p.
func (t *Text) TermIndex(p int) int { return t.pos[p] }

// Break reports whether a match cannot continue from position p to p+1.
func (t *Text) Break(p int) bool { return t.brk[p] }

// Joined reports whether positions a..b (a <= b) are connected.
func (t *Text) Joined(a, b int) bool {
	for p := a; p < b; p++ {
		if t.brk[p] {
			return false
		}
	}
	return true
}

// Sentences returns contiguous term index ranges [start, end) covering all
// terms; abbreviation dots do not end a sentence.
func (t *Text) Sentences() [][2]int { return t.sents }

// joins reports whether content term i connects to the next content term
// at index next (len(terms) when there is none).
func joins(terms []lexicon.Term, i, next int) bool {
	if next >= len(terms) {
		return false
	}
	if abbreviationLike(&terms[i]) {
		for k := i + 1; k < next; k++ {
			if !isDot(&terms[k]) {
				return false
			}
		}
		return true
	}
	if terms[i].Token.SentenceEnd {
		return false
	}
	return next == i+1
}

// abbreviationLike: a dotted abbreviation («дер.») or an initial («И.»).
func abbreviationLike(t *lexicon.Term) bool {
	if !t.Token.Dotted {
		return false
	}
	if utf8.RuneCountInString(t.Token.Raw) == 1 {
		return true
	}
	for _, l := range t.Lemmas {
		if l.Flags&lexicon.FlagAbbrev != 0 {
			return true
		}
	}
	return false
}

func isDot(t *lexicon.Term) bool {
	return t.Token.Kind == textnorm.TokenPunct && t.Token.Raw == "."
}

func endsSentence(terms []lexicon.Term, i int) bool {
	t := &terms[i]
	if !t.Token.SentenceEnd || abbreviationLike(t) {
		return false
	}
	if isDot(t) && i > 0 && abbreviationLike(&terms[i-1]) {
		return false
	}
	return true
}
