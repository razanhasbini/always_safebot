package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func Connect(databaseURL string) *sql.DB {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal("Failed to open DB:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	log.Println(" Connected to PostgreSQL")
	return db
}
