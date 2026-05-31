package cli

import "testing"

// In tests stderr is not a TTY, so startSpinner returns a no-op; stopping must
// not panic or hang.
func TestSpinnerNoopOffTerminal(t *testing.T) {
	stop := startSpinner("working")
	stop()
}
