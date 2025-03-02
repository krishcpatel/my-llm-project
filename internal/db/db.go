package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // Postgres driver
)

// Message represents a chat message.
type Message struct {
	ID        int    `json:"id,omitempty"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at,omitempty"`
}

func OpenDB(host, port, user, password, dbname string) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	return sql.Open("postgres", dsn)
}

// InsertChatMessage inserts a single message into the DB with its embedding.
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

// GetChatMessages fetches all messages for a given conversation, sorted by creation time.
func GetChatMessages(db *sql.DB, conversationID int) ([]Message, error) {
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

	var msgs []Message
	for rows.Next() {
		var m Message
		err := rows.Scan(&m.ID, &m.Role, &m.Content, &m.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("GetChatMessages scan: %w", err)
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

// InsertConversation inserts a new conversation into the conversations table.
// userID is currently nil.
func InsertConversation(db *sql.DB, userID interface{}) (int, error) {
	var id int
	query := `INSERT INTO conversations (user_id) VALUES ($1) RETURNING id`
	err := db.QueryRow(query, userID).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("InsertConversation: %w", err)
	}
	return id, nil
}

// GetRelevantMessages returns the top N messages for a conversation that are similar to the provided query embedding.
// Here, queryEmbeddingStr is a string in the format "[0.1,0.2,...]".
func GetRelevantMessages(db *sql.DB, conversationID int, queryEmbeddingStr string, topN int) ([]Message, error) {
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

	var msgs []Message
	for rows.Next() {
		var m Message
		err := rows.Scan(&m.ID, &m.Role, &m.Content, &m.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("GetRelevantMessages scan: %w", err)
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

// ConversationExists checks if a conversation exists in the conversations table.
func ConversationExists(db *sql.DB, conversationID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM conversations WHERE id = $1)`
	err := db.QueryRow(query, conversationID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("ConversationExists: %w", err)
	}
	return exists, nil
}
