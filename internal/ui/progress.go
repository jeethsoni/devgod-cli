package ui

import (
	"fmt"
	"strings"
	"time"
)

// StartSpinner shows "message..." with animated dots while some work is running.
// It returns a stop function you MUST call when the work is done.
func StartSpinner(message string) func() {
	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		dots := 0
		for {
			select {
			case <-done:
				// Clear the line and move to next
				fmt.Print("\r")
				fmt.Println(strings.Repeat(" ", len(message)+4))
				fmt.Print("\r")
				return
			case <-ticker.C:
				dots = (dots + 1) % 4 // 0,1,2,3
				fmt.Printf("\r%s%s", message, strings.Repeat(".", dots))
			}
		}
	}()

	// stop function
	return func() {
		close(done)
	}
}
