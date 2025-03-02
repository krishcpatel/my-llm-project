package main

import (
	"log"
	"net/http"
	"os"

	"github.com/krishcpatel/my-llm-project/internal/db"
	"github.com/krishcpatel/my-llm-project/internal/web"
)

func main() {
	// Load database credentials
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	// Read the chat model from environment, defaulting if not set
	chatModel := os.Getenv("CHAT_MODEL")
	if chatModel == "" {
		chatModel = "deepseek-r1:14b"
	}

	// Open the database connection
	dbConn, err := db.OpenDB(host, port, user, password, dbname)
	if err != nil {
		log.Fatalf("Cannot open DB: %v\n", err)
	}
	defer dbConn.Close()

	// Set DB connection globally
	db.SetDBConn(dbConn)

	// Set the chat model in the web package (so that SSE uses it)
	web.SetChatModel(chatModel)

	// Register HTTP routes
	mux := http.NewServeMux()
	web.RegisterRoutes(mux)

	// Start server
	log.Println("Starting server on :4000...")
	err = http.ListenAndServe(":4000", mux)
	if err != nil {
		log.Fatal(err)
	}
}
