package gitflow

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/jeethsoni/devgod-cli/internal/ui"
)

// getReviewers fetches GitHub collaborators for the repo using gh api.
// Returns a list of GitHub usernames (without @).
func getReviewers(owner, repo string) ([]string, error) {
	// Example:
	// gh api repos/:owner/:repo/collaborators --jq '.[].login' --paginate
	cmd := exec.Command(
		"gh", "api",
		fmt.Sprintf("repos/%s/%s/collaborators", owner, repo),
		"--jq", ".[].login",
		"--paginate",
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to fetch repo collaborators: %w", err)
	}

	var reviewers []string
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			reviewers = append(reviewers, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read collaborators output: %w", err)
	}

	return reviewers, nil
}

// getReviewersOrAsk tries to fetch repo collaborators and lets the user
// select zero or more reviewers. Returns an empty slice if none selected.
func getReviewersOrAsk() ([]string, error) {
	owner, repo, err := parseGitHubOwnerRepo()
	if err != nil {
		fmt.Println(ui.Yellow("⚠️ Could not determine GitHub owner/repo from git remote."))
		fmt.Println("Reviewers will not be pre-filled.")
		return []string{}, nil
	}

	allReviewers, err := getReviewers(owner, repo)
	if err != nil {
		fmt.Println(ui.Yellow("⚠️ Could not fetch reviewers from GitHub:"), err)
		fmt.Println("You can still create the PR without reviewers.")
		return []string{}, nil
	}

	if len(allReviewers) == 0 {
		fmt.Println(ui.Yellow("⚠️ No collaborators found for this repo via GitHub."))
		return []string{}, nil
	}

	fmt.Println()
	fmt.Println("Available reviewers (collaborators):")
	for i, r := range allReviewers {
		fmt.Printf("  %2d) %s\n", i+1, r)
	}
	fmt.Println()

	selected, err := ui.SelectMultiple(
		allReviewers,
		"Select reviewers by number (comma-separated, or blank for none):",
	)
	if err != nil {
		return nil, err
	}

	if len(selected) == 0 {
		fmt.Println("No reviewers selected.")
	}

	return selected, nil
}
