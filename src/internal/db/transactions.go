package db

import (
	"database/sql"
	"time"
)

type TransactionDBEntry struct {
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

/*
type TransactionDBListFilter struct {
	ID *int64

	AccountID  *int64
	CategoryID *int64
	PlaceID    *int64
	GroupID    *int64

	AmountFrom *float64
	AmountTo   *float64

	DateFrom time.Time // DateFrom <= datetime <= DateTo
	DateTo   time.Time // DateFrom <= datetime <= DateTo

	Limit  int
	Offset int
}*/

type CreateTransactionDBInput struct {
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

func GetTransactionByID(db *sql.DB, id int64) (TransactionDBEntry, error) {
	var transaction TransactionDBEntry
	var accountID sql.NullInt64
	var categoryID sql.NullInt64
	var placeID sql.NullInt64
	var groupID sql.NullInt64

	err := db.QueryRow(`
SELECT id, identifier, type, amount, description, datetime, account_id, category_id, place_id, group_id, created_at, updated_at
FROM transactions
WHERE id = ?
`, id).Scan(
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

	return transaction, nil
}

func ListTransactions(db *sql.DB, filters []Filter) ([]TransactionDBEntry, error) {
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

	transactions := make([]TransactionDBEntry, 0)
	for rows.Next() {
		var transaction TransactionDBEntry
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

func InsertTransaction(db *sql.DB, input CreateTransactionDBInput) (int64, error) {
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
