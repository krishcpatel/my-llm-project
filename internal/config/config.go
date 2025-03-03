package config

import (
	"os"
)

type Config struct {
	Port       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	ChatModel      string
	EmbeddingModel string
}

func LoadConfig() (*Config, error) {
	// read environment variables or fallback to defaults
	return &Config{
		Port:           defaultString(os.Getenv("PORT"), "4000"),
		DBHost:         defaultString(os.Getenv("DB_HOST"), "localhost"),
		DBPort:         defaultString(os.Getenv("DB_PORT"), "5432"),
		DBUser:         os.Getenv("DB_USER"),
		DBPassword:     os.Getenv("DB_PASSWORD"),
		DBName:         os.Getenv("DB_NAME"),
		ChatModel:      defaultString(os.Getenv("CHAT_MODEL"), "deepseek-r1:14b"),
		EmbeddingModel: defaultString(os.Getenv("EMBEDDING_MODEL"), "deekseek-r1:7b"),
	}, nil
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
