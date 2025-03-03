package controllers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/krishcpatel/my-llm-project/internal/models"
	"github.com/krishcpatel/my-llm-project/internal/repositories"
)

// ChatPageData for the template
type ChatPageData struct {
	ConversationID string
	Messages       []models.Message
}

// Load template (views/templates/chat.html)
var chatTmpl = template.Must(template.ParseFiles("internal/views/templates/chat.html"))

func ChatPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get conversation_id
	conversationID := r.URL.Query().Get("conversation_id")
	if conversationID == "" {
		conversationID = "1"
	}

	cid, err := strconv.Atoi(conversationID)
	if err != nil {
		log.Println("Invalid conversation_id, defaulting to 1")
		cid = 1
	}

	// fetch messages
	msgs, err := repositories.GetChatMessages(models.GetDBConn(), cid)
	if err != nil {
		log.Println("Error retrieving messages:", err)
		msgs = []models.Message{}
	}

	data := ChatPageData{
		ConversationID: conversationID,
		Messages:       msgs,
	}

	if err := chatTmpl.Execute(w, data); err != nil {
		log.Println("Error executing chat template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
