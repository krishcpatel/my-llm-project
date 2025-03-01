package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OllamaChunk represents each NDJSON line from Ollama
type OllamaChunk struct {
	Model    string `json:"model,omitempty"`
	Response string `json:"response,omitempty"`
	Done     bool   `json:"done,omitempty"`
}

// StreamOllamaChunks returns two channels:
//   - a channel of partial response strings
//   - a channel for errors
func StreamOllamaChunks(model, prompt string) (<-chan string, <-chan error) {
	out := make(chan string)
	errs := make(chan error, 1) // buffered so we can send an error and close

	go func() {
		defer close(out)
		defer close(errs)

		// Prepare request body
		reqData := map[string]string{
			"model":  model,
			"prompt": prompt,
		}
		reqBody, _ := json.Marshal(reqData)

		// Make POST request to Ollama
		resp, err := http.Post("http://localhost:11411/api/generate", "application/json", bytes.NewReader(reqBody))
		if err != nil {
			errs <- fmt.Errorf("ollama request failed: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errs <- fmt.Errorf("ollama returned status code %d", resp.StatusCode)
			return
		}

		// Read NDJSON lines
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					// No more data
					break
				}
				errs <- fmt.Errorf("error reading Ollama stream: %w", err)
				return
			}

			// Parse JSON chunk
			var chunk OllamaChunk
			if err := json.Unmarshal(line, &chunk); err != nil {
				// skip malformed lines
				continue
			}

			// Send partial text to out channel
			out <- chunk.Response

			if chunk.Done {
				// The model indicated it's finished
				break
			}
		}
	}()

	return out, errs
}
