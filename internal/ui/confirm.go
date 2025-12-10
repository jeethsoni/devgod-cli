package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Confirm asks the user a yes/no question in the terminal.
// Returns true if user answers "y" or "yes" (case-insensitive).
func Confirm(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", prompt)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(Red("Error reading input:"), err)
			return false
		}

		answer := strings.ToLower(strings.TrimSpace(input))

		if answer == "y" || answer == "yes" {
			return true
		}
		if answer == "" || answer == "n" || answer == "no" {
			return false
		}

		fmt.Println("Please type 'y' or 'n'.")
	}
}
