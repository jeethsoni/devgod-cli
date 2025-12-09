package gitflow

import (
	"fmt"

	"github.com/jeethsoni/devgod-cli/internal/ai"
)

func StartTask(intent string) error {
	if !IsGitRepo() {
		return fmt.Errorf("not inside a git repo")
	}

	// Generate AI branch name
	branchName, err := ai.GenerateBranchName(intent)
	if err != nil {
		return err
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
	fmt.Println("Now, make your changes and run: devgod-cli git to finish.")
	return nil
}

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

	// Generate AI commit message
	commitMsg, err := ai.GenerateCommitMessage(state.ActiveTask.Intent, diff)
	if err != nil {
		return err
	}

	// Commit changes
	if err := Commit(commitMsg); err != nil {
		return err
	}

	fmt.Println("Committed changes with message:", commitMsg)
	return nil
}
