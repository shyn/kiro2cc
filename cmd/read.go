package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/shyn/kiro2cc/internal/auth"
	"github.com/shyn/kiro2cc/internal/config"
)

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Read and display token information",
	Long:  "Read the Kiro authentication token from the cache and display its information.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Default()
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}
		authService := auth.NewService(cfg)
		
		token, err := authService.GetToken()
		if err != nil {
			return fmt.Errorf("failed to read token: %w", err)
		}

		fmt.Println("Token information:")
		fmt.Printf("Access Token: %s\n", token.AccessToken)
		fmt.Printf("Refresh Token: %s\n", token.RefreshToken)
		if token.ExpiresAt != "" {
			fmt.Printf("Expires At: %s\n", token.ExpiresAt)
		}
		
		return nil
	},
}