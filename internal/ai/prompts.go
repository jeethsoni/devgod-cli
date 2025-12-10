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
- <type> MUST be one of: feat, fix, chore, refactor, docs, style, test
- <slug> MUST be 2–6 meaningful words about the task.
- Use lowercase only.
- Use hyphens (-) between words in the slug.
- NEVER output just "fix", "feat", "chore", "refactor", "docs", "style", or "test".
- NEVER include explanations, quotes, or extra text.
- Keep the full branch reasonably short (~40 characters if possible).

GOOD EXAMPLES:
intent: add user onboarding flow for new accounts
branch: feat/onboarding-flow

intent: fix crash when password empty during login
branch: fix/empty-password-login-crash

intent: improve query performance in product list page
branch: refactor/product-query-optimization

intent: write setup documentation for new repo
branch: docs/setup-guide

OUTPUT FORMAT:
Return ONLY the branch name, like:
fix/empty-password-login-crash`

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
