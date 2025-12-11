package cmd

import (
	"github.com/jeethsoni/devgod-cli/internal/gitflow"
	"github.com/spf13/cobra"
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Create a pull request for the current branch",
	Long:  "Creates a pull request on the remote repository for the current branch using AI-generated title and description.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return gitflow.CreatePR()
	},
}

func init() {
	rootCmd.AddCommand(prCmd)
}
