package gitflow

import (
	"strings"

	"github.com/jeethsoni/devgod-cli/internal/shell"
)

// RepoRoot returns the absolute path to the git repo root.
func RepoRoot() (string, error) {
	out, err := shell.Run("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// IsGitRepo returns true if the current directory is inside a git repo.
func IsGitRepo() bool {
	_, err := RepoRoot()
	return err == nil
}

func CurrentBranch() (string, error) {
	out, err := shell.Run("git", "rev-parse", "--abbrev-ref", "HEAD")
	return strings.TrimSpace(out), err
}

func CheckoutNewBranch(name string) error {
	_, err := shell.Run("git", "checkout", "-b", name)
	return err
}

func StageAll() error {
	_, err := shell.Run("git", "add", ".")
	return err
}

func Commit(message string) error {
	_, err := shell.Run("git", "commit", "-m", message)
	return err
}

func HasUnstagedChanges() bool {
	out, _ := shell.Run("git", "status", "--porcelain")
	return strings.TrimSpace(out) != ""
}

func StagedDiff() (string, error) {
	return shell.Run("git", "diff", "--cached")
}
