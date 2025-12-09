package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "devgod-cli",
	Short: "devgod-cli is your AI-powered assistant for git workflows",
	Long:  "devgod-cli helps you automate git workflows using AI, from branch creation to commit messages and PR creation.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("devgod-cli: try `devgod git 'your task'`")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
