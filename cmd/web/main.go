package main

import (
	"log"
	"net/http"
	"os"

	"github.com/krishcpatel/my-llm-project/internal/db"
	"github.com/krishcpatel/my-llm-project/internal/web"
)

func main() {
	// Read DB connection info from environment variables or hardcode as needed
	host := os.Getenv("DB_HOST")         // e.g. "localhost"
	port := os.Getenv("DB_PORT")         // e.g. "5432"
	user := os.Getenv("DB_USER")         // e.g. "myuser"
	password := os.Getenv("DB_PASSWORD") // e.g. "mypassword"
	dbname := os.Getenv("DB_NAME")       // e.g. "mydb"

	// Open the database
	dbConn, err := db.OpenDB(host, port, user, password, dbname)
	if err != nil {
		log.Fatalf("Cannot open DB: %v\n", err)
	}
	defer dbConn.Close()

	// Set up a quick ping
	if err := dbConn.Ping(); err != nil {
		log.Fatalf("Cannot ping DB: %v\n", err)
	}
	log.Println("Connected to Postgres successfully.")

	// Pass dbConn to the web package so handlers can use it
	web.SetDBConn(dbConn)

	// Create a simple mux
	mux := http.NewServeMux()

	// Route for chat page (HTML)
	mux.HandleFunc("/chat", web.ChatPage)

	// SSE streaming endpoint
	mux.HandleFunc("/chat/stream", web.ChatStream)

	// Start server
	log.Println("Starting server on :4000...")
	err = http.ListenAndServe(":4000", mux)
	if err != nil {
		log.Fatal(err)
	}
}
