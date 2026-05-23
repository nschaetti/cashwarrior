package db

import (
	"database/sql"
	"os"
	"path/filepath"
)

func checkTableExists(db *sql.DB, tableName string) (bool, error) {
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name=?", tableName)
	if err != nil {
		return false, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)
	return true, nil
}

func checkDatabaseTables(db *sql.DB) error {
	// Check tables exist
	exists, err := checkTableExists(db, "entries")
	if err != nil {
		return err
	}
	if !exists {
		return createEntriesTable(db)
	}
	return nil
}

func Open(dbPath string) (*sql.DB, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Open the database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// If not initialized, initialize it
	err = Init(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func Close(db *sql.DB) error {
	return db.Close()
}
