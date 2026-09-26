package gazetteer

// Snapshot is an immutable set of compiled sources. It is safe for
// concurrent use; Gazetteer publishes a new one on every successful refresh.
type Snapshot struct {
	sources []*compiledSource
	version string
}
