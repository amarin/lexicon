package gazetteer

// MaxLemmaKeys caps the lemma keys generated for one alias.
const MaxLemmaKeys = 8

// Combine returns the cartesian product of alts in odometer order (the last
// position changes fastest), at most limit sequences. capped reports that
// further sequences existed. A position without alternatives yields nil.
func Combine(alts [][]string, limit int) (out [][]string, capped bool) {
	if len(alts) == 0 {
		return nil, false
	}
	for _, a := range alts {
		if len(a) == 0 {
			return nil, false
		}
	}
	idx := make([]int, len(alts))
	for {
		if len(out) == limit {
			return out, true
		}
		seq := make([]string, len(alts))
		for i, a := range alts {
			seq[i] = a[idx[i]]
		}
		out = append(out, seq)
		i := len(alts) - 1
		for ; i >= 0; i-- {
			idx[i]++
			if idx[i] < len(alts[i]) {
				break
			}
			idx[i] = 0
		}
		if i < 0 {
			return out, false
		}
	}
}
