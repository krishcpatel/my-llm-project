package models

// Message represents a chat message.
type Message struct {
	ID        int    `json:"id,omitempty"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at,omitempty"`
}
