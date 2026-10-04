package rules

import "github.com/amarin/lexicon/textnorm"

// matcher executes a program with backtracking and a visited set over
// (instruction, position): each pair is explored at most once per start
// position, which keeps leftmost-first semantics and bounds the work by
// len(insts) × (n+1). visited holds the generation that last explored a
// pair, so starting a new scan costs nothing.
type matcher struct {
	insts   []inst
	l       Lattice
	n       int
	visited []uint32
	gen     uint32
	caps    []int // 2 per role: start, end
	spans   []int // 1 per role: LatticeSpan.ID or -1
	used    []int // IDs of the candidates consumed on the current path
}

func (m *matcher) reset() {
	m.gen++
	for i := range m.caps {
		m.caps[i] = -1
	}
	for i := range m.spans {
		m.spans[i] = -1
	}
	m.used = m.used[:0]
}

// run returns the end position of the first accepting path from (pc, pos).
func (m *matcher) run(pc, pos int) (int, bool) {
	for {
		k := pc*(m.n+1) + pos
		if m.visited[k] == m.gen {
			return 0, false
		}
		m.visited[k] = m.gen
		in := &m.insts[pc]
		switch in.op {
		case opMatch:
			return pos, true
		case opJmp:
			pc = in.x
		case opSplit:
			if end, ok := m.run(in.x, pos); ok {
				return end, true
			}
			pc = in.y
		case opSave:
			role, start := in.slot/2, in.slot%2 == 0
			oldCap, oldSpan := m.caps[in.slot], m.spans[role]
			m.caps[in.slot] = pos
			if start {
				m.spans[role] = -1
			}
			if end, ok := m.run(pc+1, pos); ok {
				return end, true
			}
			m.caps[in.slot], m.spans[role] = oldCap, oldSpan
			return 0, false
		case opElem:
			if pos >= m.n {
				return 0, false
			}
			sel := in.sel
			if sel.not != nil && m.matches(sel.not, pos) {
				return 0, false
			}
			if sel.spans() {
				for _, sp := range m.l.SpansAt(pos) {
					if sp.End <= pos || sp.End > m.n || !sel.matchSpan(sp) || !m.wordsOK(sel, pos, sp.End) {
						continue
					}
					old := -1
					if in.slot >= 0 {
						old, m.spans[in.slot] = m.spans[in.slot], sp.ID
					}
					m.used = append(m.used, sp.ID)
					if end, ok := m.run(pc+1, sp.End); ok {
						return end, true
					}
					m.used = m.used[:len(m.used)-1]
					if in.slot >= 0 {
						m.spans[in.slot] = old
					}
				}
				return 0, false
			}
			if !sel.cond.ok(m.l.Term(pos)) {
				return 0, false
			}
			pc, pos = pc+1, pos+1
		}
	}
}

// wordsOK: every word of terms [from, to) satisfies s's term conditions.
func (m *matcher) wordsOK(s *selector, from, to int) bool {
	if s.cond.empty() {
		return true
	}
	for i := from; i < to; i++ {
		if t := m.l.Term(i); t.Token.Kind == textnorm.TokenWord && !s.cond.ok(t) {
			return false
		}
	}
	return true
}

// matches reports whether s — a not-condition — would match at pos: a
// span of one of its types starts here (and its words satisfy the term
// conditions), or, without types, the term satisfies them.
func (m *matcher) matches(s *selector, pos int) bool {
	if !s.spans() {
		return s.cond.ok(m.l.Term(pos))
	}
	for _, sp := range m.l.SpansAt(pos) {
		if sp.End > pos && sp.End <= m.n && s.matchSpan(sp) && m.wordsOK(s, pos, sp.End) {
			return true
		}
	}
	return false
}
