package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/shyn/kiro2cc/internal/config"
	"github.com/shyn/kiro2cc/pkg/types"
)

type Service interface {
	GetToken() (*types.TokenData, error)
	RefreshToken() error
	GetTokenFilePath() string
}

type service struct {
	config     *config.Config
	httpClient HTTPClient
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	Post(url, contentType string, body io.Reader) (*http.Response, error)
}

func NewService(cfg *config.Config) Service {
	return &service{
		config:     cfg,
		httpClient: &http.Client{},
	}
}

func NewServiceWithClient(cfg *config.Config, httpClient HTTPClient) Service {
	return &service{
		config:     cfg,
		httpClient: httpClient,
	}
}

func (s *service) GetTokenFilePath() string {
	// New default path
	newPath := s.config.Auth.TokenFilePath
	if _, err := os.Stat(newPath); err == nil {
		return newPath
	}

	// Fallback to old path for backward compatibility
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return newPath // Return new path even if it doesn't exist
	}
	oldPath := filepath.Join(homeDir, ".aws", "sso", "cache", "kiro-auth-token.json")
	if _, err := os.Stat(oldPath); err == nil {
		return oldPath
	}

	// If neither exists, return the new path as the default to be created
	return newPath
}

func (s *service) GetToken() (*types.TokenData, error) {
	tokenPath := s.GetTokenFilePath()
	if tokenPath == "" {
		return nil, fmt.Errorf("unable to determine token file path")
	}

	data, err := os.ReadFile(tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("token file not found at %s or legacy path. Please log in with Kiro first", tokenPath)
		}
		return nil, fmt.Errorf("failed to read token file at %s: %w", tokenPath, err)
	}

	var token types.TokenData
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token file: %w", err)
	}

	return &token, nil
}

func (s *service) RefreshToken() error {
	currentToken, err := s.GetToken()
	if err != nil {
		return fmt.Errorf("failed to get current token: %w", err)
	}

	refreshReq := types.RefreshRequest{
		RefreshToken: currentToken.RefreshToken,
	}

	reqBody, err := json.Marshal(refreshReq)
	if err != nil {
		return fmt.Errorf("failed to serialize refresh request: %w", err)
	}

	resp, err := s.httpClient.Post(
		s.config.Auth.RefreshTokenURL,
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return fmt.Errorf("failed to send refresh request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("refresh token failed with status %d: %s", resp.StatusCode, string(body))
	}

	var refreshResp types.RefreshResponse
	if err := json.NewDecoder(resp.Body).Decode(&refreshResp); err != nil {
		return fmt.Errorf("failed to parse refresh response: %w", err)
	}

	newToken := types.TokenData(refreshResp)
	return s.saveToken(&newToken)
}

func (s *service) saveToken(token *types.TokenData) error {
	tokenPath := s.config.Auth.TokenFilePath // Always save to the new path
	if tokenPath == "" {
		return fmt.Errorf("unable to determine token file path")
	}

	newData, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize new token: %w", err)
	}

	if err := os.WriteFile(tokenPath, newData, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}
