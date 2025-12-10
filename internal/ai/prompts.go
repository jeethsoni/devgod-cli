package ai

import (
	"fmt"
	"strings"
)

const DefaultModel = "qwen2.5-coder:7b"

// GenerateBranchName uses AI to create a clean git branch name.
func GenerateBranchName(intent string) (string, error) {
	systemPrompt := `You are a senior engineer generating git branch names.

You MUST follow these rules:

- Your ONLY output must be a valid git branch name.
- The format MUST be: <type>/<slug>
- <type> MUST be one of: feat, fix, chore, refactor, docs, style, test.
- <slug> MUST be 2–6 meaningful words about the task, in lowercase kebab-case.
- Use hyphens (-) between all words in the slug.
- Keep the branch reasonably short (~40 characters if possible).
- NEVER output only the type (feat, fix, chore, refactor, docs, style, test).
- NEVER include explanations, quotes, or any other text besides the branch name.

ISSUE ID RULES:

- You will be given an "Issue ID" field in the input. It might be empty or "none".
- ONLY include an issue ID in the slug if the Issue ID field is a non-empty value that is not "none", "null", or "n/a".
- When an Issue ID is present, append it at the END of the slug, separated by a hyphen.
  Example: signup-login-flow-JIRA-452
- NEVER invent or guess an issue ID.
- NEVER derive an issue ID from the task description or any other text.
- If no Issue ID is provided (or the Issue ID is empty / "none" / "null" / "n/a"), you MUST NOT include any ticket-like token (e.g., ABC-123, JIRA-1, BUG-42) in the branch name.

GOOD EXAMPLES:

intent: add user onboarding flow for new accounts
issue id: none
-> feat/onboarding-flow

intent: fix crash when password empty during login
issue id: BUG-21
-> fix/empty-password-login-crash-BUG-21

intent: improve query performance in product list page
issue id: PERF-88
-> refactor/product-query-optimization-PERF-88

intent: write setup documentation for new repo
issue id: none
-> docs/setup-guide

OUTPUT FORMAT:
Return ONLY the branch name, like:
fix/empty-password-login-crash
or:
fix/empty-password-login-crash-BUG-21`

	userPrompt := fmt.Sprintf("intent: %s", strings.TrimSpace(intent))

	if intent == "" {
		return "", fmt.Errorf("intent cannot be empty")
	}

	raw, err := Chat(DefaultModel, systemPrompt, userPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to call AI: %w", err)
	}

	branch := strings.TrimSpace(raw)

	// If the model ever returns things like "branch: fix/...", strip that prefix.
	branch = strings.TrimPrefix(branch, "branch:")
	branch = strings.TrimSpace(branch)

	// Take only the first line, in case it babbles.
	if idx := strings.Index(branch, "\n"); idx != -1 {
		branch = branch[:idx]
	}

	branch = strings.TrimSpace(branch)
	branch = strings.ReplaceAll(branch, " ", "-")

	// Minimal validation: still no fallback, just error if garbage
	if branch == "" ||
		len(branch) < 5 ||
		!strings.Contains(branch, "/") {
		return "", fmt.Errorf("model returned invalid branch name: %q", branch)
	}

	prefix := strings.SplitN(branch, "/", 2)[0]
	if prefix != "feat" && prefix != "fix" && prefix != "chore" &&
		prefix != "refactor" && prefix != "docs" && prefix != "style" && prefix != "test" {
		return "", fmt.Errorf("model returned invalid prefix in branch name: %q", branch)
	}

	return branch, nil
}

// GenerateCommitMessage uses AI to generate a single-line commit message.
func GenerateCommitMessage(intent, diff string) (string, error) {
	systemPrompt := `You are a senior engineer writing concise conventional commit messages.

					Given:
					- an intent (what the developer was trying to do)
					- a git diff (staged changes)

					RULES:
					- Respond with a SINGLE LINE commit message.
					- Format: type: short summary
					- Types: feat, fix, chore, refactor, docs, test, style
					- Use lowercase.
					- Aim for 50–72 characters if possible.
					- No trailing period.
					- Do NOT include anything except the commit line.
					- Do NOT explain your reasoning.

					GOOD EXAMPLES:
					fix: prevent crash when password is empty
					feat: add onboarding flow for new users
					chore: remove deprecated auth endpoints
					refactor: simplify product query building`

	userPrompt := fmt.Sprintf("intent:\n%s\n\nstaged diff:\n%s", strings.TrimSpace(intent), strings.TrimSpace(diff))

	raw, err := Chat(DefaultModel, systemPrompt, userPrompt)
	if err != nil {
		return "", err
	}

	msg := strings.TrimSpace(raw)
	msg = strings.Split(msg, "\n")[0]

	return msg, nil
}
