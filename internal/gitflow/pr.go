package gitflow

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/jeethsoni/devgod-cli/internal/ai"
	"github.com/jeethsoni/devgod-cli/internal/ui"
)

type PRSizeStats struct {
	FilesChanged int
	LinesAdded   int
	LinesDeleted int
}

// PRSizeStatsBetween computes how many files and lines changed between base and head.
func PRSize(baseBranch, headBranch string) (*PRSizeStats, error) {
	cmd := exec.Command("git", "diff", "--numstat", fmt.Sprintf("%s..%s", baseBranch, headBranch))
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run git diff --numstat: %w", err)
	}

	stats := &PRSizeStats{}
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 3 {
			continue
		}

		added := parseNumstatField(parts[0])
		deleted := parseNumstatField(parts[1])

		stats.FilesChanged++
		stats.LinesAdded += added
		stats.LinesDeleted += deleted
	}

	return stats, nil
}

func parseNumstatField(val string) int {
	val = strings.TrimSpace(val)
	if val == "" || val == "-" {
		return 0
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return n
}

// DiffSummary returns a name-status summary between base and head,
// similar to "git diff --name-status base..head".
func DiffSummary(baseBranch, headBranch string) (string, error) {
	cmd := exec.Command("git", "diff", "--name-status", fmt.Sprintf("%s..%s", baseBranch, headBranch))
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run git diff --name-status: %w", err)
	}
	return string(out), nil
}

// CreatePR generates PR metadata and creates a GitHub PR using gh.
func CreatePR() error {
	if !IsGitRepo() {
		return fmt.Errorf("not inside a git repo")
	}

	// Ensure gh is present and authenticated
	if err := ensureGitHubCLIInstalled(); err != nil {
		return err
	}
	if err := ensureGHAuthenticated(); err != nil {
		return err
	}

	// Load devgod state to get intent + branch
	state, err := LoadState()
	if err != nil {
		return err
	}
	if state.ActiveTask == nil {
		return fmt.Errorf("no active task found. Run `devgod git \"your intent\"` first")
	}

	branch, err := CurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	baseBranch := "main" // could be configurable later

	// 🔍 1. Compute PR size stats
	stats, err := PRSize(baseBranch, branch)
	if err != nil {
		return fmt.Errorf("failed to compute PR size: %w", err)
	}
	totalLines := stats.LinesAdded + stats.LinesDeleted

	fmt.Println()
	fmt.Println("📏 PR size check")
	fmt.Printf("   Files changed: %d\n", stats.FilesChanged)
	fmt.Printf("   Lines added:   %d\n", stats.LinesAdded)
	fmt.Printf("   Lines deleted: %d\n", stats.LinesDeleted)
	fmt.Printf("   Total changes: %d lines\n", totalLines)
	fmt.Println()

	// Best-practice thresholds
	const (
		idealLinesMax = 50
		softFilesMax  = 10
		softLinesMax  = 200
		hardFilesMax  = 20
		hardLinesMax  = 400
	)

	// Hard block: way too big
	if stats.FilesChanged > hardFilesMax || totalLines > hardLinesMax {
		fmt.Println(ui.Red("❌ This PR is very large according to best practices."))
		fmt.Printf("   Recommended: <= %d files, <= %d lines of changes.\n", softFilesMax, softLinesMax)
		fmt.Printf("   Current: %d files, %d lines.\n", stats.FilesChanged, totalLines)
		fmt.Println()
		fmt.Println("Please split this work into smaller PRs (for example feature slices or logical chunks) and try again.")
		return fmt.Errorf("PR too large; creation blocked by devgod size guard")
	}

	// Soft warning: above recommended but not crazy
	if stats.FilesChanged > softFilesMax || totalLines > softLinesMax {
		fmt.Println(ui.Yellow("⚠️ This PR is larger than recommended."))
		fmt.Printf("   Ideal: around %d lines and as few files as possible.\n", idealLinesMax)
		fmt.Printf("   Recommended maximum: %d files, %d lines of changes.\n", softFilesMax, softLinesMax)
		fmt.Printf("   Current: %d files, %d lines.\n", stats.FilesChanged, totalLines)
		fmt.Println()

		if !ui.Confirm("Proceed with creating this larger PR anyway?") {
			fmt.Println(ui.Red("PR creation cancelled. Consider splitting into smaller PRs."))
			return nil
		}
	} else {
		// Tiny/ideal PR
		fmt.Println(ui.Green("✅ PR size looks good (within recommended range)."))
		fmt.Println()
	}

	// 🔁 2. Build context for AI (summary keeps it concise)
	summary, err := DiffSummary(baseBranch, branch)
	if err != nil {
		return fmt.Errorf("failed to compute diff summary: %w", err)
	}

	// 🧠 3. Ask AI for PR title + body
	stop := ui.StartSpinner("Asking the PR gods to write your title & description...")
	meta, err := ai.GeneratePRMetadata(state.ActiveTask.Intent, summary, branch, baseBranch)
	stop()
	if err != nil {
		return fmt.Errorf("failed to generate PR metadata with AI: %w", err)
	}

	// 4. Reviewers selection
	reviewers, err := getReviewersOrAsk()
	if err != nil {
		return fmt.Errorf("failed to select reviewers: %w", err)
	}

	// 5. Show a preview BEFORE hitting GitHub
	fmt.Println("🚀 devgod PR preview")
	fmt.Println("─────────────────────────────────────────────")
	fmt.Println("Branch:")
	fmt.Printf("   %s\n\n", branch)
	fmt.Println("Base branch:")
	fmt.Printf("   %s\n\n", baseBranch)
	fmt.Println("PR size:")
	fmt.Printf("   %d files, +%d/-%d (~%d lines)\n\n", stats.FilesChanged, stats.LinesAdded, stats.LinesDeleted, totalLines)
	fmt.Println("Title:")
	fmt.Printf("   %s\n\n", strings.TrimSpace(meta.Title))
	fmt.Println("Description (body):")
	fmt.Println(meta.Body)
	fmt.Println()

	if len(reviewers) > 0 {
		fmt.Println("Reviewers to request:")
		for _, r := range reviewers {
			fmt.Printf("   - @%s\n", r)
		}
		fmt.Println()
	} else {
		fmt.Println("Reviewers:")
		fmt.Println("   (none selected)")
		fmt.Println()
	}

	// Final confirm before creating the PR
	if !ui.Confirm("Create this PR on GitHub?") {
		fmt.Println("❌ PR creation cancelled.")
		return nil
	}

	// 6. Call gh to actually create the PR
	if err := createGitHubPR(branch, baseBranch, meta.Title, meta.Body, reviewers); err != nil {
		return fmt.Errorf("failed to create PR on GitHub: %w", err)
	}

	fmt.Println(ui.Green("✅ PR created successfully on GitHub."))
	return nil
}

// createGitHubPR calls `gh pr create` with the given metadata and reviewers.
func createGitHubPR(headBranch, baseBranch, title, body string, reviewers []string) error {
	args := []string{
		"pr", "create",
		"--head", headBranch,
		"--base", baseBranch,
		"--title", title,
		"--body", body,
	}

	for _, r := range reviewers {
		r = strings.TrimSpace(r)
		if r != "" {
			args = append(args, "--reviewer", r)
		}
	}

	cmd := exec.Command("gh", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		fmt.Println(out.String())
		return err
	}

	return nil
}
