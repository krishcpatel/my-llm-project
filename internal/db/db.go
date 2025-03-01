package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // Postgres driver
)

func OpenDB(host, port, user, password, dbname string) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	return sql.Open("postgres", dsn)
}

// InsertChatMessage inserts a single user or assistant message into the DB.
// Returns the auto-generated ID or an error.
func InsertChatMessage(db *sql.DB, conversationID int, role, content string) (int, error) {
	var id int
	query := `
	  INSERT INTO chat_messages (conversation_id, role, content)
	  VALUES ($1, $2, $3)
	  RETURNING id
	`
	err := db.QueryRow(query, conversationID, role, content).Scan(&id)
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

// Message is a struct that mirrors the columns in chat_messages table.
type Message struct {
	ID        int
	Role      string
	Content   string
	CreatedAt string // or time.Time
}
