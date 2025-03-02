package chat

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/krishcpatel/my-llm-project/internal/ai"
	"github.com/krishcpatel/my-llm-project/internal/db"
	"github.com/krishcpatel/my-llm-project/internal/engines"
)

// ChatStream handles SSE chat responses.
func ChatStream(w http.ResponseWriter, r *http.Request) {
	dbConn := db.GetDBConn() // Get the global DB connection
	if dbConn == nil {
		http.Error(w, "Database connection is not initialized", http.StatusInternalServerError)
		log.Println("ERROR: dbConn is nil - Ensure SetDBConn is called in main.go")
		return
	}

	// Parse query parameters
	conversationIDStr := r.URL.Query().Get("conversation_id")
	convParam := r.URL.Query().Get("conv")
	if convParam == "" {
		http.Error(w, "Missing conversation data", http.StatusBadRequest)
		log.Println("ERROR: conv parameter is missing")
		return
	}

	// Decode and parse the conversation JSON.
	decodedConv, err := url.QueryUnescape(convParam)
	if err != nil {
		http.Error(w, "Failed to decode conversation", http.StatusBadRequest)
		log.Println("ERROR: Failed to decode conversation:", err)
		return
	}
	var conversation []db.Message
	if err := json.Unmarshal([]byte(decodedConv), &conversation); err != nil {
		http.Error(w, "Invalid conversation format", http.StatusBadRequest)
		log.Println("ERROR: Invalid conversation JSON format:", err)
		return
	}

	// Convert conversation_id; default to 1 if missing.
	conversationID := 1
	if conversationIDStr != "" {
		if cid, err := strconv.Atoi(conversationIDStr); err == nil {
			conversationID = cid
		} else {
			http.Error(w, "Invalid conversation_id", http.StatusBadRequest)
			return
		}
	}

	// Check if the conversation exists in the database.
	exists, err := db.ConversationExists(dbConn, conversationID)
	if err != nil {
		log.Println("Error checking conversation existence:", err)
		http.Error(w, "Error checking conversation", http.StatusInternalServerError)
		return
	}
	if !exists {
		// Conversation doesn't exist, so create a new conversation.
		newConvID, err := db.InsertConversation(dbConn, nil)
		if err != nil {
			log.Println("Error creating new conversation:", err)
			http.Error(w, "Failed to create conversation", http.StatusInternalServerError)
			return
		}
		conversationID = newConvID
		// Optionally, you might want to inform the client of the new conversation ID.
	}

	// Store the user message (last message in conversation)
	userPrompt := conversation[len(conversation)-1].Content

	// Compute user embedding (using your engines package)
	userEmbedding, err := engines.ComputeEmbedding(userPrompt)
	if err != nil {
		log.Println("Error computing user embedding:", err)
		// Fallback: use a zero vector (adjust dimension as needed, e.g., 768)
		userEmbedding = make([]float64, 768)
	}
	userEmbeddingStr := engines.FormatVector(userEmbedding)
	_, err = db.InsertChatMessage(dbConn, conversationID, "user", userPrompt, userEmbeddingStr)
	if err != nil {
		log.Println("Error storing user message:", err)
		http.Error(w, "Failed to store user message", http.StatusInternalServerError)
		return
	}

	// Build full conversation context for prompt
	combinedPrompt := buildPrompt(conversation)

	// Decide whether to use RAG prompt
	ragPrompt, err := engines.BuildRAGPrompt(conversationID, userPrompt, 3)
	if err != nil {
		log.Println("Error building RAG prompt, falling back to full conversation:", err)
		// combinedPrompt remains unchanged.
	} else {
		combinedPrompt = ragPrompt
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Stream response from the AI service
	model := os.Getenv("CHAT_MODEL")
	if model == "" {
		model = "deepseek-r1:14b"
	}
	resultChan, errChan := ai.StreamOllamaChunks(model, combinedPrompt)
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
			time.Sleep(100 * time.Millisecond)
			break streamLoop

		case <-ticker.C:
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		}
	}

	// Compute assistant embedding and store assistant response
	if assistantBuf != "" {
		assistantEmbedding, err := engines.ComputeEmbedding(assistantBuf)
		if err != nil {
			log.Println("Error computing assistant embedding:", err)
			assistantEmbedding = make([]float64, 768)
		}
		assistantEmbeddingStr := engines.FormatVector(assistantEmbedding)
		_, err = db.InsertChatMessage(dbConn, conversationID, "assistant", assistantBuf, assistantEmbeddingStr)
		if err != nil {
			log.Println("Error storing assistant message:", err)
		}
	}
}
