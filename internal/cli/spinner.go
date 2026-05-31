package cli

import (
	"fmt"
	"os"
	"time"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// startSpinner animates a labeled spinner on stderr and returns a stop function
// that clears the line. It is a no-op when stderr is not a terminal (so piped
// output, --json, and tests stay clean).
func startSpinner(label string) func() {
	w := os.Stderr
	if !colorEnabled(w) {
		return func() {}
	}
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		t := time.NewTicker(90 * time.Millisecond)
		defer t.Stop()
		i := 0
		for {
			select {
			case <-stop:
				fmt.Fprint(w, "\r\033[K") // return to line start, clear to EOL
				return
			case <-t.C:
				fmt.Fprintf(w, "\r%s %s", spinnerFrames[i%len(spinnerFrames)], label)
				i++
			}
		}
	}()
	return func() {
		close(stop)
		<-done
	}
}
