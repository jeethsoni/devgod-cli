package gitflow

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jeethsoni/devgod-cli/internal/ui"
)

// isGHAuthenticated checks if the user is logged into GitHub CLI.
// It returns (bool, string) where the string is the raw output from gh.
func isGHAuthenticated() (bool, string) {
	cmd := exec.Command("gh", "auth", "status", "--hostname", "github.com")
	out, err := cmd.CombinedOutput()
	output := string(out)

	if err != nil {
		// Not logged in or some error; treat as unauthenticated.
		return false, output
	}

	// Typical success output: "Logged in to github.com as <username>"
	if strings.Contains(output, "Logged in to github.com") {
		return true, output
	}

	return false, output
}

// ensureGHAuthenticated ensures the user is logged into GitHub via gh.
// If not, it offers to run `gh auth login` interactively.
// Returns an error if the user declines or if login fails.
func ensureGHAuthenticated() error {
	ok, raw := isGHAuthenticated()
	if ok {
		// Optional: debug line
		// fmt.Println(ui.Green("✔️ GitHub CLI is already authenticated."))
		return nil
	}

	// User is NOT logged in
	fmt.Println(ui.Red("❌ You are not logged into GitHub CLI."))
	if strings.TrimSpace(raw) != "" {
		// Show what gh said (helpful for debugging)
		fmt.Println(strings.TrimSpace(raw))
	}
	fmt.Println()

	if !ui.Confirm("Run `gh auth login` now?") {
		fmt.Println(ui.Yellow("To use devgod PR features, please run:"))
		fmt.Println()
		fmt.Println("   gh auth login")
		fmt.Println()
		fmt.Println("Then re-run `devgod pr`.")
		return fmt.Errorf("user is not authenticated with GitHub CLI")
	}

	// Run `gh auth login` interactively
	if err := runGHAuthLoginInteractive(); err != nil {
		fmt.Println(ui.Red("❌ `gh auth login` failed:"))
		return err
	}

	// Re-check auth status after login attempt
	ok, _ = isGHAuthenticated()
	if !ok {
		return fmt.Errorf("GitHub CLI authentication did not complete successfully")
	}

	fmt.Println(ui.Green("✔️ GitHub CLI authentication complete."))
	return nil
}

// runGHAuthLoginInteractive runs `gh auth login` attached to the user's TTY
// so they can interact with the GitHub CLI prompts normally.
func runGHAuthLoginInteractive() error {
	cmd := exec.Command("gh", "auth", "login")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println(ui.Yellow("Launching `gh auth login`..."))
	fmt.Println("Follow the prompts to authenticate with GitHub.")
	fmt.Println()

	return cmd.Run()
}
