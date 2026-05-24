package db

import (
	"database/sql"
	"time"
)

type Account struct {
	ID        int64
	Name      string
	Currency  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AccountListFilter struct {
	NameLike string
	Limit    int
	Offset   int
}

type CreateAccountInput struct {
	Name     string
	Currency string
}

func GetAccountByID(db *sql.DB, id int64) (Account, error) {
	var account Account
	err := db.QueryRow(`
SELECT id, name, currency, created_at, updated_at
FROM accounts
WHERE id = ?
`, id).Scan(
		&account.ID,
		&account.Name,
		&account.Currency,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	return account, err
}

func ListAccounts(db *sql.DB, filter AccountListFilter) ([]Account, error) {
	query := `
SELECT id, name, currency, created_at, updated_at
FROM accounts
`
	args := make([]interface{}, 0)

	if filter.NameLike != "" {
		query += "WHERE name LIKE ?\n"
		args = append(args, "%"+filter.NameLike+"%")
	}

	query += "ORDER BY id\n"
	if filter.Limit > 0 {
		query += "LIMIT ?\n"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += "OFFSET ?\n"
		args = append(args, filter.Offset)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := make([]Account, 0)
	for rows.Next() {
		var account Account
		err = rows.Scan(
			&account.ID,
			&account.Name,
			&account.Currency,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

func AccountExists(db *sql.DB, name string) (bool, error) {
	var exists bool
	err := db.QueryRow(`
SELECT EXISTS(
	SELECT 1 FROM accounts WHERE name = ?
)
`, name).Scan(&exists)
	return exists, err
}

func InsertAccount(db *sql.DB, input CreateAccountInput) (int64, error) {
	result, err := db.Exec(`
INSERT INTO accounts (name, currency)
VALUES (?, ?)
`, input.Name, input.Currency)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}
