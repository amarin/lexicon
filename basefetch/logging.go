package basefetch

import (
	"fmt"
	"sync"

	"github.com/amarin/logging"
)

// loggingOnce guards ensureLogging so the backend probe/init runs at most
// once per process.
var loggingOnce sync.Once

// loggingOnceErr holds the error (if any) from the single ensureLogging
// attempt, replayed to every caller.
var loggingOnceErr error

// ensureLogging makes sure a github.com/amarin/logging backend exists before
// gomorphy's pymorphy.NewLoader calls logging.NewNamedLogger, which panics
// with "logging: set backend first" when no backend was ever installed.
//
// It never replaces a backend the host application already set up: it first
// probes whether logging.NewNamedLogger already works (recovering the panic
// if not) and only calls logging.Init when the probe shows no backend is
// present.
func ensureLogging() error {
	loggingOnce.Do(func() {
		if backendPresent() {
			return
		}

		if err := logging.Init(
			logging.WithLevel(logging.LevelWarn),
			logging.WithTarget(logging.StdErr),
			logging.WithFormat(logging.FormatText),
		); err != nil {
			loggingOnceErr = fmt.Errorf("basefetch: init logging backend: %w", err)
		}
	})

	return loggingOnceErr
}

// backendPresent reports whether a github.com/amarin/logging backend is
// already initialised, by probing logging.NewNamedLogger and recovering the
// panic it raises when no backend is set.
func backendPresent() (present bool) {
	defer func() {
		if recover() != nil {
			present = false
		}
	}()

	logging.NewNamedLogger("basefetch-probe")

	return true
}
