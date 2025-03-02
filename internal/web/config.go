package web

var chatModel string

// SetChatModel sets the model name for chat generation and embeddings.
func SetChatModel(model string) {
	chatModel = model
}

// GetChatModel returns the configured chat model.
func GetChatModel() string {
	return chatModel
}
