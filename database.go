package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

// Global database connection (used by message, notification, and report handlers)
var db *sql.DB

// InitDatabase initializes the global database connection
func InitDatabase(connStr string) error {
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		return err
	}

	log.Println("✅ Global database connection established")
	return nil
}
