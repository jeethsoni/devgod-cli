package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "devgod-cli",
	Short: "devgod-cli is your AI-powered assistant for git workflows",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("devgod-cli: try `devgod-cli git 'your task'`")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
