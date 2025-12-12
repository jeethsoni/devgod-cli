package ai

import (
	"fmt"
	"strings"
)

const DefaultModel = "llama3.1"

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
	const commitMessagePrompt = `
You are generating a Git commit message.

PRIMARY SOURCE OF TRUTH:
- The staged changes (diff or summary)

Your job is to produce a SINGLE, SHORT commit message.

HARD RULES (NO EXCEPTIONS):
- Output MUST be EXACTLY ONE LINE.
- FORMAT MUST be: "<type>: <short description>"
- <type> MUST be one of: feat, fix, chore, refactor, docs, style, test
- <short description> MUST be 3–10 words ONLY.
- Total length MUST be <= 60 characters.
- DO NOT include a body, only the one subject line.
- DO NOT describe the code in detail.
- DO NOT explain what the commit does in paragraphs.
- DO NOT start with "This commit" or anything similar.
- DO NOT mention files, functions, or modules by name unless necessary.
- DO NOT add lists, bullets, or multiple sentences.
- DO NOT add quotes, markdown, or extra formatting.

CHANGE COMPLETENESS RULE:
- If the staged diff includes file deletions (D) or renames (R),
  the commit message MUST reflect this.
- The description may use generic wording like:
  "remove unused file", "delete obsolete code", or "clean up old files".
- Do NOT ignore deletions or renames in favor of intent alone.

SAFETY RULES:
- If you detect obvious secrets (API keys, passwords, tokens, private keys,
  .pem contents, .env values, personal data):
  - Output exactly:
    WARNING: possible secret or sensitive data in diff; remove it before committing.
- If you detect obviously large/binary artifacts that should not be in git
  (e.g. big media, archives, compiled binaries):
  - Output exactly:
    WARNING: large or binary files detected; consider Git LFS instead of committing.

NORMAL CASE (NO SECRETS, NO LARGE FILES):
- Choose the <type> based on the intent+diff:
  - feat: new functionality
  - fix: bug fix
  - chore: tooling / config / plumbing
  - refactor: structural code change without new behavior
  - docs: documentation only
  - style: formatting / cosmetic only
  - test: tests only
- The description should summarize the changes at a high level only.
- Use imperative mood for the description:
  e.g. "add pr flow", NOT "added pr flow" or "adds pr flow".
- Do NOT restate the entire diff.
- Do NOT write an essay.

ABSOLUTE OUTPUT RULE:
- ENTIRE output MUST be ONE SINGLE LINE:
  - Either:
      "<type>: <short description>"
    or:
      a single WARNING line starting with "WARNING:" as described above.
- NO extra text before or after.
`

	userPrompt := fmt.Sprintf(
		`STAGED CHANGES:%s
	INTENT:%s`,
		strings.TrimSpace(diff),
		strings.TrimSpace(intent),
	)
	raw, err := Chat(DefaultModel, commitMessagePrompt, userPrompt)
	if err != nil {
		return "", err
	}

	msg := strings.TrimSpace(raw)
	msg = strings.Split(msg, "\n")[0]

	return msg, nil
}
