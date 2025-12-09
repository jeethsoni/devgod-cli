package shell

import (
	"fmt"
	"os/exec"
	"strings"
)

func Run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	output := string(out)

	if err != nil {
		return output, fmt.Errorf("command failed: %s %s\n%w\noutput:\n%s", name, strings.Join(args, " "), err, output)
	}

	return output, nil
}
