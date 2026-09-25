package basefetch

import (
	"os"
	"os/exec"
	"testing"

	"github.com/amarin/gomorphy/pkg/pymorphy"
	"github.com/amarin/logging"
)

// TestEnsureLoggingInitializesBackend reproduces the "logging: set backend
// first" panic that pymorphy.NewLoader triggers via
// github.com/amarin/logging.NewNamedLogger when no logging backend was ever
// installed (the basefetch test binary has no other logging init, so this is
// a faithful repro). ensureLogging must make the construction below safe.
func TestEnsureLoggingInitializesBackend(t *testing.T) {
	if err := ensureLogging(); err != nil {
		t.Fatalf("ensureLogging: %v", err)
	}

	// Must not panic: before the fix this call panicked with
	// "logging: set backend first".
	_ = pymorphy.NewLoader(t.TempDir())
}

// TestEnsureLoggingKeepsHostBackend documents that ensureLogging must never
// replace a backend the host application already initialised. This can't be
// verified in-process alongside TestEnsureLoggingInitializesBackend: the
// github.com/amarin/logging backend is process-global and Once-guarded, and
// Go doesn't guarantee test execution order across files in a package, so
// whichever of the two tests runs first would decide whether a backend is
// already present for the other. It re-execs this test binary in a fresh
// subprocess so the "host already set a backend" scenario starts from a
// clean, unshared global state.
func TestEnsureLoggingKeepsHostBackend(t *testing.T) {
	if os.Getenv("BASEFETCH_KEEP_HOST_BACKEND_HELPER") == "1" {
		hostBackendHelper(t)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestEnsureLoggingKeepsHostBackend", "-test.v")
	cmd.Env = append(os.Environ(), "BASEFETCH_KEEP_HOST_BACKEND_HELPER=1")

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("subprocess failed: %v\n%s", err, out)
	}

	t.Logf("subprocess output:\n%s", out)
}

// hostBackendHelper runs inside the re-exec'd subprocess. It sets up a
// backend the way a host application would, at a level ensureLogging's own
// defaults would never pick (LevelError, vs. ensureLogging's LevelWarn), then
// calls ensureLogging and checks the host's level survived untouched.
func hostBackendHelper(t *testing.T) {
	if err := logging.Init(
		logging.WithLevel(logging.LevelError),
		logging.WithTarget(logging.StdErr),
		logging.WithFormat(logging.FormatText),
	); err != nil {
		t.Fatalf("host logging.Init: %v", err)
	}

	if err := ensureLogging(); err != nil {
		t.Fatalf("ensureLogging: %v", err)
	}

	got := logging.NewNamedLogger("host-check").Level()
	if got != logging.LevelError {
		t.Fatalf("ensureLogging replaced the host backend: logger level = %v, want %v (host's)", got, logging.LevelError)
	}
}
