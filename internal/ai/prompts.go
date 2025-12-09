package ai

import (
	"fmt"
	"strings"
)

const DefaultModel = "qwen2.5-coder:7b"

// GenerateBranchName uses AI to create a clean git branch name.
func GenerateBranchName(intent string) (string, error) {
	systemPrompt := `You are a senior engineer generating git branch names.

RULES:
- Your ONLY output must be a valid git branch name.
- Use one of these prefixes:
  feat/, fix/, chore/, refactor/, docs/, style/, test/
- Summarize long intents into 3–6 meaningful words.
- Use lowercase.
- Use hyphens between words.
- NEVER output just the prefix (like "fix" or "feat").
- NEVER include quotes, spaces, or explanations.
- Aim to keep the full branch under ~40 characters.
- Make branch names short but meaningful.

EXAMPLES:
intent: add user onboarding flow
branch: feat/onboarding-flow

intent: fix crash when password empty during login
branch: fix/empty-password-login-crash

intent: improve query performance in product list page
branch: refactor/product-query-optimization

intent: write setup documentation for new repo
branch: docs/setup-guide

OUTPUT FORMAT:
Return ONLY the branch name. Nothing else.`

	userPrompt := fmt.Sprintf("intent: %s", strings.TrimSpace(intent))

	raw, err := Chat(DefaultModel, systemPrompt, userPrompt)
	if err != nil {
		return "", err
	}

	branch := strings.TrimSpace(raw)
	branch = strings.Split(branch, "\n")[0]       // first line only
	branch = strings.ReplaceAll(branch, " ", "-") // safety: no spaces

	// 🔍 Basic validation to avoid garbage like "fix" or "feat"
	if branch == "" ||
		len(branch) < 5 ||
		!strings.Contains(branch, "/") ||
		strings.HasSuffix(branch, "/") ||
		branch == "fix" || branch == "feat" || branch == "chore" || branch == "refactor" {
		return "", fmt.Errorf("model returned invalid branch name: %q", branch)
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
