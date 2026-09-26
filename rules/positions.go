package rules

import "go.yaml.in/yaml/v3"

// fileLabel is how errors name a rule file.
func fileLabel(name string) string {
	if name == "" {
		return "<input>"
	}
	return name
}

// setPositions copies source lines from root — the node tree of the
// document f was decoded from — into f's sets and rules. A rule whose node
// is not found (e.g. brought in by a merge key) gets its set's line.
func setPositions(f *File, root *yaml.Node) {
	doc := deref(root)
	if doc != nil && doc.Kind == yaml.DocumentNode && len(doc.Content) == 1 {
		doc = deref(doc.Content[0])
	}
	sets := child(doc, "sets")
	for i := range f.Sets {
		rs := &f.Sets[i]
		n := item(sets, i)
		rs.line = lineOf(n, 0)
		hints, triggers := child(n, "hints"), child(n, "triggers")
		for j := range rs.Hints {
			rs.Hints[j].line = lineOf(item(hints, j), rs.line)
		}
		for j := range rs.Triggers {
			rs.Triggers[j].line = lineOf(item(triggers, j), rs.line)
		}
	}
}

// deref follows YAML aliases (*name) to their anchored node.
func deref(n *yaml.Node) *yaml.Node {
	for n != nil && n.Kind == yaml.AliasNode {
		n = n.Alias
	}
	return n
}

// child returns the value of key in mapping n, or nil.
func child(n *yaml.Node, key string) *yaml.Node {
	n = deref(n)
	if n == nil || n.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			return deref(n.Content[i+1])
		}
	}
	return nil
}

// item returns element i of sequence n, or nil.
func item(n *yaml.Node, i int) *yaml.Node {
	n = deref(n)
	if n == nil || n.Kind != yaml.SequenceNode || i >= len(n.Content) {
		return nil
	}
	return n.Content[i]
}

func lineOf(n *yaml.Node, fallback int) int {
	if n == nil {
		return fallback
	}
	return n.Line
}
