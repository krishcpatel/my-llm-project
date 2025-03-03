package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// ComputeEmbedding calls the Ollama embedding API to generate an embedding for the given text.
func ComputeEmbedding(text string) ([]float64, error) {
	embeddingModel := os.Getenv("EMBEDDING_MODEL")
	if embeddingModel == "" {
		embeddingModel = "deekseek-r1:7b"
	}

	payload := map[string]interface{}{
		"model":    embeddingModel,
		"input":    text,
		"truncate": true,
	}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embed payload: %w", err)
	}

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

// FormatVector turns []float64 into "[x,y,z]" for PGVector
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
