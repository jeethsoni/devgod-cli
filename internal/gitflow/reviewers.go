package gitflow

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type githubUser struct {
	Login string `json:"login"`
}

type githubTeam struct {
	Slug string `json:"slug"`
}

// getReviewers fetches collaborators + teams for the current GitHub repo
// using GitHub CLI's automatic repo detection.
//
// No need for owner/repo arguments because `gh api repos/:owner/:repo/...`
// lets GitHub CLI infer everything from the local git remote.
func getReviewers() ([]string, error) {
	var reviewers []string

	//
	// 1️⃣ Fetch collaborators (users)
	//
	collabCmd := exec.Command("gh", "api", "repos/:owner/:repo/collaborators")
	collabOut, err := collabCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repo collaborators: %w", err)
	}

	var users []githubUser
	if err := json.Unmarshal(collabOut, &users); err != nil {
		return nil, fmt.Errorf("failed to parse collaborators JSON: %w", err)
	}

	for _, u := range users {
		if u.Login != "" {
			reviewers = append(reviewers, u.Login)
		}
	}

	//
	// 2️⃣ Fetch teams (optional, ignore errors)
	//
	teamCmd := exec.Command("gh", "api", "repos/:owner/:repo/teams", "--json", "slug")
	teamOut, err := teamCmd.Output()
	if err == nil {
		var teams []githubTeam
		if err := json.Unmarshal(teamOut, &teams); err == nil {
			for _, t := range teams {
				if t.Slug != "" {
					reviewers = append(reviewers, t.Slug)
				}
			}
		}
	}

	return reviewers, nil
}
