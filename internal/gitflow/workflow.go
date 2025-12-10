package gitflow

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jeethsoni/devgod-cli/internal/ai"
	"github.com/jeethsoni/devgod-cli/internal/ui"
)

var hasLetter = regexp.MustCompile(`[A-Za-z]`)

func validateIntent(intent string) error {
	intent = strings.TrimSpace(intent)
	if intent == "" {
		return fmt.Errorf("intent cannot be empty")
	}

	// Must contain at least one letter (block pure numbers like "30390359")
	if !hasLetter.MatchString(intent) {
		return fmt.Errorf("intent must contain at least one letter, got %q", intent)
	}

	// Very short stuff is probably accidental, nudge user to be clearer
	if len([]rune(intent)) < 6 {
		return fmt.Errorf("intent too short; describe what you want to do, e.g. \"fix login crash when password is empty\"")
	}

	return nil
}

// StartTask creates a new branch for the task based on the intent.
func StartTask(intent string) error {
	if !IsGitRepo() {
		return fmt.Errorf("not inside a git repo")
	}

	intent = strings.TrimSpace(intent)

	if err := validateIntent(intent); err != nil {
		return err
	}

	// 🧠 AI branch naming with loading dots
	stop := ui.StartSpinner("Asking AI for branch name")
	branchName, err := ai.GenerateBranchName(intent)
	stop()
	if err != nil {
		return fmt.Errorf("failed to generate branch name: %w", err)
	}

	// Checkout new branch
	if err := CheckoutNewBranch(branchName); err != nil {
		return err
	}

	// Save state
	state := &RepoState{
		ActiveTask: &ActiveTask{
			Intent: intent,
			Branch: branchName,
		},
	}
	if err := SaveState(state); err != nil {
		return err
	}

	fmt.Println("Created branch:", branchName)
	fmt.Println("Now, make your changes and run: devgod git to finish.")
	return nil
}

// FinishTask stages changes, generates commit message, and creates commit.
func FinishTask() error {
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

	// Check if the user is still on the correct branch
	// Check if the user is still on the correct branch
	currentBranch, err := CurrentBranch()
	if err == nil && currentBranch != state.ActiveTask.Branch {
		fmt.Println(ui.Yellow("⚠️ You are NOT on the branch for this task."))
		fmt.Println("Expected branch:", state.ActiveTask.Branch)
		fmt.Println("Current branch: ", currentBranch)
		fmt.Println()

		if !ui.Confirm("Switch to the correct branch now?") {
			fmt.Println(ui.Red("Commit cancelled. Switch to the correct branch and try again."))
			return nil
		}

		// Try to switch branches
		if err := CheckoutBranch(state.ActiveTask.Branch); err != nil {
			fmt.Println(ui.Red("❌ Failed to switch branches automatically."))
			fmt.Println("Please run:")
			fmt.Println("  git checkout", state.ActiveTask.Branch)
			fmt.Println("and then retry.")
			return err
		}

		fmt.Println(ui.Green("✔️ Switched to the correct branch."))
		fmt.Println()
	}

	// Stage changes
	if HasUnstagedChanges() {
		if err := StageAll(); err != nil {
			return err
		}
	}

	diff, err := StagedDiff()
	if err != nil {
		return err
	}

	if strings.TrimSpace(diff) == "" {
		fmt.Println("No staged changes to commit.")
		return nil
	}

	// AI commit message
	commitMsg, err := ai.GenerateCommitMessage(state.ActiveTask.Intent, diff)
	if err != nil {
		fmt.Println(ui.Red("❌ Failed to generate commit message with AI."))
		fmt.Println("Please complete this commit manually using git (e.g. `git commit -m \"...\"`) and then continue your flow.")
		return err
	}

	summary, _ := StagedSummary() // ignore error; not critical for commit

	plan := ui.CommitPlan{
		Branch:        state.ActiveTask.Branch,
		Intent:        state.ActiveTask.Intent,
		StagedSummary: summary,
		CommitMessage: commitMsg,
	}

	// Show plan
	ui.PrintCommitPlan(plan)

	// Ask user before committing
	if !ui.Confirm("Create this commit?") {
		fmt.Println("❌ Commit cancelled.")
		return nil
	}

	if err := Commit(commitMsg); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	fmt.Println("✅ Commit created:")
	fmt.Println(commitMsg)

	return nil
}
