package gitflow

import (
	"strings"

	"github.com/yourname/devgod-cli/internal/shell"
)

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
