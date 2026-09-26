package ner

// Weights parameterize span scores (decision D6 of the v0.2 plan):
// score = (origin + Types[type]) × words + LengthBonus × (words − 1)
// + evidence − AmbiguityPenalty × (alternatives − 1).
type Weights struct {
	Surface, Lemma, Trigger float32 // per word, by how the span was found
	LengthBonus             float32 // per word beyond the first
	Types                   map[string]float32
	AmbiguityPenalty        float32 // per extra ref/normal form
}

// DefaultWeights returns the weights used when Config.Weights is zero.
func DefaultWeights() Weights {
	return Weights{Surface: 3, Lemma: 2, Trigger: 1, LengthBonus: 0.5, AmbiguityPenalty: 0.25}
}
