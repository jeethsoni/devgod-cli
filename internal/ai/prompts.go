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
// Priority: summary -> diff -> intent.
func GenerateCommitMessage(intent, summary, diff string) (string, error) {
	const commitMessagePrompt = `
You are generating a Git commit message.

PRIMARY SOURCE OF TRUTH (IN ORDER):
1) STAGED SUMMARY (what changed: A/M/D/R)
2) STAGED DIFF (details)
3) TASK INTENT (wording help only)

If there is any conflict, the staged summary/diff ALWAYS win.

Your job is to output ONE SINGLE LINE in this format:
"<type>: <short description>"

HARD RULES (NO EXCEPTIONS):
- Output MUST be EXACTLY ONE LINE.
- FORMAT MUST be: "<type>: <short description>"
- <type> MUST be one of: feat, fix, chore, refactor, docs, style, test
- <short description> MUST be 3–10 words ONLY.
- Total length MUST be <= 60 characters.
- No body. No extra lines. No markdown. No quotes. No emojis.
- Do NOT mention files/functions/modules by name unless absolutely necessary.
- Do NOT output contradictory subjects like "feat: fix ...".
  If the description contains "fix", the type MUST be "fix".

TYPE SELECTION (STRICT):
- Use "fix" when the change prevents or corrects incorrect behavior in an existing flow
  (e.g., PR/commit flow failing, errors, broken behavior, missing required steps).
- Use "feat" ONLY when it introduces a new user-facing capability (not just preventing an error).
- Use "chore" for tooling/config/maintenance without behavior change.
- Use "refactor" only for structural code changes without behavior change.
- Use "docs/style/test" only when the staged changes are exclusively those categories.

CHANGE COMPLETENESS RULE:
- If staged changes include deletions (D) or renames (R), the message MUST reflect that
  using generic wording ("remove unused code", "clean up old files", "rename ..." without filenames).

SAFETY RULES:
- If you detect obvious secrets (API keys, passwords, tokens, private keys, .pem contents, .env values, personal data):
  Output exactly:
  WARNING: possible secret or sensitive data in diff; remove it before committing.
- If you detect obviously large/binary artifacts that should not be in git (big media, archives, compiled binaries):
  Output exactly:
  WARNING: large or binary files detected; consider Git LFS instead of committing.

QUALITY RULES:
- Avoid vague descriptions like "fix pr flow".
- Prefer concrete outcomes like:
  "push branch before creating pr"
  "handle unpushed branches before pr creation"
  "prevent pr creation failure when branch is local"
- Use imperative mood: "add", "update", "handle", "prevent", "push".

ABSOLUTE OUTPUT RULE:
- Output MUST be exactly ONE LINE:
  - Either "<type>: <short description>"
  - Or a single WARNING line starting with "WARNING:"
`

	userPrompt := fmt.Sprintf(
		`STAGED SUMMARY (PRIMARY):
%s

STAGED DIFF (DETAILS):
%s

TASK INTENT (SECONDARY):
%s
`,
		strings.TrimSpace(summary),
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
