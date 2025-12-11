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
