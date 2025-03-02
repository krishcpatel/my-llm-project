package web

import (
	"net/http"

	"github.com/krishcpatel/my-llm-project/internal/web/chat"
)

// RegisterRoutes sets up HTTP handlers for the web server
func RegisterRoutes(mux *http.ServeMux) {
	// Serve static files (CSS, JS)
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Chat endpoints
	mux.HandleFunc("/chat", chat.ChatPage)
	mux.HandleFunc("/chat/stream", chat.ChatStream)
	// New conversation creation endpoint
	mux.HandleFunc("/chat/create", chat.CreateConversation)
}
