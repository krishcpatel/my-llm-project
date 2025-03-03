package repositories

import (
	"database/sql"
	"fmt"

	"github.com/krishcpatel/my-llm-project/internal/models"
)

// InsertConversation inserts a new conversation row.
func InsertConversation(db *sql.DB, userID interface{}) (int, error) {
	var id int
	query := `INSERT INTO conversations (user_id) VALUES ($1) RETURNING id`
	err := db.QueryRow(query, userID).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("InsertConversation: %w", err)
	}
	return id, nil
}

// ConversationExists checks if a conversation exists.
func ConversationExists(db *sql.DB, conversationID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM conversations WHERE id = $1)`
	err := db.QueryRow(query, conversationID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("ConversationExists: %w", err)
	}
	return exists, nil
}

// InsertChatMessage inserts a single message with embedding.
func InsertChatMessage(db *sql.DB, conversationID int, role, content, embedding string) (int, error) {
	var id int
	query := `
        INSERT INTO chat_messages (conversation_id, role, content, embedding)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `
	err := db.QueryRow(query, conversationID, role, content, embedding).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("InsertChatMessage: %w", err)
	}
	return id, nil
}

// GetChatMessages fetches all messages for a conversation.
func GetChatMessages(db *sql.DB, conversationID int) ([]models.Message, error) {
	rows, err := db.Query(`
        SELECT id, role, content, created_at
        FROM chat_messages
        WHERE conversation_id = $1
        ORDER BY id
    `, conversationID)
	if err != nil {
		return nil, fmt.Errorf("GetChatMessages: %w", err)
	}
	defer rows.Close()

	var msgs []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("GetChatMessages scan: %w", err)
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

// GetRelevantMessages uses PGVector to find the top N relevant messages.
func GetRelevantMessages(db *sql.DB, conversationID int, queryEmbeddingStr string, topN int) ([]models.Message, error) {
	query := `
        SELECT id, role, content, created_at
        FROM chat_messages
        WHERE conversation_id = $1
        ORDER BY embedding <-> $2
        LIMIT $3
    `
	rows, err := db.Query(query, conversationID, queryEmbeddingStr, topN)
	if err != nil {
		return nil, fmt.Errorf("GetRelevantMessages: %w", err)
	}
	defer rows.Close()

	var msgs []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("GetRelevantMessages scan: %w", err)
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

// GetLastMessages fetches the last N messages (descending), but we often want them in ascending order after fetching.
func GetLastMessages(db *sql.DB, conversationID, limit int) ([]models.Message, error) {
	query := `
        SELECT id, role, content, created_at
        FROM chat_messages
        WHERE conversation_id = $1
        ORDER BY id DESC
        LIMIT $2
    `
	rows, err := db.Query(query, conversationID, limit)
	if err != nil {
		return nil, fmt.Errorf("GetLastMessages: %w", err)
	}
	defer rows.Close()

	var msgs []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("GetLastMessages scan: %w", err)
		}
		msgs = append(msgs, m)
	}

	// Since we queried in descending order, reverse them to return ascending
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	return msgs, rows.Err()
}
