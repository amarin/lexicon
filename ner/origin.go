package ner

// origin records how a candidate was found.
type origin uint8

const (
	originLemma origin = iota + 1
	originSurface
	originTrigger
)

// weight is the per-word origin weight.
func (o origin) weight(w Weights) float64 {
	switch o {
	case originSurface:
		return float64(w.Surface)
	case originLemma:
		return float64(w.Lemma)
	default:
		return float64(w.Trigger)
	}
}
