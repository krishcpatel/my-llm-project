package models

// Conversation represents a conversation.
type Conversation struct {
	ID     int         `json:"id,omitempty"`
	UserID interface{} `json:"user_id,omitempty"`
	// Additional fields...
}
