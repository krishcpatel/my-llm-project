package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/krishcpatel/my-llm-project/internal/models"
	"github.com/krishcpatel/my-llm-project/internal/repositories"
)

func CreateConversation(w http.ResponseWriter, r *http.Request) {
	convID, err := repositories.InsertConversation(models.GetDBConn(), nil)
	if err != nil {
		log.Println("Error creating new conversation:", err)
		http.Error(w, "Failed to create conversation", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"conversation_id": convID,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
