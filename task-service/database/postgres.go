package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func InitDB() *sql.DB {
    connStr := "postgresql://postgres:4840707101@localhost:8000/gosprints?sslmode=disable"

    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal("Error opening database:", err)
    }

    if err = db.Ping(); err != nil {
        log.Fatal("Error connecting to database:", err)
    }
    fmt.Println("Connected to PostgreSQL!")

	return db
}