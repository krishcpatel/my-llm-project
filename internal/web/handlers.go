package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/krishcpatel/my-llm-project/internal/ai"
	"github.com/krishcpatel/my-llm-project/internal/db"
)

// ChatMessage is used for building conversation context.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// dbConn is a package-level variable for DB connection.
var dbConn *sql.DB

// SetDBConn sets the global db connection.
func SetDBConn(conn *sql.DB) {
	dbConn = conn
}

var tmpl = template.Must(template.ParseFiles("internal/web/templates/chat.html"))

// ChatPage serves the main chat page.
func ChatPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := tmpl.Execute(w, nil)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// buildPromptFromMessages builds a prompt string from a slice of ChatMessage.
func buildPromptFromMessages(msgs []ChatMessage) string {
	prompt := ""
	for _, m := range msgs {
		if m.Role == "user" {
			prompt += "User: " + m.Content + "\n"
		} else {
			prompt += "Assistant: " + m.Content + "\n"
		}
	}
	prompt += "\nAssistant:"
	return prompt
}

// ChatStream handles SSE requests. It accepts an optional "conv" query parameter.
// If provided, it builds a prompt from the conversation; if not, it falls back to "prompt".
func ChatStream(w http.ResponseWriter, r *http.Request) {
	// Get conversation_id; default to 1 if missing.
	conversationIDStr := r.URL.Query().Get("conversation_id")
	var conversationID int
	if conversationIDStr == "" {
		conversationID = 1
	} else {
		cid, err := strconv.Atoi(conversationIDStr)
		if err != nil {
			http.Error(w, "Invalid conversation_id", http.StatusBadRequest)
			return
		}
		conversationID = cid
	}

	// Attempt to get the conversation JSON (if present).
	convJSON := r.URL.Query().Get("conv")
	var combinedPrompt string
	if convJSON != "" {
		var conversation []ChatMessage
		if err := json.Unmarshal([]byte(convJSON), &conversation); err != nil {
			log.Println("Error parsing conversation JSON, falling back to 'prompt' parameter:", err)
			combinedPrompt = r.URL.Query().Get("prompt")
		} else {
			combinedPrompt = buildPromptFromMessages(conversation)
		}
	} else {
		combinedPrompt = r.URL.Query().Get("prompt")
	}

	if combinedPrompt == "" {
		http.Error(w, "No prompt provided", http.StatusBadRequest)
		return
	}

	// Optionally, insert the user message into DB if using the conv parameter;
	// For this example, if conv was provided, we assume the user message is already stored.
	// Otherwise, if only prompt is provided, you might insert that as a new user message.
	if convJSON == "" {
		// Insert the prompt as a user message.
		_, err := db.InsertChatMessage(dbConn, conversationID, "user", combinedPrompt)
		if err != nil {
			log.Println("Error storing user message:", err)
			http.Error(w, "Failed to store user message", http.StatusInternalServerError)
			return
		}
	}

	// Set SSE headers.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Call Ollama with the combined prompt.
	resultChan, errChan := ai.StreamOllamaChunks("deepseek-r1:14b", combinedPrompt)

	// Create a ticker for keep-alive pings.
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	var assistantBuf string

streamLoop:
	for {
		select {
		case partial, ok := <-resultChan:
			if !ok {
				// No more chunks; send done event.
				fmt.Fprintf(w, "event: done\ndata:\n\n")
				flusher.Flush()
				break streamLoop
			}
			// Send partial text as SSE data.
			fmt.Fprintf(w, "data: %s\n\n", partial)
			flusher.Flush()

			assistantBuf += partial

		case err, ok := <-errChan:
			if !ok {
				// errChan closed normally.
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
			// Send a keep-alive ping to prevent idle timeouts.
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		}
	}

	// After streaming, store the assistant's complete response in the database.
	if assistantBuf != "" {
		_, err := db.InsertChatMessage(dbConn, conversationID, "assistant", assistantBuf)
		if err != nil {
			log.Println("Error storing assistant message:", err)
		}
	}
}
