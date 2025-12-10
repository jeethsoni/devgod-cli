package ui

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"

	spin "github.com/tj/go-spin"
)

// StartSpinner starts an animated CLI spinner with the given message and
// returns a func you can call to stop it.
//
// Usage:
//
//	branchName, err := ai.GenerateBranchName(intent)
//	stop()
func StartSpinner(message string) func() {
	s := &spinnerState{
		message: message,
		done:    make(chan struct{}),
	}

	go s.loop()

	// Return stop function that is safe to call once
	return func() {
		if atomic.CompareAndSwapInt32(&s.stopped, 0, 1) {
			close(s.done)
			// Clear the spinner line so the next print starts clean
			fmt.Fprint(os.Stdout, "\r\033[K")
		}
	}
}

type spinnerState struct {
	message string
	done    chan struct{}
	stopped int32
}

func (s *spinnerState) loop() {
	sp := spin.New()
	for {
		select {
		case <-s.done:
			return
		default:
			frame := sp.Next()
			// \r = carriage return to start of line (overwrite in place)
			fmt.Fprintf(os.Stdout, "\r%s %s", frame, s.message)
			time.Sleep(80 * time.Millisecond)
		}
	}
}
