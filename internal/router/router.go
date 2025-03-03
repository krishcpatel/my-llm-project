package router

import (
	"net/http"

	"github.com/krishcpatel/my-llm-project/internal/controllers"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// static files
	fs := http.FileServer(http.Dir("internal/views/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// chat endpoints
	mux.HandleFunc("/chat", controllers.ChatPage)
	mux.HandleFunc("/chat/stream", controllers.ChatStream)
	mux.HandleFunc("/chat/create", controllers.CreateConversation)

	return mux
}
