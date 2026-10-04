package rules

// PatternMatch is one match: terms [Start, End), the captured roles (roles
// inside unused optional elements are absent) and the IDs of every
// candidate a type element consumed, in order.
type PatternMatch struct {
	Start, End int
	Roles      map[string]Capture
	Spans      []int
}
