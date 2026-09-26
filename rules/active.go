package rules

// Active lists the rules of the sets active for one document, in file and
// set order.
type Active struct {
	Sets     []string
	Hints    []*HintRule
	Triggers []*TriggerRule
}
