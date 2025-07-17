package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/shyn/kiro2cc/internal/auth"
	"github.com/shyn/kiro2cc/internal/config"
)

var claudeCmd = &cobra.Command{
	Use:   "claude [args...]",
	Short: "Run claude-code with automatic server and environment setup",
	Long: `This command streamlines the use of claude-code by:
1. Checking if the kiro2cc server is running.
2. Starting the server in the background if it's not running.
3. Refreshing the authentication token.
4. Setting the necessary environment variables (ANTHROPIC_BASE_URL, ANTHROPIC_API_KEY).
5. Executing 'claude' with any provided arguments.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Default()
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}

		// 1. Check and start server
		if !isRunning(cfg.Server.PIDFilePath) {
			fmt.Println("Server not running, starting it in the background...")
			if err := startDaemon(cfg); err != nil {
				return fmt.Errorf("failed to start server daemon: %w", err)
			}
		} else {
			fmt.Println("Server is already running.")
		}

		// 2. Refresh token
		fmt.Println("Refreshing token...")
		authService := auth.NewService(cfg)
		if err := authService.RefreshToken(); err != nil {
			return fmt.Errorf("failed to refresh token: %w", err)
		}
		token, err := authService.GetToken()
		if err != nil {
			return fmt.Errorf("failed to get token after refresh: %w", err)
		}
		fmt.Println("Token refreshed successfully.")

		// 3. Set environment variables
		baseURL := fmt.Sprintf("http://localhost:%s", cfg.Server.Port)
		os.Setenv("ANTHROPIC_BASE_URL", baseURL)
		os.Setenv("ANTHROPIC_API_KEY", token.AccessToken)

		// 4. Find and execute claude
		claudePath, err := exec.LookPath("claude")
		if err != nil {
			return fmt.Errorf("'claude' executable not found in your PATH. Please ensure it is installed and accessible")
		}

		fmt.Printf("Executing %s with proxy URL %s\n", claudePath, baseURL)
		fmt.Println("----------------------------------------")

		// Replace the current process with claude
		err = syscall.Exec(claudePath, append([]string{"claude"}, args...), os.Environ())
		if err != nil {
			return fmt.Errorf("failed to execute 'claude': %w", err)
		}

		return nil
	},
}

func init() {
	// Command is added in root.go
}
