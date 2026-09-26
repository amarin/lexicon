package nertest

// Counts are true/false positives and false negatives.
type Counts struct{ TP, FP, FN int }

// Precision is TP/(TP+FP), 1 when nothing was predicted.
func (c Counts) Precision() float64 {
	if c.TP+c.FP == 0 {
		return 1
	}
	return float64(c.TP) / float64(c.TP+c.FP)
}

// Recall is TP/(TP+FN), 1 when nothing was expected.
func (c Counts) Recall() float64 {
	if c.TP+c.FN == 0 {
		return 1
	}
	return float64(c.TP) / float64(c.TP+c.FN)
}

// F1 is the harmonic mean of precision and recall.
func (c Counts) F1() float64 {
	p, r := c.Precision(), c.Recall()
	if p+r == 0 {
		return 0
	}
	return 2 * p * r / (p + r)
}
