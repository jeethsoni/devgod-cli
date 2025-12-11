package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

const prModel = "llama3.1"

type PRMetadata struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Reviewers []string `json:"reviewers"`
}

func GeneratePRMetadata(intent, diff, branch, baseBranch string) (*PRMetadata, error) {
	systemPrompt := `
You are a senior software engineer writing GitHub Pull Request titles and descriptions.

You will receive:
- A high-level task intent (for wording only)
- A git diff or diff summary (THIS IS THE ONLY SOURCE OF TRUTH)

Your job is to output ONE JSON OBJECT:

{
  "title": "<short-title>",
  "body": "<markdown-body>"
}

========================
STRICT RULES (NO EXCEPTIONS)
========================

JSON RULES:
- Output MUST be valid JSON.
- No text before or after the JSON.
- No code fences.
- Only "title" and "body" keys are allowed.

TITLE RULES:
- 3–9 words.
- One line only.
- No quotes, no backticks, no emojis, no brackets.
- Must summarize the purpose of the PR.
- Must NOT reference branch names or issue IDs unless visible in the diff.

BODY RULES:
- "body" must be a markdown string with 2–6 natural sentences.
- NO sections like Summary, Changes, Testing, etc.
- NO leading phrases like:
  "This pull request"
  "In this pull request"
  "This PR"
  "In this PR"
  "The purpose of this PR"
  "This change does..."
- Instead begin directly with the **action or outcome**, e.g.:
  "Adds a helper function for number addition."
  "Introduces a new CLI command for user login."

STYLE RULES:
- Tone must be professional and concise.
- Describe WHAT changed and WHY it matters at a high level.
- No line-by-line explanation.
- No code fences.
- No bullet lists unless multiple independent changes require clarity.
- Must NOT describe behavior not shown in the diff.

DIFF-ONLY TRUTH RULE:
- Only describe changes visible in the diff.
- Do NOT invent tests, error handling, validation, performance improvements, or any behavior not present.

EMPTY DIFF RULE:
{
  "title": "No code changes",
  "body": "No code modifications are included."
}`

	userPrompt := fmt.Sprintf(`
Task intent:
%s

Branch: %s
Base branch: %s

RAW GIT DIFF:
%s
`, strings.TrimSpace(intent), branch, baseBranch, strings.TrimSpace(diff))

	raw, err := Chat(prModel, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI PR metadata generation failed: %w", err)
	}

	raw = strings.TrimSpace(raw)

	// Remove accidental code fences
	if strings.HasPrefix(raw, "```") {
		raw = stripCodeFences(raw)
	}

	// Parse JSON strictly — NO fallback
	meta := &PRMetadata{}
	if err := json.Unmarshal([]byte(raw), meta); err != nil {
		return nil, fmt.Errorf("failed to parse AI JSON: %w\nraw output:\n%s", err, raw)
	}

	// Minimal validation
	if meta.Title == "" || meta.Body == "" {
		return nil, fmt.Errorf("AI returned incomplete PR metadata:\n%s", raw)
	}

	return meta, nil
}

func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}

	lines := strings.Split(s, "\n")
	if len(lines) <= 2 {
		return s
	}

	lines = lines[1:]
	last := strings.TrimSpace(lines[len(lines)-1])
	if strings.HasPrefix(last, "```") {
		lines = lines[:len(lines)-1]
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}
