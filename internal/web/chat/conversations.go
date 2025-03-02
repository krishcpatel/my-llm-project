package chat

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/krishcpatel/my-llm-project/internal/db"
)

// CreateConversation creates a new conversation record and returns its ID.
func CreateConversation(w http.ResponseWriter, r *http.Request) {
	// For now, user_id is nil.
	convID, err := db.InsertConversation(db.GetDBConn(), nil)
	if err != nil {
		log.Println("Error creating conversation:", err)
		http.Error(w, "Failed to create conversation", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"conversation_id": convID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
