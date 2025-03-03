package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/krishcpatel/my-llm-project/internal/config"
	"github.com/krishcpatel/my-llm-project/internal/models"
	"github.com/krishcpatel/my-llm-project/internal/router"
)

func main() {
	// Load environment variables & config
	cfg, err := config.LoadConfig() // implement as you wish
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Open the database connection
	dbConn, err := models.OpenDB(
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)
	if err != nil {
		log.Fatalf("Cannot open DB: %v\n", err)
	}
	defer dbConn.Close()

	// Set DB connection globally (or use DI in your app)
	models.SetDBConn(dbConn)

	// Start server
	r := router.NewRouter()
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Starting server on %s...\n", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
