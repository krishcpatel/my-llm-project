package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/krishcpatel/my-llm-project/internal/models"
	"github.com/krishcpatel/my-llm-project/internal/repositories"
	"github.com/krishcpatel/my-llm-project/internal/services"
)

func ChatStream(w http.ResponseWriter, r *http.Request) {
	dbConn := models.GetDBConn()
	if dbConn == nil {
		http.Error(w, "DB not initialized", http.StatusInternalServerError)
		log.Println("ERROR: dbConn is nil")
		return
	}

	convIDStr := r.URL.Query().Get("conversation_id")
	convParam := r.URL.Query().Get("conv")
	if convParam == "" {
		http.Error(w, "Missing conversation data", http.StatusBadRequest)
		return
	}

	decodedConv, err := url.QueryUnescape(convParam)
	if err != nil {
		http.Error(w, "Failed to decode conv", http.StatusBadRequest)
		return
	}

	var conversation []models.Message
	if err := json.Unmarshal([]byte(decodedConv), &conversation); err != nil {
		http.Error(w, "Invalid conversation format", http.StatusBadRequest)
		return
	}

	// default conversation ID
	conversationID := 1
	if convIDStr != "" {
		cid, parseErr := strconv.Atoi(convIDStr)
		if parseErr == nil {
			conversationID = cid
		} else {
			http.Error(w, "Invalid conversation_id", http.StatusBadRequest)
			return
		}
	}

	// ensure conversation exists
	exists, err := repositories.ConversationExists(dbConn, conversationID)
	if err != nil {
		log.Println("Error checking conversation:", err)
		http.Error(w, "Conversation check error", http.StatusInternalServerError)
		return
	}

	if !exists {
		newCID, err := repositories.InsertConversation(dbConn, nil)
		if err != nil {
			http.Error(w, "Failed to create conversation", http.StatusInternalServerError)
			return
		}
		conversationID = newCID
	}

	// store user message
	userPrompt := conversation[len(conversation)-1].Content
	userEmbedding, err := services.ComputeEmbedding(userPrompt)
	if err != nil {
		log.Println("Error computing user embedding:", err)
		userEmbedding = make([]float64, 768) // fallback
	}
	userEmbeddingStr := services.FormatVector(userEmbedding)
	if _, err := repositories.InsertChatMessage(dbConn, conversationID, "user", userPrompt, userEmbeddingStr); err != nil {
		http.Error(w, "Failed to store user message", http.StatusInternalServerError)
		return
	}

	combinedPrompt, err := services.BuildCombinedPrompt(
		conversationID,
		userPrompt,
		10, // e.g. last 10 messages
		3,  // e.g. top 3 relevant older messages
	)
	if err != nil {
		log.Println("Error building combined prompt:", err)
		// fallback or return
		combinedPrompt = userPrompt // minimal fallback
	}

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// stream from the AI
	model := os.Getenv("CHAT_MODEL")
	if model == "" {
		model = "deepseek-r1:14b"
	}

	// SSE streaming code...
	log.Println("DEBUG: Final prompt being sent to model:\n", combinedPrompt)
	resultChan, errChan := services.StreamOllamaChunks(model, combinedPrompt)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	var assistantBuf string

streamLoop:
	for {
		select {
		case partial, ok := <-resultChan:
			if !ok {
				fmt.Fprintf(w, "event: done\ndata:\n\n")
				flusher.Flush()
				break streamLoop
			}
			fmt.Fprintf(w, "data: %s\n\n", partial)
			flusher.Flush()
			assistantBuf += partial

		case err, ok := <-errChan:
			if !ok {
				fmt.Fprintf(w, "event: done\ndata:\n\n")
				flusher.Flush()
				break streamLoop
			}
			if err != nil {
				log.Println("Error streaming from Ollama:", err)
				fmt.Fprintf(w, "data: [Error: %v]\n\n", err)
				flusher.Flush()
			}
			break streamLoop

		case <-ticker.C:
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		}
	}

	// store assistant response
	if assistantBuf != "" {
		assistEmbedding, err := services.ComputeEmbedding(assistantBuf)
		if err != nil {
			log.Println("Error computing assistant embedding:", err)
			assistEmbedding = make([]float64, 768)
		}
		assistEmbeddingStr := services.FormatVector(assistEmbedding)
		if _, err := repositories.InsertChatMessage(dbConn, conversationID, "assistant", assistantBuf, assistEmbeddingStr); err != nil {
			log.Println("Error storing assistant message:", err)
		}
	}
}
