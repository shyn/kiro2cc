package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/shyn/kiro2cc/internal/auth"
	"github.com/shyn/kiro2cc/internal/config"
)

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh the access token",
	Long:  "Refresh the Kiro access token using the stored refresh token.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Default()
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}
		authService := auth.NewService(cfg)
		
		if err := authService.RefreshToken(); err != nil {
			return fmt.Errorf("failed to refresh token: %w", err)
		}

		fmt.Println("Token refreshed successfully!")
		
		// Display the new token
		token, err := authService.GetToken()
		if err != nil {
			return fmt.Errorf("failed to read refreshed token: %w", err)
		}
		
		fmt.Printf("New Access Token: %s\n", token.AccessToken)
		
		return nil
	},
}