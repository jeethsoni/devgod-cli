package gitflow

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ActiveTask struct {
	Intent           string `json:"intent"`
	Branch           string `json:"branch"`
	SuggestedSubject string `json:"suggested_subject"`
}

type RepoState struct {
	ActiveTask *ActiveTask `json:"active_task,omitempty"`
}

func stateFilePath() (string, error) {
	root, err := RepoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, ".git", "devgod-state.json"), nil
}

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

	var state RepoState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}
