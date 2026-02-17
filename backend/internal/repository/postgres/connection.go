package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB

func InitDB(connStr string, maxOpenConns, maxIdleConns, connMaxLifetimeMin int) error {
	db, err := sql.Open("pgx", connStr)
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

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(time.Duration(connMaxLifetimeMin) * time.Minute)

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
