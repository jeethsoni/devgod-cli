package gitflow

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/jeethsoni/devgod-cli/internal/ui"
)

func ghInstalled() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

func installGH() error {
	switch runtime.GOOS {
	case "darwin":
		return installGHMac()
	case "linux":
		return installGHLinux()
	case "windows":
		return installGHWindows()
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func installGHMac() error {
	fmt.Println(ui.Yellow("Installing GitHub CLI using Homebrew…"))
	cmd := exec.Command("brew", "install", "gh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func installGHLinux() error {
	fmt.Println(ui.Yellow("Installing GitHub CLI using apt…"))
	cmd := exec.Command("sudo", "apt", "install", "-y", "gh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func installGHWindows() error {
	fmt.Println(ui.Yellow("Installing GitHub CLI using winget…"))
	cmd := exec.Command("winget", "install", "-e", "--id", "GitHub.cli")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ensureGitHubCLIInstalled() error {
	if ghInstalled() {
		return nil
	}

	fmt.Println(ui.Yellow("⚠️ GitHub CLI (gh) is not installed."))

	if !ui.Confirm("Install GitHub CLI now?") {
		fmt.Println(ui.Yellow("To use devgod PR features, please install GitHub CLI from:"))
		fmt.Println()
		fmt.Println("   https://cli.github.com/")
		fmt.Println()
		fmt.Println("Then re-run `devgod pr`.")
		return fmt.Errorf("GitHub CLI (gh) is required to create PRs")
	}

	if err := installGH(); err != nil {
		fmt.Println(ui.Red("❌ Failed to install GitHub CLI automatically."))
		fmt.Println(ui.Yellow("Please install it manually from https://cli.github.com/ and try again."))
		return fmt.Errorf("failed to install GitHub CLI: %w", err)
	}

	fmt.Println(ui.Green("✅ GitHub CLI (gh) installed."))
	return nil
}
