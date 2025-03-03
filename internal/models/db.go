package models

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // Postgres driver
)

var dbConn *sql.DB

func OpenDB(host, port, user, password, dbname string) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)
	return sql.Open("postgres", dsn)
}

func SetDBConn(conn *sql.DB) {
	dbConn = conn
	log.Println("Database connection initialized.")
}

func GetDBConn() *sql.DB {
	return dbConn
}
