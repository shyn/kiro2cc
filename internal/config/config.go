package config

import (
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Server        ServerConfig
	Auth          AuthConfig
	CodeWhisperer CodeWhispererConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PIDFilePath  string
}

type AuthConfig struct {
	TokenFilePath   string
	RefreshTokenURL string
}

type CodeWhispererConfig struct {
	BaseURL    string
	ProfileArn string
	ProxyURL   string
}

// GetConfigDir gets the configuration directory for kiro2cc.
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "kiro2cc"), nil
}

// Default creates a default configuration.
func Default() (*Config, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	// Ensure the config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	return &Config{
		Server: ServerConfig{
			Port:         "8080",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			PIDFilePath:  filepath.Join(configDir, "kiro2cc.pid"),
		},
		Auth: AuthConfig{
			RefreshTokenURL: "https://prod.us-east-1.auth.desktop.kiro.dev/refreshToken",
			TokenFilePath:   filepath.Join(configDir, "kiro2cc-token.json"),
		},
		CodeWhisperer: CodeWhispererConfig{
			BaseURL:    "https://codewhisperer.us-east-1.amazonaws.com",
			ProfileArn: "arn:aws:codewhisperer:us-east-1:699475941385:profile/EHGA3GRVQMUK",
			ProxyURL:   "127.0.0.1:9000",
		},
	}, nil
}
