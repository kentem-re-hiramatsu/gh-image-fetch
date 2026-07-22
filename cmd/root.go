package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "image-fetch",
	Short: "Download GitHub user-attachments assets using gh CLI credentials",
	Long: `gh-image-fetch is a gh extension that downloads images and files
attached to GitHub issues and pull requests (user-attachments URLs)
using the authentication already configured for the gh CLI.`,
	SilenceUsage:  true,
	SilenceErrors: false,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
