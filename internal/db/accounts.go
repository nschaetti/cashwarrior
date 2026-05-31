package db

import "time"

type Account struct {
	ID             int64
	Name           string
	Currency       string
	InitialBalance float64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type AccountListFilter struct {
	NameLike string
	Limit    int
	Offset   int
}

type CreateAccountInput struct {
	Name           string
	Currency       string
	InitialBalance float64
}

func GetAccountByID(db DBTX, id int64) (Account, error) {
	var account Account
	err := db.QueryRow(`
SELECT id, name, currency, initial_balance, created_at, updated_at
FROM accounts
WHERE id = ?
`, id).Scan(
		&account.ID,
		&account.Name,
		&account.Currency,
		&account.InitialBalance,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	return account, err
}

func GetAccountByName(db DBTX, name string) (Account, error) {
	var account Account
	err := db.QueryRow(`
SELECT id, name, currency, initial_balance, created_at, updated_at
FROM accounts
WHERE name = ?
`, name).Scan(
		&account.ID,
		&account.Name,
		&account.Currency,
		&account.InitialBalance,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	return account, err
}

func ListAccounts(db DBTX, filter AccountListFilter) ([]Account, error) {
	query := `
SELECT id, name, currency, initial_balance, created_at, updated_at
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
			&account.InitialBalance,
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

func AccountExists(db DBTX, name string) (bool, error) {
	var exists bool
	err := db.QueryRow(`
SELECT EXISTS(
	SELECT 1 FROM accounts WHERE name = ?
)
`, name).Scan(&exists)
	return exists, err
}

func InsertAccount(db DBTX, input CreateAccountInput) (int64, error) {
	result, err := db.Exec(`
INSERT INTO accounts (name, currency, initial_balance)
VALUES (?, ?, ?)
`, input.Name, input.Currency, input.InitialBalance)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func UpdateAccountName(db DBTX, accountID int64, name string) error {
	_, err := db.Exec(`
UPDATE accounts
SET name = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, name, accountID)
	return err
}

func UpdateAccountCurrency(db DBTX, accountID int64, currency string) error {
	_, err := db.Exec(`
UPDATE accounts
SET currency = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, currency, accountID)
	return err
}

func UpdateAccountInitialBalance(db DBTX, accountID int64, initialBalance float64) error {
	_, err := db.Exec(`
UPDATE accounts
SET initial_balance = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, initialBalance, accountID)
	return err
}

func DeleteAccountByID(db DBTX, accountID int64) error {
	_, err := db.Exec(`
DELETE FROM accounts
WHERE id = ?
`, accountID)
	return err
}

func CountTransactionsByAccountID(db DBTX, accountID int64) (int, error) {
	var count int
	err := db.QueryRow(`
SELECT COUNT(*)
FROM transactions
WHERE account_id = ?
`, accountID).Scan(&count)
	return count, err
}

func ListAccountCurrencies(db DBTX) ([]string, error) {
	rows, err := db.Query(`
SELECT DISTINCT currency
FROM accounts
ORDER BY currency
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	currencies := make([]string, 0)
	for rows.Next() {
		var currency string
		err = rows.Scan(&currency)
		if err != nil {
			return nil, err
		}
		currencies = append(currencies, currency)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return currencies, nil
}
