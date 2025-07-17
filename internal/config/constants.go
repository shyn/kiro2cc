package config

var ModelMapping = map[string]string{
	"claude-sonnet-4-20250514":  "CLAUDE_SONNET_4_20250514_V1_0",
	"claude-3-5-haiku-20241022": "CLAUDE_3_7_SONNET_20250219_V1_0",
}

const (
	DefaultFallbackContent = "answer for user qeustion"
	ChatTriggerType        = "MANUAL"
	Origin                 = "AI_EDITOR"
	SystemMessageResponse  = "I will follow these instructions"
)