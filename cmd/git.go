package cmd

import (
	"strings"

	"github.com/jeethsoni/devgod-cli/internal/gitflow"
	"github.com/spf13/cobra"
)

// gitCmd represents the git command
var gitCmd = &cobra.Command{
	Use:   "git [intent]",
	Short: "AI-powered git workflow",
	Long:  "Generate branches and commits from simple English instructions.",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Join all args to form intent
		intent := strings.Join(args, " ")

		if strings.TrimSpace(intent) == "" {
			// No intent then start finish mode
			return gitflow.FinishTask()
		}

		// Intent given then start mode
		return gitflow.StartTask(intent)
	},
}

func init() {
	rootCmd.AddCommand(gitCmd)
}
