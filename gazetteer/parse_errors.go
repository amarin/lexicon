package gazetteer

import "fmt"

// ParseErrors is returned by TSVSource.Entries after every good entry was
// yielded. It is not fatal: the builder copies the lines into the report.
type ParseErrors struct {
	Source string
	Lines  []LineError
}

func (e *ParseErrors) Error() string {
	first := e.Lines[0]
	return fmt.Sprintf("gazetteer: source %s: %d bad line(s); first: line %d: %s",
		e.Source, len(e.Lines), first.Line, first.Msg)
}
