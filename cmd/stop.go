package cmd

import (
	"github.com/spf13/cobra"
	"github.com/shyn/kiro2cc/internal/config"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the kiro2cc background server",
	Long:  "Finds and stops the kiro2cc server process that is running in the background.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Default()
		if err != nil {
			return err
		}
		return stopServer(cfg)
	},
}

func init() {
	// Command is added in root.go
}