package chat

import (
	"fmt"

	"github.com/krishcpatel/my-llm-project/internal/db"
)

// buildPrompt creates a full conversation prompt from stored messages.
func buildPrompt(msgs []db.Message) string {
	prompt := ""
	for _, m := range msgs {
		prompt += fmt.Sprintf("%s: %s\n", m.Role, m.Content)
	}
	return prompt + "\nAssistant:"
}
