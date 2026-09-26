package rules

import (
	"fmt"
	"strings"
)

// ruleError is a Compile error located in a rule file (decision D15):
// "<file>:<line>: <set>/<rule>: <message>".
type ruleError struct {
	file string // File.name; "" = "<input>"
	line int    // 0 = unknown (a File built in Go)
	set  string // set name, "#<index>" when empty
	rule string // "hint 0", "trigger 1" (v0.3: `pattern "age"`); "" for the set itself
	msg  string
}

// where formats a source position: "<file>:<line>", or "<file>" without a line.
func where(file string, line int) string {
	if line <= 0 {
		return fileLabel(file)
	}
	return fmt.Sprintf("%s:%d", fileLabel(file), line)
}

func (e *ruleError) Error() string {
	var b strings.Builder
	b.WriteString(where(e.file, e.line))
	b.WriteString(": ")
	b.WriteString(e.set)
	if e.rule != "" {
		b.WriteString("/")
		b.WriteString(e.rule)
	}
	b.WriteString(": ")
	b.WriteString(e.msg)
	return b.String()
}
