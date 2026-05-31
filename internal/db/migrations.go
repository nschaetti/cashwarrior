package db

import (
	"database/sql"
	"fmt"

	"github.com/nschaetti/cashwarrior/internal/config"
)

type TableCreationFunc func(*sql.DB, config.Config) error

var Tables = map[string]TableCreationFunc{
	"accounts":           createAccountsTable,
	"budgets":            createBudgetsTable,
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
	initial_balance REAL NOT NULL DEFAULT 0.0,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`)
	if err != nil {
		return err
	}

	hasInitialBalance, err := accountHasInitialBalanceColumn(db)
	if err != nil {
		return err
	}
	if !hasInitialBalance {
		_, err = db.Exec(`
ALTER TABLE accounts ADD COLUMN initial_balance REAL NOT NULL DEFAULT 0.0
`)
		if err != nil {
			return err
		}
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

func accountHasInitialBalanceColumn(db *sql.DB) (bool, error) {
	rows, err := db.Query("PRAGMA table_info(accounts)")
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var dataType string
		var notNull int
		var defaultValue sql.NullString
		var pk int

		if err = rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return false, err
		}
		if name == "initial_balance" {
			return true, nil
		}
	}

	if err = rows.Err(); err != nil {
		return false, err
	}

	return false, nil
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

func createBudgetsTable(db *sql.DB, config config.Config) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS budgets (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	category_id INTEGER NOT NULL,
	name TEXT NOT NULL UNIQUE,
	description TEXT NOT NULL DEFAULT '',
	amount REAL NOT NULL,
	currency TEXT NOT NULL,
	period TEXT NOT NULL CHECK(period IN ('day', 'week', 'month', '2months', '3months', '4months', '6months', 'year')),
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (category_id) REFERENCES categories(id)
);
`)
	return err
}

func createCategoriesTable(db *sql.DB, config config.Config) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS categories (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	parent_id INTEGER,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (parent_id) REFERENCES categories(id)
);
`)
	if err != nil {
		return err
	}

	hasParentID, err := categoryHasParentIDColumn(db)
	if err != nil {
		return err
	}
	if !hasParentID {
		_, err = db.Exec(`
ALTER TABLE categories ADD COLUMN parent_id INTEGER
`)
		if err != nil {
			return err
		}
	}

	rootID, err := ensureRootCategory(db)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
UPDATE categories
SET parent_id = ?
WHERE parent_id IS NULL AND name <> ?
`, rootID, rootCategoryName)
	return err
}

func categoryHasParentIDColumn(db *sql.DB) (bool, error) {
	rows, err := db.Query("PRAGMA table_info(categories)")
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var dataType string
		var notNull int
		var defaultValue sql.NullString
		var pk int

		err = rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			return false, err
		}
		if name == "parent_id" {
			return true, nil
		}
	}

	if err = rows.Err(); err != nil {
		return false, err
	}

	return false, nil
}

func ensureRootCategory(db *sql.DB) (int64, error) {
	exists, err := CategoryExists(db, rootCategoryName)
	if err != nil {
		return 0, err
	}
	if !exists {
		_, err = db.Exec(`
INSERT INTO categories (name, parent_id)
VALUES (?, NULL)
`, rootCategoryName)
		if err != nil {
			return 0, err
		}
	}

	rootID, err := getCategoryIDByName(db, rootCategoryName)
	if err != nil {
		return 0, err
	}

	_, err = db.Exec(`
UPDATE categories
SET parent_id = NULL
WHERE id = ?
`, rootID)
	if err != nil {
		return 0, fmt.Errorf("reset root category parent: %w", err)
	}

	return rootID, nil
}

func createTransactionsTable(db *sql.DB, config config.Config) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS transactions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	identifier TEXT NOT NULL UNIQUE,
	type TEXT NOT NULL DEFAULT 'expense' CHECK(type IN ('expense', 'income', 'transfer_out', 'transfer_in')),
	amount REAL NOT NULL,
	description TEXT NOT NULL,
	datetime DATETIME NOT NULL,
	account_id INTEGER,
	category_id INTEGER,
	place_id INTEGER NOT NULL,
	group_id INTEGER,
	deleted BOOLEAN NOT NULL DEFAULT FALSE,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (account_id) REFERENCES accounts(id),
	FOREIGN KEY (category_id) REFERENCES categories(id),
	FOREIGN KEY (place_id) REFERENCES places(id),
	FOREIGN KEY (group_id) REFERENCES transaction_groups(id)
);
`)
	if err != nil {
		return err
	}

	hasDeleted, err := transactionHasDeletedColumn(db)
	if err != nil {
		return err
	}
	if hasDeleted {
		return nil
	}

	_, err = db.Exec(`
ALTER TABLE transactions ADD COLUMN deleted BOOLEAN NOT NULL DEFAULT FALSE
`)
	return err
}

func transactionHasDeletedColumn(db *sql.DB) (bool, error) {
	rows, err := db.Query("PRAGMA table_info(transactions)")
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var dataType string
		var notNull int
		var defaultValue sql.NullString
		var pk int

		if err = rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return false, err
		}
		if name == "deleted" {
			return true, nil
		}
	}

	if err = rows.Err(); err != nil {
		return false, err
	}

	return false, nil
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
	deleted BOOLEAN NOT NULL DEFAULT FALSE,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (from_transaction_id) REFERENCES transactions(id) ON DELETE CASCADE,
	FOREIGN KEY (to_transaction_id) REFERENCES transactions(id) ON DELETE CASCADE,
	FOREIGN KEY (from_account_id) REFERENCES accounts(id),
	FOREIGN KEY (to_account_id) REFERENCES accounts(id)
);
`)
	if err != nil {
		return err
	}

	hasDeleted, err := transferHasDeletedColumn(db)
	if err != nil {
		return err
	}
	if hasDeleted {
		return nil
	}

	_, err = db.Exec(`
ALTER TABLE transfers ADD COLUMN deleted BOOLEAN NOT NULL DEFAULT FALSE
`)
	return err
}

func transferHasDeletedColumn(db *sql.DB) (bool, error) {
	rows, err := db.Query("PRAGMA table_info(transfers)")
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var dataType string
		var notNull int
		var defaultValue sql.NullString
		var pk int

		if err = rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return false, err
		}
		if name == "deleted" {
			return true, nil
		}
	}

	if err = rows.Err(); err != nil {
		return false, err
	}

	return false, nil
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
	if err != nil {
		return err
	}

	exists, err := PlaceExists(db, "transfer")
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = InsertPlace(db, CreatePlaceInput{Name: "transfer"})
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
		"budgets",
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
