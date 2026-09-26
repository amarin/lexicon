package gazetteer

import "sync"

// versionedStub is a stubAnalyzer whose version can change.
type versionedStub struct {
	stubAnalyzer
	mu sync.Mutex
	v  string
}

func (v *versionedStub) Version() string { v.mu.Lock(); defer v.mu.Unlock(); return v.v }
