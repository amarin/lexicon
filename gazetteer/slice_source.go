package gazetteer

import "context"

// SliceSource serves a fixed list of entries under a fixed version.
type SliceSource struct {
	name, version string
	entries       []Entry
}

// NewSliceSource returns a source over entries; the slice is not copied.
func NewSliceSource(name, version string, entries []Entry) *SliceSource {
	return &SliceSource{name: name, version: version, entries: entries}
}

// Name implements Source.
func (s *SliceSource) Name() string { return s.name }

// Version implements Source.
func (s *SliceSource) Version(context.Context) (string, error) { return s.version, nil }

// Entries implements Source.
func (s *SliceSource) Entries(ctx context.Context, yield func(Entry) error) error {
	for _, e := range s.entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := yield(e); err != nil {
			return err
		}
	}
	return nil
}
