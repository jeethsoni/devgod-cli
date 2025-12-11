package gitflow

import (
	"fmt"
	"os/exec"
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

// Returns a summary of staged changes.
func StagedSummary() (string, error) {
	// `git status --short` is familiar and readable
	return shell.Run("git", "status", "--short")
}

func CheckoutBranch(name string) error {
	_, err := shell.Run("git", "checkout", name)
	return err
}

// parseGitHubOwnerRepo extracts "owner" and "repo" from the git remote URL.
// Supports:
// - git@github.com:owner/repo.git
// - https://github.com/owner/repo.git
// - http://github.com/owner/repo.git
func parseGitHubOwnerRepo() (string, string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to read git remote URL: %w", err)
	}

	url := strings.TrimSpace(string(out))

	var repoPath string

	switch {
	case strings.HasPrefix(url, "git@github.com:"):
		repoPath = strings.TrimPrefix(url, "git@github.com:")
	case strings.HasPrefix(url, "https://github.com/"):
		repoPath = strings.TrimPrefix(url, "https://github.com/")
	case strings.HasPrefix(url, "http://github.com/"):
		repoPath = strings.TrimPrefix(url, "http://github.com/")
	default:
		return "", "", fmt.Errorf("origin does not look like a GitHub URL: %s", url)
	}

	// Remove trailing .git if present
	repoPath = strings.TrimSuffix(repoPath, ".git")

	parts := strings.Split(repoPath, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected GitHub repo path: %s", repoPath)
	}

	owner := parts[0]
	repo := parts[1]

	return owner, repo, nil
}
