package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Adjust this to whatever model you're using elsewhere (e.g. "llama3.1", "qwen2.5", etc.)
const prModel = "qwen2.5-coder:7b"

// PRMetadata holds the AI-generated PR title, body, and reviewer suggestions.
type PRMetadata struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Reviewers []string `json:"reviewers"`
}

// GeneratePRMetadata uses Ollama to generate a PR title, body, and suggested reviewers.
//
// intent:     the original task intent (from ActiveTask.Intent).
// diff:       git diff vs base branch for context.
// branch:     feature branch name.
// baseBranch: usually "main".
func GeneratePRMetadata(intent, diff, branch, baseBranch string) (*PRMetadata, error) {
	systemPrompt := `
You are an expert software engineer generating a GitHub Pull Request title and description.

STRICT RULES:
- You MUST base everything ONLY on the provided git diff.
- You MUST NOT invent or assume changes that do not appear in the diff.
- Describe exactly what changed, no more and no less.
- If the diff shows only one file, mention only that file.
- If the diff is small, write a small PR description.
- NEVER mention tests, additional files, or modifications not shown in the diff.
- NEVER include reviewers. The user will select reviewers manually.

OUTPUT FORMAT (MANDATORY):
{
  "title": "<short-title>",
  "body": "<description-based-only-on-the-diff>"
}

TITLE RULES:
- 3–9 words max
- Must summarize ONLY what the diff shows

BODY RULES:
- Start with a short summary sentence
- Include a short list of exactly what changed
- Keep it concise and factual
- Do not add sections like Risks unless meaningful from the diff itself
- No invented content
`

	userPrompt := fmt.Sprintf(`
Task Intent:
%s

Branch: %s
Base Branch: %s

Git Diff (vs %s):
%s
`, intent, branch, baseBranch, baseBranch, diff)

	// Call your existing Ollama chat client
	raw, err := Chat(prModel, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI PR metadata generation failed: %w", err)
	}

	raw = strings.TrimSpace(raw)

	// Some models like to wrap JSON in ```json ... ``` fences; strip them if present.
	if strings.HasPrefix(raw, "```") {
		raw = stripCodeFences(raw)
	}

	var meta PRMetadata
	if err := json.Unmarshal([]byte(raw), &meta); err != nil {
		return nil, fmt.Errorf("could not parse PR metadata JSON: %w\nraw: %s", err, raw)
	}

	return &meta, nil
}

// stripCodeFences removes leading/trailing ``` or ```json fences from a string.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		// remove first line (``` or ```json)
		lines := strings.Split(s, "\n")
		if len(lines) >= 2 {
			// drop first line, then look for closing fence at end
			lines = lines[1:]
			if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "```") {
				lines = lines[:len(lines)-1]
			}
			s = strings.Join(lines, "\n")
		}
	}
	return strings.TrimSpace(s)
}
