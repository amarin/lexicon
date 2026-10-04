package rules

import "errors"

// lineError carries the source line of the pattern element that caused
// err; compileSet turns it into a located ruleError (decision P11).
type lineError struct {
	line int
	err  error
}

func (e *lineError) Error() string { return e.err.Error() }
func (e *lineError) Unwrap() error { return e.err }

// atLine attaches line to err unless err already carries one (an element
// inside a group reports its own line, not the group's).
func atLine(line int, err error) error {
	var le *lineError
	if err == nil || errors.As(err, &le) {
		return err
	}
	return &lineError{line: line, err: err}
}
