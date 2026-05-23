package db

import (
	"database/sql"
)

type TableCreationFunc func(db *sql.DB) error

var Tables = map[string]TableCreationFunc{
	"entries": createEntriesTable,
	"tags":    createTagsTable,
}

func createEntriesTable(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS entries (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	public_id TEXT UNIQUE NOT NULL,
	date TEXT NOT NULL,
	amount_cents INTEGER NOT NULL,
	currency TEXT NOT NULL,
	account TEXT,
	category TEXT,
	description TEXT,
	status TEXT NOT NULL DEFAULT 'active',
	transfer_id TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
`)
	return err
}

func createTagsTable(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS entry_tags (
	entry_id INTEGER NOT NULL,
	tag TEXT NOT NULL,
	PRIMARY KEY (entry_id, tag),
	FOREIGN KEY (entry_id) REFERENCES entries(id)
);
`)
	return err
}

func Init(db *sql.DB) error {
	// Create tables
	for _, fn := range Tables {
		err := fn(db)
		if err != nil {
			return err
		}
	}
	return nil
}
