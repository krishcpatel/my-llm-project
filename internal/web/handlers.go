package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/krishcpatel/my-llm-project/internal/ai"
)

var tmpl = template.Must(template.ParseFiles("internal/web/templates/chat.html"))

type ChatMessage struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"` // message text
}

// ChatPage ...
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

// ChatStream with conversation memory in query param.
func ChatStream(w http.ResponseWriter, r *http.Request) {
	convJSON := r.URL.Query().Get("conv")
	if convJSON == "" {
		http.Error(w, "No conversation provided", http.StatusBadRequest)
		return
	}

	var conversation []ChatMessage
	if err := json.Unmarshal([]byte(convJSON), &conversation); err != nil {
		http.Error(w, "Invalid conversation JSON", http.StatusBadRequest)
		return
	}

	// Build a single prompt from the conversation.
	// This is naive. You might do a more refined system message + user/assistant pairs.
	prompt := buildPrompt(conversation)

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Stream partial text from Ollama
	resultChan, errChan := ai.StreamOllamaChunks("deepseek-r1:14b", prompt)

	// Keep-alive ticker if you want
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case partial, ok := <-resultChan:
			if !ok {
				// No more chunks => done
				fmt.Fprintf(w, "event: done\ndata:\n\n")
				flusher.Flush()
				return
			}
			// SSE chunk
			fmt.Fprintf(w, "data: %s\n\n", partial)
			flusher.Flush()

		case err, ok := <-errChan:
			if !ok {
				// errChan closed => normal
				fmt.Fprintf(w, "event: done\ndata:\n\n")
				flusher.Flush()
				return
			}
			if err == nil {
				continue
			}
			log.Println("Error streaming from Ollama:", err)
			fmt.Fprintf(w, "data: [Error: %v]\n\n", err)
			flusher.Flush()
			time.Sleep(100 * time.Millisecond)
			return

		case <-ticker.C:
			// Keep-alive comment
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}

// buildPrompt just concatenates the user & assistant messages.
// You might refine this to handle roles or system instructions more elegantly.
func buildPrompt(conv []ChatMessage) string {
	prompt := ""
	for _, msg := range conv {
		if msg.Role == "user" {
			prompt += "User: " + msg.Content + "\n"
		} else {
			prompt += "Assistant: " + msg.Content + "\n"
		}
	}
	prompt += "\nAssistant:"
	return prompt
}
