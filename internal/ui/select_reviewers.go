package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// SelectMultiple lets the user choose multiple items by numeric index.
// Example:
//
//	items:  ["alice", "bob", "carol"]
//	prompt: "Select reviewers..."
//
// User input: "1,3" -> returns ["alice", "carol"]
func SelectMultiple(items []string, prompt string) ([]string, error) {
	if len(items) == 0 {
		return []string{}, nil
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println(prompt)
		fmt.Print("> ")

		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read input: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			// User chose nothing
			return []string{}, nil
		}

		parts := strings.Split(line, ",")
		var indices []int
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			n, err := strconv.Atoi(p)
			if err != nil || n < 1 || n > len(items) {
				fmt.Println(Red("Invalid selection:"), p)
				fmt.Printf("Please enter numbers between 1 and %d, separated by commas.\n", len(items))
				indices = nil
				break
			}
			indices = append(indices, n-1) // zero-based
		}

		if indices == nil {
			continue
		}

		// Deduplicate while preserving order
		seen := make(map[int]bool)
		var selected []string
		for _, idx := range indices {
			if !seen[idx] {
				seen[idx] = true
				selected = append(selected, items[idx])
			}
		}

		return selected, nil
	}
}
