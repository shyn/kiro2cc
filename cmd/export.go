package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/shyn/kiro2cc/internal/auth"
	"github.com/shyn/kiro2cc/internal/config"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export environment variables",
	Long:  "Export environment variables for other tools to use the Anthropic API proxy.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Default()
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}
		authService := auth.NewService(cfg)
		
		token, err := authService.GetToken()
		if err != nil {
			return fmt.Errorf("failed to read token, please install Kiro and login first: %w", err)
		}

		baseURL := fmt.Sprintf("http://localhost:%s", cfg.Server.Port)

		if runtime.GOOS == "windows" {
			fmt.Println("CMD")
			fmt.Printf("set ANTHROPIC_BASE_URL=%s\n", baseURL)
			fmt.Printf("set ANTHROPIC_API_KEY=%s\n\n", token.AccessToken)
			fmt.Println("Powershell")
			fmt.Printf(`$env:ANTHROPIC_BASE_URL="%s"`, baseURL)
			fmt.Printf("\n")
			fmt.Printf(`$env:ANTHROPIC_API_KEY="%s"`, token.AccessToken)
			fmt.Printf("\n")
		} else {
			fmt.Printf("export ANTHROPIC_BASE_URL=%s\n", baseURL)
			fmt.Printf("export ANTHROPIC_API_KEY=\"%s\"\n", token.AccessToken)
		}
		
		return nil
	},
}