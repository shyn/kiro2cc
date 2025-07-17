package types

type TokenData struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    string `json:"expiresAt,omitempty"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type RefreshResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    string `json:"expiresAt,omitempty"`
}

type AnthropicTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

type InputSchema struct {
	Json map[string]any `json:"json"`
}

type ToolSpecification struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type CodeWhispererTool struct {
	ToolSpecification ToolSpecification `json:"toolSpecification"`
}

type HistoryUserMessage struct {
	UserInputMessage struct {
		Content string `json:"content"`
		ModelId string `json:"modelId"`
		Origin  string `json:"origin"`
	} `json:"userInputMessage"`
}

type HistoryAssistantMessage struct {
	AssistantResponseMessage struct {
		Content  string `json:"content"`
		ToolUses []any  `json:"toolUses"`
	} `json:"assistantResponseMessage"`
}

type AnthropicRequest struct {
	Model       string                    `json:"model"`
	MaxTokens   int                       `json:"max_tokens"`
	Messages    []AnthropicRequestMessage `json:"messages"`
	System      []AnthropicSystemMessage  `json:"system,omitempty"`
	Tools       []AnthropicTool           `json:"tools,omitempty"`
	Stream      bool                      `json:"stream"`
	Temperature *float64                  `json:"temperature,omitempty"`
	Metadata    map[string]any            `json:"metadata,omitempty"`
}

type AnthropicStreamResponse struct {
	Type         string `json:"type"`
	Index        int    `json:"index"`
	ContentDelta struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"delta,omitempty"`
	Content []struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"content,omitempty"`
	StopReason   string `json:"stop_reason,omitempty"`
	StopSequence string `json:"stop_sequence,omitempty"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage,omitempty"`
}

type AnthropicRequestMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type AnthropicSystemMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ContentBlock struct {
	Type      string  `json:"type"`
	Text      *string `json:"text,omitempty"`
	ToolUseId *string `json:"tool_use_id,omitempty"`
	Content   *string `json:"content,omitempty"`
	Name      *string `json:"name,omitempty"`
	Input     *any    `json:"input,omitempty"`
}

type CodeWhispererRequest struct {
	ConversationState struct {
		ChatTriggerType string `json:"chatTriggerType"`
		ConversationId  string `json:"conversationId"`
		CurrentMessage  struct {
			UserInputMessage struct {
				Content                 string `json:"content"`
				ModelId                 string `json:"modelId"`
				Origin                  string `json:"origin"`
				UserInputMessageContext struct {
					ToolResults []struct {
						Content []struct {
							Text string `json:"text"`
						} `json:"content"`
						Status    string `json:"status"`
						ToolUseId string `json:"toolUseId"`
					} `json:"toolResults,omitempty"`
					Tools []CodeWhispererTool `json:"tools,omitempty"`
				} `json:"userInputMessageContext"`
			} `json:"userInputMessage"`
		} `json:"currentMessage"`
		History []any `json:"history"`
	} `json:"conversationState"`
	ProfileArn string `json:"profileArn"`
}

type CodeWhispererEvent struct {
	ContentType string `json:"content-type"`
	MessageType string `json:"message-type"`
	Content     string `json:"content"`
	EventType   string `json:"event-type"`
}