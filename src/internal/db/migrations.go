package db

import (
	"database/sql"

	"github.com/nschaetti/cashwarrior/internal/config"
)

type TableCreationFunc func(*sql.DB, config.Config) error

var Tables = map[string]TableCreationFunc{
	"accounts":           createAccountsTable,
	"categories":         createCategoriesTable,
	"tags":               createTagsTable,
	"transaction_groups": createTransactionGroupsTable,
	"places":             createPlacesTable,
	"transactions":       createTransactionsTable,
	"transfers":          createTransfersTable,
	"transaction_tags":   createTransactionTagsTable,
}

func createAccountsTable(db *sql.DB, config config.Config) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS accounts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	currency TEXT NOT NULL DEFAULT 'CHF',
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`)
	if err != nil {
		return err
	}

	exists, err := AccountExists(db, config.Default.Account)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = InsertAccount(db, CreateAccountInput{
		Name:     config.Default.Account,
		Currency: config.Default.Currency,
	})
	return err
}

func createTagsTable(db *sql.DB, config config.Config) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS tags (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`)
	return err
}

func createCategoriesTable(db *sql.DB, config config.Config) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS categories (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`)
	return err
}

func createTransactionsTable(db *sql.DB, config config.Config) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS transactions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	type TEXT NOT NULL DEFAULT 'expense' CHECK(type IN ('expense', 'income', 'transfer_out', 'transfer_in')),
	amount REAL NOT NULL,
	description TEXT NOT NULL,
	datetime DATETIME NOT NULL,
	account_id INTEGER,
	category_id INTEGER,
	place_id INTEGER NOT NULL,
	group_id INTEGER,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (account_id) REFERENCES accounts(id),
	FOREIGN KEY (category_id) REFERENCES categories(id),
	FOREIGN KEY (place_id) REFERENCES places(id),
	FOREIGN KEY (group_id) REFERENCES transaction_groups(id)
);
`)
	return err
}

func createTransactionGroupsTable(db *sql.DB, config config.Config) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS transaction_groups (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`)
	return err
}

func createTransfersTable(db *sql.DB, config config.Config) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS transfers (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	from_transaction_id INTEGER NOT NULL,
	to_transaction_id INTEGER NOT NULL,
	from_account_id INTEGER NOT NULL,
	to_account_id INTEGER NOT NULL,
	amount REAL NOT NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (from_transaction_id) REFERENCES transactions(id) ON DELETE CASCADE,
	FOREIGN KEY (to_transaction_id) REFERENCES transactions(id) ON DELETE CASCADE,
	FOREIGN KEY (from_account_id) REFERENCES accounts(id),
	FOREIGN KEY (to_account_id) REFERENCES accounts(id)
);
`)
	return err
}

func createPlacesTable(db *sql.DB, config config.Config) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS places (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`)
	return err
}

func createTransactionTagsTable(db *sql.DB, config config.Config) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS transaction_tags (
	transaction_id INTEGER NOT NULL,
	tag_id INTEGER NOT NULL,
	PRIMARY KEY (transaction_id, tag_id),
	FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE CASCADE,
	FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);
`)
	return err
}

func Init(db *sql.DB, config config.Config) error {
	_, err := db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return err
	}

	// Create tables
	orderedTables := []string{
		"accounts",
		"categories",
		"tags",
		"transaction_groups",
		"places",
		"transactions",
		"transfers",
		"transaction_tags",
	}

	for _, tableName := range orderedTables {
		fn := Tables[tableName]
		err := fn(db, config)
		if err != nil {
			return err
		}
	}
	return nil
}
