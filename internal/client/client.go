package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/shyn/kiro2cc/internal/config"
	"github.com/shyn/kiro2cc/pkg/types"
)

type CodeWhispererClient interface {
	SendRequest(req *types.CodeWhispererRequest, accessToken string, stream bool) (*http.Response, error)
}

type client struct {
	config     *config.Config
	httpClient HTTPClient
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewCodeWhispererClient(cfg *config.Config) CodeWhispererClient {
	return &client{
		config:     cfg,
		httpClient: &http.Client{},
	}
}

func NewCodeWhispererClientWithHTTPClient(cfg *config.Config, httpClient HTTPClient) CodeWhispererClient {
	return &client{
		config:     cfg,
		httpClient: httpClient,
	}
}

func (c *client) SendRequest(req *types.CodeWhispererRequest, accessToken string, stream bool) (*http.Response, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %w", err)
	}

	url := c.config.CodeWhisperer.BaseURL + "/generateAssistantResponse"
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	if stream {
		httpReq.Header.Set("Accept", "text/event-stream")
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

