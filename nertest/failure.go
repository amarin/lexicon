package nertest

// Failure is one strict mismatch: a missed gold span or a spurious prediction.
type Failure struct {
	Case string
	Kind string // "missed" or "spurious"
	Type string
	Text string
}
