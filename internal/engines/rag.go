package engines

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/krishcpatel/my-llm-project/internal/db"
)

// ComputeEmbedding calls the Ollama embedding API to generate an embedding for the given text.
func ComputeEmbedding(text string) ([]float64, error) {
	// Read the embedding model from environment; default to "all-minilm" if not set.
	embeddingModel := os.Getenv("EMBEDDING_MODEL")
	if embeddingModel == "" {
		embeddingModel = "deekseek-r1:7b"
	}

	// Prepare payload (note: this matches your curl example)
	payload := map[string]interface{}{
		"model":    embeddingModel,
		"input":    text,
		"truncate": true,
	}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embed payload: %w", err)
	}

	// Increase timeout (e.g., 15 seconds) to allow the service more time to respond.
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post("http://localhost:11411/api/embed", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding service returned status code %d", resp.StatusCode)
	}

	var embedResp struct {
		Model      string      `json:"model"`
		Embeddings [][]float64 `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode embedding response: %w", err)
	}
	if len(embedResp.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}
	return embedResp.Embeddings[0], nil
}

// FormatVector converts a []float64 into a string literal that PGVector accepts.
func FormatVector(vec []float64) string {
	s := "["
	for i, v := range vec {
		if i > 0 {
			s += ","
		}
		s += fmt.Sprintf("%f", v)
	}
	s += "]"
	return s
}

// RetrieveRelevantContext computes the query embedding for the new query using the Ollama API,
// formats the embedding, and then retrieves the top N relevant past messages using PGVector similarity search.
func RetrieveRelevantContext(conversationID int, query string, topN int) (string, error) {
	// Compute the embedding for the query text using the Ollama embedding endpoint.
	queryEmbedding, err := ComputeEmbedding(query)
	if err != nil {
		return "", fmt.Errorf("ComputeEmbedding error: %w", err)
	}

	// Convert the embedding slice into a string literal accepted by PGVector.
	embeddingStr := FormatVector(queryEmbedding)

	// Retrieve the most relevant past messages.
	msgs, err := db.GetRelevantMessages(db.GetDBConn(), conversationID, embeddingStr, topN)
	if err != nil {
		return "", fmt.Errorf("GetRelevantMessages: %w", err)
	}

	// Build a context string from the retrieved messages.
	contextText := ""
	for _, msg := range msgs {
		contextText += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
	}
	return contextText, nil
}

// BuildRAGPrompt builds a prompt that includes retrieved context along with the new user query.
func BuildRAGPrompt(conversationID int, newQuery string, topN int) (string, error) {
	context, err := RetrieveRelevantContext(conversationID, newQuery, topN)
	if err != nil {
		return "", err
	}

	// Construct the prompt; adjust formatting as needed.
	prompt := fmt.Sprintf("Context:\n%s\nUser: %s\nAssistant:", context, newQuery)
	return prompt, nil
}
