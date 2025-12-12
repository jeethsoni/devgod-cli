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
	// Use the same divider style everywhere
	fmt.Println(Divider.Render(strings.Repeat("─", 45)))
}

func PrintCommitPlan(plan CommitPlan) {
	fmt.Println()

	// Title
	fmt.Println(TitleStyle.Render("🚀 DEVGOD COMMIT PREVIEW"))

	// Branch
	fmt.Println("🌿 " + BranchLabelStyle.Render("Branch:"))
	fmt.Println("   " + ValueStyle.Render(plan.Branch))

	// Intent
	fmt.Println()
	fmt.Println("🎯 " + IntentLabelStyle.Render("Intent:"))
	fmt.Println("   " + ValueStyle.Render(plan.Intent))

	// Staged changes
	fmt.Println()
	fmt.Println("📦 " + SectionTitleStyle.Render("Changes staged for commit:"))
	if strings.TrimSpace(plan.StagedSummary) == "" {
		fmt.Println("   " + ValueStyle.Render("(none)"))
	} else {
		lines := strings.Split(plan.StagedSummary, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			// ColorizeStatusLine now uses lipgloss and only colors the status letter
			fmt.Println("   " + ColorizeStatusLine(line))
		}
	}

	// Proposed commit message
	fmt.Println()
	fmt.Println("✍️  " + CommitLabelStyle.Render("Proposed commit message:"))
	fmt.Println("   " + ValueStyle.Render(plan.CommitMessage))

	separator()
	fmt.Println()
}
