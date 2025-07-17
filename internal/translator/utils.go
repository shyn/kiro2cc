package translator

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/shyn/kiro2cc/internal/config"
	"github.com/shyn/kiro2cc/pkg/types"
)

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant bits
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func getMessageContent(content any) string {
	switch v := content.(type) {
	case string:
		if len(v) == 0 {
			return config.DefaultFallbackContent
		}
		return v
	case []interface{}:
		var texts []string
		for _, block := range v {
			if m, ok := block.(map[string]interface{}); ok {
				var cb types.ContentBlock
				if data, err := json.Marshal(m); err == nil {
					if err := json.Unmarshal(data, &cb); err == nil {
						switch cb.Type {
						case "tool_result":
							if cb.Content != nil {
								texts = append(texts, *cb.Content)
							}
						case "text":
							if cb.Text != nil {
								texts = append(texts, *cb.Text)
							}
						}
					}
				}
			}
		}
		if len(texts) == 0 {
			s, err := json.Marshal(content)
			if err != nil {
				return config.DefaultFallbackContent
			}
			log.Printf("uncatch: %s", string(s))
			return config.DefaultFallbackContent
		}
		return strings.Join(texts, "\n")
	default:
		s, err := json.Marshal(content)
		if err != nil {
			return config.DefaultFallbackContent
		}
		log.Printf("uncatch: %s", string(s))
		return config.DefaultFallbackContent
	}
}