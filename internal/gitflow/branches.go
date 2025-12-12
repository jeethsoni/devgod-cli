// internal/gitflow/branches.go
package gitflow

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/jeethsoni/devgod-cli/internal/ui"
)

// getRemoteBranches fetches all branch names from the GitHub repo using gh api.
func getRemoteBranches() ([]string, error) {
	owner, repo, err := parseGitHubOwnerRepo()
	if err != nil {
		return nil, fmt.Errorf("could not determine GitHub owner/repo from git remote: %w", err)
	}

	cmd := exec.Command(
		"gh", "api",
		fmt.Sprintf("repos/%s/%s/branches", owner, repo),
		"--paginate",
		"--jq", ".[].name",
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to fetch remote branches: %w\n%s", err, out.String())
	}

	var branches []string
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name != "" {
			branches = append(branches, name)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read branches output: %w", err)
	}

	if len(branches) == 0 {
		return nil, fmt.Errorf("no branches found on remote")
	}

	// Sort, but prioritize common "base" branches at the top.
	preferredOrder := []string{"main", "master", "develop", "dev", "qa", "staging"}
	sort.Strings(branches)

	seen := make(map[string]bool)
	var ordered []string

	// First: put preferred branches in order if they exist
	for _, p := range preferredOrder {
		for _, b := range branches {
			if b == p && !seen[b] {
				ordered = append(ordered, b)
				seen[b] = true
			}
		}
	}
	// Then: everything else
	for _, b := range branches {
		if !seen[b] {
			ordered = append(ordered, b)
			seen[b] = true
		}
	}

	return ordered, nil
}

// selectBaseBranchInteractive shows all remote branches and lets the user
// choose which one to use as the PR base.
func selectBaseBranchInteractive() (string, error) {
	branches, err := getRemoteBranches()
	if err != nil {
		return "", err
	}

	fmt.Println()
	fmt.Println(ui.Green("Available base branches (from GitHub):"))
	for i, b := range branches {
		fmt.Printf("  %2d) %s\n", i+1, b)
	}
	fmt.Println()

	selected, err := ui.SelectOne(branches, (ui.Cyan("Select base branch by number:")))
	if err != nil {
		return "", err
	}

	fmt.Println()

	fmt.Println(ui.Green("✔️ Base branch:"), selected)
	return selected, nil
}
