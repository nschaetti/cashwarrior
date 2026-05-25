package db

import (
	"database/sql"
	"time"
)

type Transaction struct {
	ID int64 // primary key

	Identifier  string    // Unique identifier for the transaction
	Type        string    // expense, income, transfer_out, transfer_in
	Amount      float64   // positive for income, negative for expense
	Description string    // free-form text
	Datetime    time.Time // date and time of the transaction

	AccountID  *int64 // foreign key to accounts
	CategoryID *int64 // foreign key to categories
	PlaceID    *int64 // foreign key to places
	GroupID    *int64 // foreign key to groups

	Notes *string // free-form text

	CreatedAt time.Time // creation datetime of the DB entry
	UpdatedAt time.Time // last update datetime of the DB entry
}

type TransactionIDFilter struct {
	ID int64
}

func (f TransactionIDFilter) GenerateSQL() (string, []any) {
	return "id = ?", []any{f.ID}
}

type CreateTransactionInput struct {
	Identifier  string
	Type        string
	Amount      float64
	Description string
	Datetime    time.Time
	AccountID   *int64
	CategoryID  *int64
	PlaceID     int64
	GroupID     *int64
}

func GetLastTransaction(db *sql.DB) (Transaction, error) {
	var transaction Transaction
	return getTransactionFromQueryRow(db.QueryRow(`
SELECT id, identifier, type, amount, description, datetime, account_id, category_id, place_id, group_id, created_at, updated_at
FROM transactions
ORDER BY id DESC
LIMIT 1
`), transaction)
}

func GetTransactionByID(db *sql.DB, id int64) (Transaction, error) {
	var transaction Transaction
	return getTransactionFromQueryRow(db.QueryRow(`
SELECT id, identifier, type, amount, description, datetime, account_id, category_id, place_id, group_id, created_at, updated_at
FROM transactions
WHERE id = ?
`, id), transaction)
}

func GetTransactionByIdentifier(db *sql.DB, identifier string) (Transaction, error) {
	var transaction Transaction
	return getTransactionFromQueryRow(db.QueryRow(`
SELECT id, identifier, type, amount, description, datetime, account_id, category_id, place_id, group_id, created_at, updated_at
FROM transactions
WHERE identifier = ?
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

func ListTransactions(db *sql.DB, filters []Filter) ([]Transaction, error) {
	query := `
SELECT id, identifier, type, amount, description, datetime, account_id, category_id, place_id, group_id, created_at, updated_at
FROM transactions
WHERE 1 = 1
`
	args := make([]interface{}, 0)

	for _, filter := range filters {
		filterSQL, filterArgs := filter.GenerateSQL()
		query += "AND " + filterSQL + "\n"
		args = append(args, filterArgs...)
	}

	query += "ORDER BY id\n"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := make([]Transaction, 0)
	for rows.Next() {
		var transaction Transaction
		var accountID sql.NullInt64
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

func TransactionExists(db *sql.DB, id int64) (bool, error) {
	var exists bool
	err := db.QueryRow(`
SELECT EXISTS(
	SELECT 1 FROM transactions WHERE id = ?
)
`, id).Scan(&exists)
	return exists, err
}

func InsertTransaction(db *sql.DB, input CreateTransactionInput) (int64, error) {
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

func (t Transaction) GetAccount(db *sql.DB) (*Account, error) {
	if t.AccountID == nil {
		return nil, nil
	}

	account, err := GetAccountByID(db, *t.AccountID)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (t Transaction) GetCategory(db *sql.DB) (*Category, error) {
	if t.CategoryID == nil {
		return nil, nil
	}

	category, err := GetCategoryByID(db, *t.CategoryID)
	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (t Transaction) GetPlace(db *sql.DB) (*Place, error) {
	if t.PlaceID == nil {
		return nil, nil
	}

	place, err := GetPlaceByID(db, *t.PlaceID)
	if err != nil {
		return nil, err
	}

	return &place, nil
}

func (t Transaction) GetGroup(db *sql.DB) (*TransactionGroup, error) {
	if t.GroupID == nil {
		return nil, nil
	}

	group, err := GetTransactionGroupByID(db, *t.GroupID)
	if err != nil {
		return nil, err
	}

	return &group, nil
}
