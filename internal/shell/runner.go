package shell

import (
	"fmt"
	"os/exec"
	"strings"
)

// Run executes a shell command and returns its combined output or an error.
func Run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput() // Capture both stdout and stderr
	output := string(out)

	// If there was an error, include the command output in the error message.
	if err != nil {
		return output, fmt.Errorf("command failed: %s %s\n%w\noutput:\n%s", name, strings.Join(args, " "), err, output)
	}

	return output, nil
}
