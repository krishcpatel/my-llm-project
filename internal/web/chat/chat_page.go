package chat

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/krishcpatel/my-llm-project/internal/db"
)

// ChatPageData holds data for rendering the chat page.
type ChatPageData struct {
	ConversationID string
	Messages       []db.Message
}

// Load the chat template from templates/chat.html
var tmpl = template.Must(template.ParseFiles("templates/chat.html"))

// ChatPage serves the chat page and loads past messages if a conversation_id is provided.
func ChatPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get conversation_id from query; default to "1" if not provided.
	conversationID := r.URL.Query().Get("conversation_id")
	if conversationID == "" {
		conversationID = "1"
	}

	// Convert conversationID to int for DB query.
	cid, err := strconv.Atoi(conversationID)
	if err != nil {
		log.Println("Invalid conversation_id, defaulting to 1")
		cid = 1
	}

	// Fetch past messages for this conversation from the database.
	messages, err := db.GetChatMessages(db.GetDBConn(), cid)
	if err != nil {
		log.Println("Error retrieving conversation messages:", err)
		// If retrieval fails, we can default to an empty conversation.
		messages = []db.Message{}
	}

	data := ChatPageData{
		ConversationID: conversationID,
		Messages:       messages,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
