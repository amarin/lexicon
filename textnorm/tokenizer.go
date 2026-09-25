package textnorm

import "unicode"

// tokenizer is the state of one Tokenize call.
type tokenizer struct {
	rules  Rules
	in     string
	p      []runePos
	out    []Token
	breaks int // line breaks since the last token (spec Q5)
}

// end is the byte offset of code point j (len(in) past the end).
func (t *tokenizer) end(j int) int {
	if j < len(t.p) {
		return t.p[j].off
	}

	return len(t.in)
}

func (t *tokenizer) scan() {
	for i := 0; i < len(t.p); {
		c := t.p[i].r

		switch {
		case isLineBreak(c):
			// A single line break is a space (hard-wrapped text keeps
			// «Нижний⏎Новгород» in one sentence); a blank line or a
			// paragraph separator ends the sentence (spec Q5).
			if c == '\r' && i+1 < len(t.p) && t.p[i+1].r == '\n' {
				i++
			}

			t.breaks++
			if t.breaks >= 2 || c == '\u2029' {
				t.markSentenceEnd()
			}

			i++
		case isSpace(c), isExtend(c), unicode.IsControl(c):
			i++
		case isDigit(c):
			i = t.number(i)
		case unicode.IsLetter(c):
			i = t.word(i)
		case unicode.IsPunct(c):
			j := t.clusterEnd(i + 1)
			t.emit(i, j, TokenPunct, "")

			if isSentencePunct(c) {
				t.markSentenceEnd()
			}

			i = j
		default:
			j := i + 1
			if isRegionalIndicator(c) && j < len(t.p) && isRegionalIndicator(t.p[j].r) {
				j++
			}

			j = t.clusterEnd(j)
			t.emit(i, j, TokenSymbol, "")
			i = j
		}
	}
}

// emit appends the token of code points [i, j) and returns it for further
// fields; the pointer is valid until the next emit.
func (t *tokenizer) emit(i, j int, kind TokenKind, form string) *Token {
	t.breaks = 0
	start, end := t.p[i].off, t.end(j)
	t.out = append(t.out, Token{
		Raw: t.in[start:end], Form: form, Start: start, End: end, RuneStart: i, RuneEnd: j, Kind: kind,
	})

	return &t.out[len(t.out)-1]
}

func (t *tokenizer) markSentenceEnd() {
	if n := len(t.out); n > 0 {
		t.out[n-1].SentenceEnd = true
	}
}

// clusterEnd extends a cluster ending before j over extending code points and
// over a pictographic symbol joined by ZWJ.
func (t *tokenizer) clusterEnd(j int) int {
	for j < len(t.p) {
		c := t.p[j].r
		if isExtend(c) || (t.p[j-1].r == zwj && unicode.Is(unicode.So, c)) {
			j++

			continue
		}

		break
	}

	return j
}

func (t *tokenizer) number(i int) int {
	j := i

	var form []byte

	for j < len(t.p) && isDigit(t.p[j].r) {
		form = append(form, byte(t.p[j].r))
		j = t.clusterEnd(j + 1)
	}

	t.emit(i, j, TokenNumber, string(form))

	return j
}

func (t *tokenizer) word(i int) int {
	j := t.wordEnd(i)
	raw := t.in[t.p[i].off:t.end(j)]

	tok := t.emit(i, j, TokenWord, NormalizeWord(t.rules, raw))
	tok.Script = t.script(i, j)
	tok.Case = caseOf(raw)
	tok.Dotted = j < len(t.p) && t.p[j].r == '.'

	return j
}

// wordEnd returns the end of the word starting at letter i: letters of the
// same category (Cyrillic or not) with their clusters, joined over single
// inner hyphens.
func (t *tokenizer) wordEnd(i int) int {
	cyr := t.p[i].cyr
	j := t.clusterEnd(i + 1)

	for j < len(t.p) {
		c := t.p[j].r
		letter := unicode.IsLetter(c) && t.p[j].cyr == cyr
		hyphen := isHyphen(c) && j+1 < len(t.p) && unicode.IsLetter(t.p[j+1].r) && t.p[j+1].cyr == cyr

		if !letter && !hyphen {
			break
		}

		j = t.clusterEnd(j + 1)
	}

	return j
}

// classify sets runePos.cyr before scanning. A letter run (letters and
// extending code points between any other code points, so hyphens end a run)
// that isCyrillicPart accepts is Cyrillic as a whole, its Latin homoglyphs
// included («Kот»); in any other run each letter keeps its own category
// («Kотw» → K | от | w, decision D2).
func (t *tokenizer) classify() {
	var run []rune

	for i := 0; i < len(t.p); {
		j := i
		run = run[:0]

		for j < len(t.p) && (unicode.IsLetter(t.p[j].r) || isExtend(t.p[j].r)) {
			run = append(run, t.p[j].r)
			j++
		}

		if j == i {
			i++

			continue
		}

		whole := isCyrillicPart(t.rules.Homoglyphs, run)
		for k := i; k < j; k++ {
			t.p[k].cyr = whole || unicode.Is(unicode.Cyrillic, t.p[k].r)
		}

		i = j
	}
}

// script of the word [i, j); a word is homogeneous in the Cyrillic category.
func (t *tokenizer) script(i, j int) Script {
	if t.p[i].cyr {
		return ScriptCyrillic
	}

	for k := i; k < j; k++ {
		if c := t.p[k].r; unicode.IsLetter(c) && !unicode.Is(unicode.Latin, c) {
			return ScriptMixed
		}
	}

	return ScriptLatin
}
