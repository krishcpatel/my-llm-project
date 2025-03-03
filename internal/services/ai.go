package services

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OllamaChunk struct {
	Model    string `json:"model,omitempty"`
	Response string `json:"response,omitempty"`
	Done     bool   `json:"done,omitempty"`
}

// StreamOllamaChunks returns partial response chunks and any errors.
func StreamOllamaChunks(model, prompt string) (<-chan string, <-chan error) {
	out := make(chan string)
	errs := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errs)

		reqData := map[string]string{
			"model":  model,
			"prompt": prompt,
		}
		reqBody, _ := json.Marshal(reqData)

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

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				errs <- fmt.Errorf("error reading Ollama stream: %w", err)
				return
			}

			var chunk OllamaChunk
			if err := json.Unmarshal(line, &chunk); err != nil {
				// skip malformed lines
				continue
			}

			out <- chunk.Response

			if chunk.Done {
				break
			}
		}
	}()

	return out, errs
}
