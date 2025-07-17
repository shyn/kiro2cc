package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kiro2cc",
	Short: "Kiro to Claude Code bridge",
	Long: `A CLI tool that manages Kiro authentication tokens and provides 
an Anthropic API proxy service. The tool acts as a bridge between 
Anthropic API requests and AWS CodeWhisperer.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(readCmd)
	rootCmd.AddCommand(refreshCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(claudeCmd)
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(stopCmd)
}