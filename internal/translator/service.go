package translator

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/shyn/kiro2cc/internal/config"
	"github.com/shyn/kiro2cc/parser"
	"github.com/shyn/kiro2cc/pkg/types"
)

type Service interface {
	ToCodeWhisperer(req *types.AnthropicRequest) (*types.CodeWhispererRequest, error)
	FromCodeWhisperer(resp []byte, model string) (map[string]any, error)
}

type service struct {
	config *config.Config
}

func NewService(cfg *config.Config) Service {
	return &service{
		config: cfg,
	}
}

func (s *service) ToCodeWhisperer(anthropicReq *types.AnthropicRequest) (*types.CodeWhispererRequest, error) {
	cwReq := &types.CodeWhispererRequest{
		ProfileArn: s.config.CodeWhisperer.ProfileArn,
	}

	cwReq.ConversationState.ChatTriggerType = config.ChatTriggerType
	cwReq.ConversationState.ConversationId = generateUUID()

	lastMessage := anthropicReq.Messages[len(anthropicReq.Messages)-1]
	cwReq.ConversationState.CurrentMessage.UserInputMessage.Content = getMessageContent(lastMessage.Content)
	cwReq.ConversationState.CurrentMessage.UserInputMessage.ModelId = config.ModelMapping[anthropicReq.Model]
	cwReq.ConversationState.CurrentMessage.UserInputMessage.Origin = config.Origin

	if len(anthropicReq.Tools) > 0 {
		tools := make([]types.CodeWhispererTool, 0, len(anthropicReq.Tools))
		for _, tool := range anthropicReq.Tools {
			cwTool := types.CodeWhispererTool{}
			cwTool.ToolSpecification.Name = tool.Name
			cwTool.ToolSpecification.Description = tool.Description
			cwTool.ToolSpecification.InputSchema = types.InputSchema{
				Json: tool.InputSchema,
			}
			tools = append(tools, cwTool)
		}
		cwReq.ConversationState.CurrentMessage.UserInputMessage.UserInputMessageContext.Tools = tools
	}

	s.buildHistory(cwReq, anthropicReq)

	return cwReq, nil
}

func (s *service) buildHistory(cwReq *types.CodeWhispererRequest, anthropicReq *types.AnthropicRequest) {
	if len(anthropicReq.System) == 0 && len(anthropicReq.Messages) <= 1 {
		return
	}

	var history []any

	assistantDefaultMsg := types.HistoryAssistantMessage{}
	assistantDefaultMsg.AssistantResponseMessage.Content = config.SystemMessageResponse
	assistantDefaultMsg.AssistantResponseMessage.ToolUses = make([]any, 0)

	for _, sysMsg := range anthropicReq.System {
		userMsg := types.HistoryUserMessage{}
		userMsg.UserInputMessage.Content = sysMsg.Text
		userMsg.UserInputMessage.ModelId = config.ModelMapping[anthropicReq.Model]
		userMsg.UserInputMessage.Origin = config.Origin
		history = append(history, userMsg)
		history = append(history, assistantDefaultMsg)
	}

	for i := 0; i < len(anthropicReq.Messages)-1; i++ {
		if anthropicReq.Messages[i].Role == "user" {
			userMsg := types.HistoryUserMessage{}
			userMsg.UserInputMessage.Content = getMessageContent(anthropicReq.Messages[i].Content)
			userMsg.UserInputMessage.ModelId = config.ModelMapping[anthropicReq.Model]
			userMsg.UserInputMessage.Origin = config.Origin
			history = append(history, userMsg)

			if i+1 < len(anthropicReq.Messages)-1 && anthropicReq.Messages[i+1].Role == "assistant" {
				assistantMsg := types.HistoryAssistantMessage{}
				assistantMsg.AssistantResponseMessage.Content = getMessageContent(anthropicReq.Messages[i+1].Content)
				assistantMsg.AssistantResponseMessage.ToolUses = make([]any, 0)
				history = append(history, assistantMsg)
				i++
			}
		}
	}

	cwReq.ConversationState.History = history
}

func (s *service) FromCodeWhisperer(resp []byte, model string) (map[string]any, error) {
	respBodyStr := string(resp)

	if strings.Contains(respBodyStr, "Improperly formed request.") {
		return nil, &TranslationError{
			Type:    "invalid_request",
			Message: "Request format error",
			Detail:  respBodyStr,
		}
	}

	events := parser.ParseEvents(resp)

	context := ""
	toolName := ""
	toolUseId := ""
	contexts := []map[string]any{}
	partialJsonStr := ""

	for _, event := range events {
		if event.Data != nil {
			if dataMap, ok := event.Data.(map[string]any); ok {
				switch dataMap["type"] {
				case "content_block_start":
					context = ""
				case "content_block_delta":
					if delta, ok := dataMap["delta"]; ok {
						if deltaMap, ok := delta.(map[string]any); ok {
							switch deltaMap["type"] {
							case "text_delta":
								if text, ok := deltaMap["text"]; ok {
									context += text.(string)
								}
							case "input_json_delta":
								toolUseId = deltaMap["id"].(string)
								toolName = deltaMap["name"].(string)
								if partialJson, ok := deltaMap["partial_json"]; ok {
									if strPtr, ok := partialJson.(*string); ok && strPtr != nil {
										partialJsonStr += *strPtr
									} else if str, ok := partialJson.(string); ok {
										partialJsonStr += str
									} else {
										log.Println("partial_json is not string or *string")
									}
								}
							}
						}
					}
				case "content_block_stop":
					if index, ok := dataMap["index"]; ok {
						switch index {
						case 1:
							toolInput := map[string]interface{}{}
							if err := json.Unmarshal([]byte(partialJsonStr), &toolInput); err != nil {
								log.Printf("json unmarshal error: %s", err.Error())
							}
							contexts = append(contexts, map[string]interface{}{
								"type":  "tool_use",
								"id":    toolUseId,
								"name":  toolName,
								"input": toolInput,
							})
						case 0:
							contexts = append(contexts, map[string]interface{}{
								"text": context,
								"type": "text",
							})
						}
					}
				}
			}
		}
	}

	return map[string]any{
		"content":       contexts,
		"model":         model,
		"role":          "assistant",
		"stop_reason":   "end_turn",
		"stop_sequence": nil,
		"type":          "message",
		"usage": map[string]any{
			"input_tokens":  len(context),
			"output_tokens": len(context),
		},
	}, nil
}

type TranslationError struct {
	Type    string
	Message string
	Detail  string
}

func (e *TranslationError) Error() string {
	return e.Message
}