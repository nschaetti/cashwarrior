package db

import (
	"database/sql"
	"fmt"
	"time"
)

type Transaction struct {
	ID int64 // primary key

	Identifier  string    // Unique identifier for the transaction
	Type        string    // expense, income, transfer_out, transfer_in
	Amount      float64   // positive for income, negative for expense
	Description string    // free-form text
	Datetime    time.Time // date and time of the transaction

	AccountID   *int64 // foreign key to accounts
	AccountName string
	Currency    string
	CategoryID  *int64 // foreign key to categories
	PlaceID     *int64 // foreign key to places
	GroupID     *int64 // foreign key to groups

	Notes *string // free-form text

	CreatedAt time.Time // creation datetime of the DB entry
	UpdatedAt time.Time // last update datetime of the DB entry
}

type TransactionIDFilter struct {
	ID string
}

func (f TransactionIDFilter) GenerateSQL() (string, []any) {
	return "transactions.identifier = ?", []any{f.ID}
}

func (f TransactionIDFilter) String() string {
	return fmt.Sprintf("<TransactionIDFilter: %s>", f.ID)
}

type TransactionAccountNameFilter struct {
	Name string
}

func (f TransactionAccountNameFilter) GenerateSQL() (string, []any) {
	return "accounts.name = ?", []any{f.Name}
}

func (f TransactionAccountNameFilter) String() string {
	return fmt.Sprintf("<TransactionAccountNameFilter: %s>", f.Name)
}

type TransactionDatetimeFilter struct {
	From time.Time
	To   time.Time
}

func (f TransactionDatetimeFilter) GenerateSQL() (string, []any) {
	return "transactions.datetime BETWEEN ? AND ?", []any{f.From, f.To}
}

func (f TransactionDatetimeFilter) String() string {
	return fmt.Sprintf("<TransactionDatetimeFilter: %s - %s>", f.From, f.To)
}

type CreateTransactionInput struct {
	Identifier  string
	Type        string
	Amount      float64
	Description string
	Datetime    time.Time
	AccountID   int64
	CategoryID  *int64
	PlaceID     *int64
	GroupID     *int64
}

func GetLastTransaction(db DBTX) (Transaction, error) {
	var transaction Transaction
	return getTransactionFromQueryRow(db.QueryRow(`
SELECT transactions.id, transactions.identifier, transactions.type, transactions.amount, transactions.description, transactions.datetime,
       transactions.account_id, COALESCE(accounts.name, ''), COALESCE(accounts.currency, ''),
       transactions.category_id, transactions.place_id, transactions.group_id,
       transactions.created_at, transactions.updated_at
FROM transactions
LEFT JOIN accounts ON accounts.id = transactions.account_id
ORDER BY transactions.id DESC
LIMIT 1
`), transaction)
}

func GetTransactionByID(db DBTX, id int64) (Transaction, error) {
	var transaction Transaction
	return getTransactionFromQueryRow(db.QueryRow(`
SELECT transactions.id, transactions.identifier, transactions.type, transactions.amount, transactions.description, transactions.datetime,
       transactions.account_id, COALESCE(accounts.name, ''), COALESCE(accounts.currency, ''),
       transactions.category_id, transactions.place_id, transactions.group_id,
       transactions.created_at, transactions.updated_at
FROM transactions
LEFT JOIN accounts ON accounts.id = transactions.account_id
WHERE transactions.id = ?
`, id), transaction)
}

func GetTransactionByIdentifier(db DBTX, identifier string) (Transaction, error) {
	var transaction Transaction
	return getTransactionFromQueryRow(db.QueryRow(`
SELECT transactions.id, transactions.identifier, transactions.type, transactions.amount, transactions.description, transactions.datetime,
       transactions.account_id, COALESCE(accounts.name, ''), COALESCE(accounts.currency, ''),
       transactions.category_id, transactions.place_id, transactions.group_id,
       transactions.created_at, transactions.updated_at
FROM transactions
LEFT JOIN accounts ON accounts.id = transactions.account_id
WHERE transactions.identifier = ?
`, identifier), transaction)
}

func getTransactionFromQueryRow(row *sql.Row, transaction Transaction) (Transaction, error) {
	var accountID sql.NullInt64
	var categoryID sql.NullInt64
	var placeID sql.NullInt64
	var groupID sql.NullInt64

	err := row.Scan(
		&transaction.ID,
		&transaction.Identifier,
		&transaction.Type,
		&transaction.Amount,
		&transaction.Description,
		&transaction.Datetime,
		&accountID,
		&transaction.AccountName,
		&transaction.Currency,
		&categoryID,
		&placeID,
		&groupID,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)
	if err != nil {
		return transaction, err
	}

	if accountID.Valid {
		v := accountID.Int64
		transaction.AccountID = &v
	}

	if categoryID.Valid {
		transaction.CategoryID = &categoryID.Int64
	}

	if placeID.Valid {
		v := placeID.Int64
		transaction.PlaceID = &v
	}

	if groupID.Valid {
		v := groupID.Int64
		transaction.GroupID = &v
	}

	return transaction, nil
}

