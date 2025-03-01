package main

import (
	"log"
	"net/http"

	"github.com/krishcpatel/my-llm-project/internal/web"
)

func main() {
	// Simple mux for routes
	mux := http.NewServeMux()

	// The chat page
	mux.HandleFunc("/chat", web.ChatPage)
	// The SSE streaming endpoint
	mux.HandleFunc("/chat/stream", web.ChatStream)

	// Serve on localhost:4000
	log.Println("Starting server on :4000...")
	err := http.ListenAndServe(":4000", mux)
	if err != nil {
		log.Fatal(err)
	}
}
