package db

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/nschaetti/cashwarrior/internal/config"
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

func Open(config config.Config) (*sql.DB, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(config.Database)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Open the database
	db, err := sql.Open("sqlite", config.Database)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// If not initialized, initialize it
	err = Init(db, config)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func Close(db *sql.DB) error {
	return db.Close()
}
