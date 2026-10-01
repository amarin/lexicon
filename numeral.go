package lexicon

// NumeralGrammemes returns the grammemes of a noun that agrees with the
// count n in a nominative phrase: «1 уезд» — nomn sing, «2 уезда» — gent
// sing, «5 уездов» — gent plur. The rule goes by the last two digits of |n|:
// 11–14 take gent plur, otherwise the last digit decides (1 — nomn sing,
// 2–4 — gent sing, the rest and 0 — gent plur). Pass the result as want to
// Registry.Inflect. Other cases («о пяти уездах») are not covered: state
// the case and number yourself. The slice is new on every call.
func NumeralGrammemes(n int) []string {
	m := n % 100
	if m < 0 {
		m = -m
	}

	if m >= 11 && m <= 14 {
		return []string{"gent", "plur"}
	}

	switch m % 10 {
	case 1:
		return []string{"nomn", "sing"}
	case 2, 3, 4:
		return []string{"gent", "sing"}
	}

	return []string{"gent", "plur"}
}
