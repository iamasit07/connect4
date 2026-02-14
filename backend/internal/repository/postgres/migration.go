package postgres

import (
	"database/sql"
	"fmt"
	"os"
)

// RunMigrations executes the schema.sql file to initialize the database
func RunMigrations(db *sql.DB) error {
	// We check a few common locations for the schema file to make this robust
	// regardless of where 'go run' or the binary is executed from.
	possiblePaths := []string{
		"script/migration/schema.sql",          // From backend root (go run ./cmd/api)
		"../script/migration/schema.sql",       // From cmd/api
		"../../script/migration/schema.sql",    // From internal/...
		"backend/script/migration/schema.sql",  // From repo root
	}

	var schemaPath string
	var content []byte
	var err error

	// Try to find the file
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			schemaPath = path
			break
		}
	}

	// If relative paths failed, try to construct absolute path based on runtime/executable?
	// For now, if schemaPath is empty, we default to the standard expectation and let it error out with context.
	if schemaPath == "" {
		schemaPath = "script/migration/schema.sql"
	}

	// Read schema file
	content, err = os.ReadFile(schemaPath)
	if err != nil {
		wd, _ := os.Getwd()
		return fmt.Errorf("failed to read migration file. Looking for '%s' (Current WD: %s). Error: %v", schemaPath, wd, err)
	}

	// Execute SQL
	// Note: We use string(content) because Exec requires a string query
	if _, err := db.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute schema.sql: %v", err)
	}

	return nil
}