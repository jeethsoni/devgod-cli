package gitflow

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jeethsoni/devgod-cli/internal/ai"
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

	branchName, err := ai.GenerateBranchName(intent)
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

	// Generate AI commit message
	commitMsg, err := ai.GenerateCommitMessage(state.ActiveTask.Intent, diff)
	if err != nil {
		return fmt.Errorf("failed to generate commit message: %w", err)
	}

	if err := Commit(commitMsg); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	fmt.Println("✅ Commit created:")
	fmt.Println(commitMsg)

	return nil
}
