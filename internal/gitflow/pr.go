package gitflow

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jeethsoni/devgod-cli/internal/ai"
	"github.com/jeethsoni/devgod-cli/internal/ui"
)

const defaultBaseBranch = "main" // change to "master" or your default if needed

// CreatePR is the main entrypoint for `devgod pr`.
// It:
//   - Ensures you're in a git repo and on the active task branch
//   - Ensures GitHub CLI (gh) is installed (and can install it with consent)
//   - Ensures the user is authenticated with `gh auth login` (and can run it for them)
//   - Generates a PR title/body/reviewers using AI (Ollama)
//   - Lets you select reviewers from real repo collaborators/teams via `gh api`
//   - Shows a PR preview and asks for confirmation
//   - Calls `gh pr create` to actually create the PR on GitHub
func CreatePR() error {
	if !IsGitRepo() {
		return fmt.Errorf("not inside a git repo")
	}

	state, err := LoadState()
	if err != nil {
		return err
	}
	if state.ActiveTask == nil {
		return fmt.Errorf("no active task found")
	}

	taskBranch := state.ActiveTask.Branch

	currentBranch, err := CurrentBranch()
	if err != nil {
		return err
	}

	// Ensure we are on the task branch
	if currentBranch != taskBranch {
		fmt.Println(ui.Yellow("⚠️ You are NOT on the branch for this task."))
		fmt.Println("Expected branch:", taskBranch)
		fmt.Println("Current branch:", currentBranch)
		fmt.Println()

		if !ui.Confirm("Switch to the correct branch now?") {
			fmt.Println(ui.Red("PR flow cancelled. Switch to the correct branch and try again."))
			return nil
		}

		if err := CheckoutBranch(taskBranch); err != nil {
			fmt.Println(ui.Red("❌ Failed to switch branches automatically."))
			fmt.Println("Please run:")
			fmt.Println("  git checkout", taskBranch)
			fmt.Println("and then retry.")
			return err
		}

		fmt.Println(ui.Green("✔️ Switched to the correct branch."))
		fmt.Println()
	}

	// 1️⃣ Ensure GitHub CLI exists (and optionally install it)
	if !ghInstalled() {
		fmt.Println(ui.Yellow("⚠️ GitHub CLI (gh) is not installed."))

		if ui.Confirm("Would you like devgod to install GitHub CLI for you?") {
			fmt.Println(ui.Green("✔️ Installing GitHub CLI…"))
			if err := installGH(); err != nil {
				fmt.Println(ui.Red("❌ GitHub CLI installation failed:"))
				return err
			}
			fmt.Println(ui.Green("✔️ GitHub CLI installed successfully!"))
		} else {
			fmt.Println(ui.Red("❌ Cannot continue PR flow without GitHub CLI."))
			return nil
		}
	}

	// 2️⃣ 🔐 Ensure the user is authenticated with GitHub CLI.
	fmt.Println(ui.Yellow("Checking GitHub CLI authentication…"))
	if err := ensureGHAuthenticated(); err != nil {
		// ensureGHAuthenticated already prints what happened
		return nil
	}
	// 3️⃣ Now it's safe to continue with git diff + AI, because gh is installed AND logged in.

	// Diff vs base branch for PR context
	diff, err := diffAgainstBase(defaultBaseBranch)
	if err != nil {
		return err
	}
	if strings.TrimSpace(diff) == "" {
		fmt.Println(ui.Yellow("No changes detected between base and this branch."))
		fmt.Println("You can still create a PR manually if needed.")
		return nil
	}

	// Ask AI (Ollama) to generate PR metadata (title, body, reviewers)
	stop := ui.StartSpinner("Asking the dev gods to craft your PR…")
	meta, err := ai.GeneratePRMetadata(state.ActiveTask.Intent, diff, taskBranch, defaultBaseBranch)
	stop()
	if err != nil {
		fmt.Println(ui.Red("❌ Failed to generate PR metadata with AI."))
		return err
	}

	// Try to fetch actual collaborators/teams from GitHub for reviewer selection
	var chosenReviewers []string
	availableReviewers, err := getReviewers()
	if err != nil {
		fmt.Println(ui.Yellow("⚠️ Could not fetch reviewers from GitHub:"), err)
		fmt.Println("Using AI-suggested reviewers only (if any).")
		chosenReviewers = meta.Reviewers
	} else if len(availableReviewers) == 0 {
		fmt.Println(ui.Yellow("⚠️ No collaborators/teams found for this repo."))
		chosenReviewers = meta.Reviewers
	} else {
		fmt.Println()
		fmt.Println("📌 Select reviewers for this PR (collaborators/teams from this repo):")
		chosen := ui.SelectMultiple("Available reviewers:", availableReviewers)
		if len(chosen) == 0 {
			fmt.Println(ui.Yellow("No reviewers selected. Falling back to AI suggestions (if any)."))
			chosenReviewers = meta.Reviewers
		} else {
			chosenReviewers = chosen
		}
	}

	// Override reviewers with chosen ones (if we have any)
	if len(chosenReviewers) > 0 {
		meta.Reviewers = chosenReviewers
	}

	// Show PR preview with final reviewers list
	printPRPreview(taskBranch, defaultBaseBranch, meta.Title, meta.Body, meta.Reviewers)

	// Confirm before actually creating PR
	if !ui.Confirm("Create this PR on GitHub?") {
		fmt.Println(ui.Red("PR flow cancelled."))
		return nil
	}

	// Make sure branch is pushed before creating PR
	stop = ui.StartSpinner("Pushing branch to origin…")
	if err := gitPushBranch(taskBranch); err != nil {
		stop()
		return err
	}
	stop()

	// Create PR via GitHub CLI
	stop = ui.StartSpinner("Creating PR on GitHub via gh…")
	if err := ghCreatePR(meta.Title, meta.Body, taskBranch, defaultBaseBranch, meta.Reviewers); err != nil {
		stop()
		fmt.Println(ui.Red("❌ Failed to create PR via GitHub CLI:"))
		return err
	}
	stop()

	return nil
}