func GetSumOfTransactions(db DBTX, filters []SQLFilter) (float64, error) {
	query := `
SELECT SUM(amount)
FROM transactions
WHERE 1 = 1
`
	args := make([]interface{}, 0)

	for _, filter := range filters {
		filterSQL, filterArgs := filter.GenerateSQL()
		query += "AND " + filterSQL + "\n"
		args = append(args, filterArgs...)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	rows.Next()
	var sum float64
	err = rows.Scan(&sum)
	if err != nil {
		return 0, err
	}
	return sum, nil
}

func ListTransactions(
	db DBTX,
	dbFilters []SQLFilter,
	runFilters []Filter[Transaction],
) ([]Transaction, error) {
	query := `
SELECT transactions.id, transactions.identifier, transactions.type, transactions.amount, transactions.description, transactions.datetime,
       transactions.account_id, COALESCE(accounts.name, ''), COALESCE(accounts.currency, ''),
       transactions.category_id, transactions.place_id, transactions.group_id,
       transactions.created_at, transactions.updated_at
FROM transactions
LEFT JOIN accounts ON accounts.id = transactions.account_id
WHERE 1 = 1
`
	args := make([]interface{}, 0)

	for _, filter := range dbFilters {
		filterSQL, filterArgs := filter.GenerateSQL()
		query += "AND " + filterSQL + "\n"
		args = append(args, filterArgs...)
	}

	query += "ORDER BY transactions.id\n"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := make([]Transaction, 0)
	for rows.Next() {
		var transaction Transaction
		var accountID sql.NullInt64
		var accountName string
		var currency string
		var categoryID sql.NullInt64
		var placeID sql.NullInt64
		var groupID sql.NullInt64

		err = rows.Scan(
			&transaction.ID,
			&transaction.Identifier,
			&transaction.Type,
			&transaction.Amount,
			&transaction.Description,
			&transaction.Datetime,
			&accountID,
			&accountName,
			&currency,
			&categoryID,
			&placeID,
			&groupID,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if accountID.Valid {
			v := accountID.Int64
			transaction.AccountID = &v
		}
		transaction.AccountName = accountName
		transaction.Currency = currency
		if categoryID.Valid {
			v := categoryID.Int64
			transaction.CategoryID = &v
		}
		if placeID.Valid {
			v := placeID.Int64
			transaction.PlaceID = &v
		}
		if groupID.Valid {
			v := groupID.Int64
			transaction.GroupID = &v
		}

		transactions = append(transactions, transaction)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func TransactionExists(db DBTX, id int64) (bool, error) {
	var exists bool
	err := db.QueryRow(`
SELECT EXISTS(
	SELECT 1 FROM transactions WHERE id = ?
)
`, id).Scan(&exists)
	return exists, err
}

func InsertTransaction(db DBTX, input CreateTransactionInput) (int64, error) {
	transactionType := input.Type
	if transactionType == "" {
		transactionType = "expense"
	}

	result, err := db.Exec(`
INSERT INTO transactions (identifier, type, amount, description, datetime, account_id, category_id, place_id, group_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
`, input.Identifier, transactionType, input.Amount, input.Description, input.Datetime, input.AccountID, input.CategoryID, input.PlaceID, input.GroupID)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (t Transaction) GetAccount(db DBTX) (*Account, error) {
	if t.AccountID == nil {
		return nil, nil
	}

	account, err := GetAccountByID(db, *t.AccountID)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (t Transaction) GetCategory(db DBTX) (*Category, error) {
	if t.CategoryID == nil {
		return nil, nil
	}

	category, err := GetCategoryByID(db, *t.CategoryID)
	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (t Transaction) GetPlace(db DBTX) (*Place, error) {
	if t.PlaceID == nil {
		return nil, nil
	}

	place, err := GetPlaceByID(db, *t.PlaceID)
	if err != nil {
		return nil, err
	}

	return &place, nil
}

func (t Transaction) GetGroup(db DBTX) (*TransactionGroup, error) {
	if t.GroupID == nil {
		return nil, nil
	}

	group, err := GetTransactionGroupByID(db, *t.GroupID)
	if err != nil {
		return nil, err
	}

	return &group, nil
}
