package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB(connStr string) error {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}

	schema, _ := os.ReadFile("db/schema.sql")
	_, err = db.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("failed to initialize database schema: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	DB = db
	log.Println("Database connected successfully")
	return nil
}

func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
