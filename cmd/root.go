package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "devgod-cli",
	Short: "devgod-cli is your AI-powered assistant for git workflows",
	Long:  "devgod-cli helps you automate git workflows using AI, from branch creation to commit messages and PR creation.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("devgod-cli: try `devgod git 'your task'`")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