// diffAgainstBase returns `git diff baseBranch...HEAD` as a string.
func diffAgainstBase(baseBranch string) (string, error) {
	cmd := exec.Command("git", "diff", fmt.Sprintf("%s...HEAD", baseBranch))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git diff %s...HEAD failed: %w\n%s", baseBranch, err, string(out))
	}
	return string(out), nil
}

// gitPushBranch runs `git push -u origin <branch>`.
// It assumes you're already on <branch>.
func gitPushBranch(branch string) error {
	cmd := exec.Command("git", "push", "-u", "origin", branch)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(ui.Red("❌ git push failed:"))
		fmt.Println(string(out))
		return err
	}
	return nil
}

// ghCreatePR uses the installed GitHub CLI to create a PR.
// It passes title, body, base, head, and reviewers.
func ghCreatePR(title, body, head, base string, reviewers []string) error {
	args := []string{
		"pr", "create",
		"--title", title,
		"--body", body,
		"--base", base,
		"--head", head,
	}

	// Add reviewers if present
	for _, r := range sanitizeReviewers(reviewers) {
		args = append(args, "--reviewer", r)
	}

	cmd := exec.Command("gh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// sanitizeReviewers strips leading '@' and trims spaces.
func sanitizeReviewers(raw []string) []string {
	var out []string
	for _, r := range raw {
		r = strings.TrimSpace(r)
		r = strings.TrimPrefix(r, "@")
		if r != "" {
			out = append(out, r)
		}
	}
	return out
}

// printPRPreview prints a nice-looking preview of the PR before creation.
func printPRPreview(branch, base, title, body string, reviewers []string) {
	fmt.Println("🚀 devgod PR preview")
	fmt.Println("─────────────────────────────────────────────")
	fmt.Println("Branch:")
	fmt.Printf("   %s\n\n", branch)

	fmt.Println("Base branch:")
	fmt.Printf("   %s\n\n", base)

	fmt.Println("Title:")
	fmt.Printf("   %s\n\n", title)

	fmt.Println("Description (body):")
	fmt.Println(body)
	fmt.Println()

	if len(reviewers) > 0 {
		fmt.Println("Reviewers:")
		for _, r := range reviewers {
			fmt.Printf("   - %s\n", r)
		}
		fmt.Println()
	}

	fmt.Println("─────────────────────────────────────────────")
}
