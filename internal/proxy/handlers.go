package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/shyn/kiro2cc/internal/auth"
	"github.com/shyn/kiro2cc/internal/client"
	"github.com/shyn/kiro2cc/internal/translator"
	"github.com/shyn/kiro2cc/parser"
	"github.com/shyn/kiro2cc/pkg/types"
)

type Handlers struct {
	authService auth.Service
	translator  translator.Service
	cwClient    client.CodeWhispererClient
	logger      Logger
}

type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

func NewHandlers(
	authService auth.Service,
	translator translator.Service,
	cwClient client.CodeWhispererClient,
	logger Logger,
) *Handlers {
	return &Handlers{
		authService: authService,
		translator:  translator,
		cwClient:    cwClient,
		logger:      logger,
	}
}

func (h *Handlers) MessagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.logger.Error("Unsupported method: %s", r.Method)
		http.Error(w, "Only POST requests are supported", http.StatusMethodNotAllowed)
		return
	}

	token, err := h.authService.GetToken()
	if err != nil {
		h.logger.Error("Failed to get token: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get token: %v", err), http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("Failed to read request body: %v", err)
		http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	h.logger.Debug("Anthropic request body:\n%s", string(body))

	var anthropicReq types.AnthropicRequest
	if err := json.Unmarshal(body, &anthropicReq); err != nil {
		h.logger.Error("Failed to parse request body: %v", err)
		http.Error(w, fmt.Sprintf("Failed to parse request body: %v", err), http.StatusBadRequest)
		return
	}

	if anthropicReq.Stream {
		h.handleStreamRequest(w, &anthropicReq, token.AccessToken)
		return
	}

	h.handleNonStreamRequest(w, &anthropicReq, token.AccessToken)
}

func (h *Handlers) handleStreamRequest(w http.ResponseWriter, anthropicReq *types.AnthropicRequest, accessToken string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	messageId := fmt.Sprintf("msg_%s", time.Now().Format("20060102150405"))

	cwReq, err := h.translator.ToCodeWhisperer(anthropicReq)
	if err != nil {
		h.sendErrorEvent(w, flusher, "Translation failed", err)
		return
	}

	resp, err := h.cwClient.SendRequest(cwReq, accessToken, true)
	if err != nil {
		h.sendErrorEvent(w, flusher, "CodeWhisperer request error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		h.logger.Error("CodeWhisperer response error, status: %d, response: %s", resp.StatusCode, string(body))

		if resp.StatusCode == 403 {
			if err := h.authService.RefreshToken(); err != nil {
				h.logger.Error("Failed to refresh token: %v", err)
			}
			h.sendErrorEvent(w, flusher, "error", fmt.Errorf("CodeWhisperer Token refreshed, please retry"))
		} else {
			h.sendErrorEvent(w, flusher, "error", fmt.Errorf("CodeWhisperer Error: %s", string(body)))
		}
		return
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.sendErrorEvent(w, flusher, "error", fmt.Errorf("Failed to read response"))
		return
	}

	events := parser.ParseEvents(respBody)

	if len(events) > 0 {
		h.sendStreamingEvents(w, flusher, events, messageId, anthropicReq)
	}
}

func (h *Handlers) sendStreamingEvents(w http.ResponseWriter, flusher http.Flusher, events []parser.SSEEvent, messageId string, anthropicReq *types.AnthropicRequest) {
	messageStart := map[string]any{
		"type": "message_start",
		"message": map[string]any{
			"id":            messageId,
			"type":          "message",
			"role":          "assistant",
			"content":       []any{},
			"model":         anthropicReq.Model,
			"stop_reason":   nil,
			"stop_sequence": nil,
			"usage": map[string]any{
				"input_tokens":  1,
				"output_tokens": 1,
			},
		},
	}
	h.sendSSEEvent(w, flusher, "message_start", messageStart)
	h.sendSSEEvent(w, flusher, "ping", map[string]string{"type": "ping"})

	contentBlockStart := map[string]any{
		"content_block": map[string]any{
			"text": "",
			"type": "text",
		},
		"index": 0,
		"type":  "content_block_start",
	}
	h.sendSSEEvent(w, flusher, "content_block_start", contentBlockStart)

	outputTokens := 0
	for _, e := range events {
		h.sendSSEEvent(w, flusher, e.Event, e.Data)
		if e.Event == "content_block_delta" {
			outputTokens++
		}
		time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
	}

	contentBlockStop := map[string]any{
		"index": 0,
		"type":  "content_block_stop",
	}
	h.sendSSEEvent(w, flusher, "content_block_stop", contentBlockStop)

	messageDelta := map[string]any{
		"type": "message_delta",
		"delta": map[string]any{
			"stop_reason":   "end_turn",
			"stop_sequence": nil,
		},
		"usage": map[string]any{
			"output_tokens": outputTokens,
		},
	}
	h.sendSSEEvent(w, flusher, "message_delta", messageDelta)

	messageStop := map[string]any{
		"type": "message_stop",
	}
	h.sendSSEEvent(w, flusher, "message_stop", messageStop)
}

func (h *Handlers) handleNonStreamRequest(w http.ResponseWriter, anthropicReq *types.AnthropicRequest, accessToken string) {
	cwReq, err := h.translator.ToCodeWhisperer(anthropicReq)
	if err != nil {
		h.logger.Error("Translation failed: %v", err)
		http.Error(w, fmt.Sprintf("Translation failed: %v", err), http.StatusInternalServerError)
		return
	}

	resp, err := h.cwClient.SendRequest(cwReq, accessToken, false)
	if err != nil {
		h.logger.Error("Failed to send request: %v", err)
		http.Error(w, fmt.Sprintf("Failed to send request: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	cwRespBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("Failed to read response: %v", err)
		http.Error(w, fmt.Sprintf("Failed to read response: %v", err), http.StatusInternalServerError)
		return
	}

	anthropicResp, err := h.translator.FromCodeWhisperer(cwRespBody, anthropicReq.Model)
	if err != nil {
		h.logger.Error("Translation from CodeWhisperer failed: %v", err)
		http.Error(w, fmt.Sprintf("Translation failed: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(anthropicResp)
}

func (h *Handlers) sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, eventType string, data any) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	fmt.Fprintf(w, "event: %s\n", eventType)
	fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
	flusher.Flush()
}

func (h *Handlers) sendErrorEvent(w http.ResponseWriter, flusher http.Flusher, message string, err error) {
	errorResp := map[string]any{
		"type": "error",
		"error": map[string]any{
			"type":    "overloaded_error",
			"message": message,
		},
	}
	h.sendSSEEvent(w, flusher, "error", errorResp)
}

func (h *Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handlers) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Access to unknown endpoint: %s", r.URL.Path)
	http.Error(w, "404 Not Found", http.StatusNotFound)
}

