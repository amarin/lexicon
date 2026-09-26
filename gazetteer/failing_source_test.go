package gazetteer

import "context"

// failingSource is a Source whose Entries always fails with err, used to
// test that the builder keeps previously compiled data on a fatal error.
type failingSource struct{ err error }

func (f failingSource) Name() string                            { return "broken" }
func (f failingSource) Version(context.Context) (string, error) { return "v2", nil }
func (f failingSource) Entries(context.Context, func(Entry) error) error {
	return f.err
}
