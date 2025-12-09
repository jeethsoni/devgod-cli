package gitflow

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Struct to hold the active task state (intent, branch, suggested subject)
type ActiveTask struct {
	Intent           string `json:"intent"`
	Branch           string `json:"branch"`
	SuggestedSubject string `json:"suggested_subject"`
}

// Struct to hold the repository state
type RepoState struct {
	ActiveTask *ActiveTask `json:"active_task,omitempty"`
}

// Returns the file path for storing the repo state.
func stateFilePath() (string, error) {
	root, err := RepoRoot()
	if err != nil {
		return "", err
	}
	// Store state in .git/devgod-state.json
	return filepath.Join(root, ".git", "devgod-state.json"), nil
}

// Writes the repository state to a file.
func SaveState(state *RepoState) error {
	path, err := stateFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Loads the repository state from a file.
func LoadState() (*RepoState, error) {
	path, err := stateFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &RepoState{}, nil
		}
		return nil, err
	}

	// Unmarshal JSON data
	var state RepoState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}
