package gitflow

import (
	"strings"

	"github.com/jeethsoni/devgod-cli/internal/shell"
)

// Returns the absolute path to the git repo root.
func RepoRoot() (string, error) {
	out, err := shell.Run("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Returns true if the current directory is inside a git repo.
func IsGitRepo() bool {
	_, err := RepoRoot()
	return err == nil
}

// Returns the name of the current git branch.
func CurrentBranch() (string, error) {
	out, err := shell.Run("git", "rev-parse", "--abbrev-ref", "HEAD")
	return strings.TrimSpace(out), err
}

// Checks out a new branch with the given name.
func CheckoutNewBranch(name string) error {
	_, err := shell.Run("git", "checkout", "-b", name)
	return err
}

// Stages all changes in the working directory.
func StageAll() error {
	_, err := shell.Run("git", "add", ".")
	return err
}

// Commits staged changes with the given message.
func Commit(message string) error {
	_, err := shell.Run("git", "commit", "-m", message)
	return err
}

// Returns true if there are unstaged changes in the working directory.
func HasUnstagedChanges() bool {
	out, _ := shell.Run("git", "status", "--porcelain")
	return strings.TrimSpace(out) != ""
}

// Returns the diff of staged changes.
func StagedDiff() (string, error) {
	return shell.Run("git", "diff", "--cached")
}
