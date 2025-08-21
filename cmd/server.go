package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/shyn/kiro2cc/internal/auth"
	"github.com/shyn/kiro2cc/internal/client"
	"github.com/shyn/kiro2cc/internal/config"
	"github.com/shyn/kiro2cc/internal/proxy"
	"github.com/shyn/kiro2cc/internal/translator"
)

var (
	port   string
	daemon bool
	stop   bool
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage the Anthropic API proxy server",
	Long:  "Start, stop, or manage the HTTP proxy server that translates Anthropic API requests.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Default()
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}

		if port != "" {
			cfg.Server.Port = port
		}

		if stop {
			return stopServer(cfg)
		}

		if daemon {
			return startDaemon(cfg)
		}

		return startServer(cfg)
	},
}

func init() {
	serverCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to run the server on")
	serverCmd.Flags().BoolVarP(&daemon, "daemon", "d", false, "Run the server in the background")
	serverCmd.Flags().BoolVar(&stop, "stop", false, "Stop the running server")
}

func startServer(cfg *config.Config) error {
	logger := proxy.NewSimpleLogger()
	authService := auth.NewService(cfg)
	translatorService := translator.NewService(cfg)
	cwClient := client.NewCodeWhispererClient(cfg)

	handlers := proxy.NewHandlers(authService, translatorService, cwClient, logger)
	server := proxy.NewServer(cfg, handlers, logger)

	fmt.Printf("Starting server on port %s...\n", cfg.Server.Port)
	return server.Start()
}

func startDaemon(cfg *config.Config) error {
	if isRunning(cfg.Server.PIDFilePath) {
		fmt.Println("Server is already running.")
		return nil
	}

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find executable path: %w", err)
	}

	cmd := exec.Command(executable, "server", "--port", cfg.Server.Port)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(cfg.Server.PIDFilePath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	pid := cmd.Process.Pid
	if err := ioutil.WriteFile(cfg.Server.PIDFilePath, []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	fmt.Printf("Server started in background with PID: %d on port %s\n", pid, cfg.Server.Port)
	time.Sleep(1 * time.Second)
	return nil
}

func stopServer(cfg *config.Config) error {
	pid, err := readPIDFile(cfg.Server.PIDFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Server is not running (PID file not found).")
			return nil
		}
		return fmt.Errorf("failed to read PID file: %w", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Println("Could not find process, cleaning up stale PID file.")
		os.Remove(cfg.Server.PIDFilePath)
		return nil
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		if err == syscall.ESRCH {
			fmt.Println("Process not found, cleaning up stale PID file.")
			os.Remove(cfg.Server.PIDFilePath)
			return nil
		}
		return fmt.Errorf("failed to send SIGTERM to process %d: %w", pid, err)
	}

	os.Remove(cfg.Server.PIDFilePath)
	fmt.Printf("Server with PID %d has been stopped.\n", pid)
	return nil
}

func isRunning(pidFile string) bool {
	pid, err := readPIDFile(pidFile)
	if err != nil {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func readPIDFile(pidFile string) (int, error) {
	data, err := ioutil.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("invalid PID in PID file: %w", err)
	}
	return pid, nil
}
