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

// PRSize computes how many files and lines changed between base and head.
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

	baseBranch, err := selectBaseBranchInteractive()
	if err != nil {
		return fmt.Errorf("failed to choose base branch: %w", err)
	}

	// Compute PR size stats
	stats, err := PRSize(baseBranch, branch)
	if err != nil {
		return fmt.Errorf("failed to compute PR size: %w", err)
	}
	totalLines := stats.LinesAdded + stats.LinesDeleted

	fmt.Println()

	// Best-practice thresholds
	const (
		idealLinesMax = 50
		softFilesMax  = 10
		softLinesMax  = 200
		hardFilesMax  = 20
		hardLinesMax  = 400
	)

	// for commit way too big
	if stats.FilesChanged > hardFilesMax || totalLines > hardLinesMax {
		fmt.Println(ui.Red("❌ This PR is very large according to best practices."))
		fmt.Printf("   Recommended: <= %d files, <= %d lines of changes.\n", softFilesMax, softLinesMax)
		fmt.Printf("   Current: %d files, %d lines.\n", stats.FilesChanged, totalLines)
		fmt.Println()
		fmt.Println("Please split this work into smaller PRs (for example feature slices or logical chunks) and try again.")
		return fmt.Errorf("PR too large; creation blocked by devgod size guard")
	}

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
	}

	// Build context for AI (summary keeps it concise)
	summary, err := DiffSummary(baseBranch, branch)
	if err != nil {
		return fmt.Errorf("failed to compute diff summary: %w", err)
	}

	// Ask AI for PR title + body
	stop := ui.StartSpinner("🪄 Asking the PR gods to write your title & description...")
	meta, err := ai.GeneratePRMetadata(state.ActiveTask.Intent, summary, branch, baseBranch)
	stop()
	if err != nil {
		return fmt.Errorf("failed to generate PR metadata with AI: %w", err)
	}

	// Reviewers selection
	reviewers, err := getReviewersOrAsk()
	if err != nil {
		return fmt.Errorf("failed to select reviewers: %w", err)
	}
	// Show a preview before hitting GitHub
	fmt.Println()
	fmt.Println(ui.TitleStyle.Render("🚀 DEVGOD PR PREVIEW"))

	fmt.Println("🌿 " + ui.BranchLabelStyle.Render("Branch:"))
	fmt.Println("   " + ui.ValueStyle.Render(branch))
	fmt.Println()

	fmt.Println("🧱 " + ui.SectionTitleStyle.Render("Base branch:"))
	fmt.Println("   " + ui.ValueStyle.Render(baseBranch))
	fmt.Println()

	fmt.Println("📝 " + ui.SectionTitleStyle.Render("Title:"))
	fmt.Println("   " + ui.ValueStyle.Render(strings.TrimSpace(meta.Title)))
	fmt.Println()

	fmt.Println("📄 " + ui.SectionTitleStyle.Render("Description:"))
	// Keep body formatting; just indent it
	body := strings.TrimRight(meta.Body, "\n")
	for _, line := range strings.Split(body, "\n") {
		if strings.TrimSpace(line) == "" {
			fmt.Println()
			continue
		}
		fmt.Println("   " + ui.ValueStyle.Render(line))
	}
	fmt.Println()

	if len(reviewers) > 0 {
		fmt.Println("👥 " + ui.SectionTitleStyle.Render("Reviewers to request:"))
		for _, r := range reviewers {
			fmt.Println("   " + ui.CyanStyle.Render("@"+r))
		}
		fmt.Println()
	} else {
		fmt.Println("👥 " + ui.SectionTitleStyle.Render("Reviewers:"))
		fmt.Println("   " + ui.ValueStyle.Render("(none selected)"))
		fmt.Println()
	}

	fmt.Println(ui.Divider.Render(strings.Repeat("─", 45)))
	fmt.Println()

	// Final confirm before creating the PR
	if !ui.Confirm("Create this PR on GitHub?") {
		fmt.Println("❌ PR creation cancelled.")
		return nil
	}

	// Call gh to actually create the PR
	if err := createGitHubPR(baseBranch, meta.Title, meta.Body, reviewers); err != nil {
		return fmt.Errorf("failed to create PR on GitHub: %w", err)
	}

	fmt.Println(ui.Green("✅ PR created successfully on GitHub."))
	return nil
}

func createGitHubPR(baseBranch, title, body string, reviewers []string) error {
	args := []string{
		"pr", "create",
		"--title", title,
		"--body", body,
	}

	if baseBranch != "" {
		args = append(args, "--base", baseBranch)
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
