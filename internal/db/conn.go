package db

import (
	"database/sql"
	"log"
)

// Global database connection
var dbConn *sql.DB

// SetDBConn assigns the database connection
func SetDBConn(conn *sql.DB) {
	dbConn = conn
	log.Println("Database connection initialized in db package.")
}

// GetDBConn returns the database connection
func GetDBConn() *sql.DB {
	return dbConn
}
