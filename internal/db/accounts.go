package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Account represents an account in the database.
type Account struct {
	ID             int64
	Name           string
	Currency       string
	InitialBalance float64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// CreateAccountInput is used to create a new account.
type CreateAccountInput struct {
	Name           string
	Currency       string
	InitialBalance float64
}

/*
 * Account ID
 */

// AccountIDFilter is used to filter the list of accounts.
type AccountIDFilter struct {
	ID int64
}

func (f AccountIDFilter) GenerateSQL() (string, []any) {
	return "id = ?", []any{f.ID}
}
func (f AccountIDFilter) String() string {
	return fmt.Sprintf("<AccountIDFilter: %d>", f.ID)
}

// AccountIDListFilter is used to filter the list of accounts.
type AccountIDListFilter struct {
	IDs []int64
}

func (f AccountIDListFilter) GenerateSQL() (string, []any) {
	return "id IN (?)", []any{f.IDs}
}
func (f AccountIDListFilter) String() string {
	return fmt.Sprintf("<AccountIDListFilter: %d>", f.IDs)
}

// AccountIDRangeFilter is used to filter the list of accounts.
type AccountIDRangeFilter struct {
	From int64
	To   int64
}

func (f AccountIDRangeFilter) GenerateSQL() (string, []any) {
	return "id BETWEEN ? AND ?", []any{f.From, f.To}
}
func (f AccountIDRangeFilter) String() string {
	return fmt.Sprintf("<AccountIDRangeFilter: %d-%d>", f.From, f.To)
}

// AccountIDGreaterThanFilter is used to filter the list of accounts.
type AccountIDGreaterThanFilter struct {
	ID int64
}

func (f AccountIDGreaterThanFilter) GenerateSQL() (string, []any) {
	return "id > ?", []any{f.ID}
}
func (f AccountIDGreaterThanFilter) String() string {
	return fmt.Sprintf("<AccountIDGreaterThanFilter: %d>", f.ID)
}

// AccountIDLessThanFilter is used to filter the list of accounts.
type AccountIDLessThanFilter struct {
	ID int64
}

func (f AccountIDLessThanFilter) GenerateSQL() (string, []any) {
	return "id < ?", []any{f.ID}
}
func (f AccountIDLessThanFilter) String() string {
	return fmt.Sprintf("<AccountIDLessThanFilter: %d>", f.ID)
}

// AccountIDGreaterThanOrEqualFilter is used to filter the list of accounts.
type AccountIDGreaterThanOrEqualFilter struct {
	ID int64
}

func (f AccountIDGreaterThanOrEqualFilter) GenerateSQL() (string, []any) {
	return "id >= ?", []any{f.ID}
}
func (f AccountIDGreaterThanOrEqualFilter) String() string {
	return fmt.Sprintf("<AccountIDGreaterThanOrEqualFilter: %d>", f.ID)
}

// AccountIDLessThanOrEqualFilter is used to filter the list of accounts.
type AccountIDLessThanOrEqualFilter struct {
	ID int64
}

func (f AccountIDLessThanOrEqualFilter) GenerateSQL() (string, []any) {
	return "id <= ?", []any{f.ID}
}
func (f AccountIDLessThanOrEqualFilter) String() string {
	return fmt.Sprintf("<AccountIDLessThanOrEqualFilter: %d>", f.ID)
}

/*
 * Account name filters
 */

// AccountNameLikeFilter is used to filter the list of accounts.
type AccountNameLikeFilter struct {
	NameLike string
	Limit    int
	Offset   int
}

func (f AccountNameLikeFilter) String() string {
	return fmt.Sprintf("<AccountNameLikeFilter: %s>", f.NameLike)
}
func (f AccountNameLikeFilter) IsEmpty() bool { return f.NameLike == "" }
func (f AccountNameLikeFilter) GenerateSQL() (string, []any) {
	return "name LIKE ?", []any{f.NameLike}
}

// AccountAccountNameFilter is used to filter the list of accounts.
type AccountAccountNameFilter struct {
	Name string
}

func (f AccountAccountNameFilter) GenerateSQL() (string, []any) {
	return "name = ?", []any{f.Name}
}
func (f AccountAccountNameFilter) String() string {
	return fmt.Sprintf("<AccountAccountNameFilter: %s>", f.Name)
}

// AccountNameListFilter is used to filter the list of accounts.
type AccountNameListFilter struct {
	Names []string
}

func (f AccountNameListFilter) GenerateSQL() (string, []any) {
	return "name IN (?)", []any{f.Names}
}
func (f AccountNameListFilter) String() string {
	return fmt.Sprintf("<AccountNameListFilter: %s>", f.Names)
}

/*
 * Currency filters
 */

// AccountCurrencyFilter is used to filter an account on a specific currency.
type AccountCurrencyFilter struct {
	Currency string
}

func (f AccountCurrencyFilter) GenerateSQL() (string, []any) {
	return "currency = ?", []any{f.Currency}
}
func (f AccountCurrencyFilter) String() string {
	return fmt.Sprintf("<AccountCurrencyFilter: %s>", f.Currency)
}

// AccountCurrencyListFilter is used to filter account on a specific currency.
type AccountCurrencyListFilter struct {
	Currencies []string
}

func (f AccountCurrencyListFilter) GenerateSQL() (string, []any) {
	return "currency IN (?)", []any{f.Currencies}
}
func (f AccountCurrencyListFilter) String() string {
	return fmt.Sprintf("<AccountCurrencyListFilter: %s>", f.Currencies)
}

type AccountCurrencyLikeFilter struct {
	CurrencyLike string
}

func (f AccountCurrencyLikeFilter) GenerateSQL() (string, []any) {
	return "currency LIKE ?", []any{f.CurrencyLike}
}
func (f AccountCurrencyLikeFilter) String() string {
	return fmt.Sprintf("<AccountCurrencyLikeFilter: %s>", f.CurrencyLike)
}

/*
 * Initial balance filters
 */

type AccountInitialBalanceFilter struct {
	InitialBalance float64
}

func (f AccountInitialBalanceFilter) GenerateSQL() (string, []any) {
	return "initial_balance = ?", []any{f.InitialBalance}
}
func (f AccountInitialBalanceFilter) String() string {
	return fmt.Sprintf("<AccountInitialBalanceFilter: %f>", f.InitialBalance)
}

type AccountInitialBalanceRangeFilter struct {
	From float64
	To   float64
}

func (f AccountInitialBalanceRangeFilter) GenerateSQL() (string, []any) {
	return "initial_balance BETWEEN ? AND ?", []any{f.From, f.To}
}
func (f AccountInitialBalanceRangeFilter) String() string {
	return fmt.Sprintf("<AccountInitialBalanceRangeFilter: %f-%f>", f.From, f.To)
}

type AccountInitialBalanceGreaterThanFilter struct {
	InitialBalance float64
}

func (f AccountInitialBalanceGreaterThanFilter) GenerateSQL() (string, []any) {
	return "initial_balance > ?", []any{f.InitialBalance}
}
func (f AccountInitialBalanceGreaterThanFilter) String() string {
	return fmt.Sprintf("<AccountInitialBalanceGreaterThanFilter: %f>", f.InitialBalance)
}

type AccountInitialBalanceLessThanFilter struct {
	InitialBalance float64
}

func (f AccountInitialBalanceLessThanFilter) GenerateSQL() (string, []any) {
	return "initial_balance < ?", []any{f.InitialBalance}
}
func (f AccountInitialBalanceLessThanFilter) String() string {
	return fmt.Sprintf("<AccountInitialBalanceLessThanFilter: %f>", f.InitialBalance)
}

type AccountInitialBalanceGreaterThanOrEqualFilter struct {
	InitialBalance float64
}

func (f AccountInitialBalanceGreaterThanOrEqualFilter) GenerateSQL() (string, []any) {
	return "initial_balance >= ?", []any{f.InitialBalance}
}
func (f AccountInitialBalanceGreaterThanOrEqualFilter) String() string {
	return fmt.Sprintf("<AccountInitialBalanceGreaterThanOrEqualFilter: %f>", f.InitialBalance)
}

type AccountInitialBalanceLessThanOrEqualFilter struct {
	InitialBalance float64
}

func (f AccountInitialBalanceLessThanOrEqualFilter) GenerateSQL() (string, []any) {
	return "initial_balance <= ?", []any{f.InitialBalance}
}
func (f AccountInitialBalanceLessThanOrEqualFilter) String() string {
	return fmt.Sprintf("<AccountInitialBalanceLessThanOrEqualFilter: %f>", f.InitialBalance)
}

/*
 * Functions
 */

// GetAccountByID returns an account by its ID.
// If the account does not exist, it returns an error.
// If the account exists, it returns the account.
// db is the database connection.
// id is the ID of the account to retrieve.
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

// GetAccountByName returns an account by its name.
// If the account does not exist, it returns an error.
// If the account exists, it returns the account.
// db is the database connection.
// name is the name of the account to retrieve.
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

// ListAccounts returns a list of accounts.
// db is the database connection.
// dbFilters is a list of filters to apply to the query.
func ListAccounts(
	db DBTX,
	dbFilters []SQLFilter,
	runFilters []Filter[Account],
	orderBy []string,
) ([]Account, error) {
	query := `
SELECT id, name, currency, initial_balance, created_at, updated_at
FROM accounts
`
	args := make([]interface{}, 0, 1)

	// Add filters
	for i, filter := range dbFilters {
		filterSQL, filterArgs := filter.GenerateSQL()
		if i == 0 {
			query += "WHERE " + filterSQL + "\n"
			args = append(args, filterArgs...)
			continue
		}
		query += "AND " + filterSQL + "\n"
		args = append(args, filterArgs...)
	}

	// Order by id
	if len(orderBy) > 0 {
		query += "ORDER BY "
		for i, orderByField := range orderBy {
			query += orderByField
			if i < len(orderBy)-1 {
				query += ", "
			}
		}
	} else {
		query += "ORDER BY id\n"
	}

	// Query
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
		if err != nil {
			panic(fmt.Sprintf("Error closing rows: %s", err))
		}
	}(rows)

	// Transform in Account objects
	accounts := make([]Account, 0)
	var addItem bool
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
		addItem = true
		for _, filter := range runFilters {
			if filter.Reject(account) {
				addItem = false
				break
			}
		}
		if addItem {
			accounts = append(accounts, account)
		}
	}

	// Check for errors
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
