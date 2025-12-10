package ui

import (
	"fmt"
	"strings"
)

type CommitPlan struct {
	Branch        string
	Intent        string
	StagedSummary string
	CommitMessage string
}

func separator() {
	fmt.Println(strings.Repeat("─", 45))
}

func PrintCommitPlan(plan CommitPlan) {
	fmt.Println()

	// Header: one emoji + cyan color, rest is plain
	fmt.Println(Cyan("🚀 devgod commit preview"))
	separator()

	fmt.Println("Branch:")
	fmt.Println("  ", plan.Branch)

	fmt.Println("\nIntent:")
	fmt.Println("  ", plan.Intent)

	fmt.Println("\nChanges staged for commit:")
	if strings.TrimSpace(plan.StagedSummary) == "" {
		fmt.Println("  (none)")
	} else {
		lines := strings.Split(plan.StagedSummary, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			fmt.Println("  ", line)
		}
	}

	fmt.Println("\nProposed commit message:")
	fmt.Println("  ", plan.CommitMessage)

	separator()
	fmt.Println()
}
