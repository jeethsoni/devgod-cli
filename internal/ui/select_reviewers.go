package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// SelectMultiple allows the user to pick multiple items from a list.
func SelectMultiple(prompt string, options []string) []string {
	fmt.Println(prompt)

	for i, opt := range options {
		fmt.Printf("%d) %s\n", i+1, opt)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Select reviewers (comma-separated numbers): ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	parts := strings.Split(input, ",")
	var selected []string

	for _, p := range parts {
		p = strings.TrimSpace(p)
		n, err := strconv.Atoi(p)
		if err != nil || n < 1 || n > len(options) {
			continue
		}
		selected = append(selected, options[n-1])
	}

	return selected
}
