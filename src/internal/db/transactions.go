package db

import (
	"database/sql"
	"time"
)

type TransactionDBEntry struct {
	ID int64 // primary key

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
}

type CreateTransactionDBInput struct {
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
SELECT id, type, amount, description, datetime, account_id, category_id, place_id, group_id, created_at, updated_at
FROM transactions
WHERE id = ?
`, id).Scan(
		&transaction.ID,
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

func ListTransactions(db *sql.DB, filter TransactionDBListFilter) ([]TransactionDBEntry, error) {
	query := `
SELECT id, type, amount, description, datetime, account_id, category_id, place_id, group_id, created_at, updated_at
FROM transactions
WHERE 1 = 1
`
	args := make([]interface{}, 0)

	if filter.AccountID != nil {
		query += "AND account_id = ?\n"
		args = append(args, *filter.AccountID)
	}
	if filter.CategoryID != nil {
		query += "AND category_id = ?\n"
		args = append(args, *filter.CategoryID)
	}
	if filter.PlaceID != nil {
		query += "AND place_id = ?\n"
		args = append(args, *filter.PlaceID)
	}
	if filter.GroupID != nil {
		query += "AND group_id = ?\n"
		args = append(args, *filter.GroupID)
	}
	if !filter.DateFrom.IsZero() {
		query += "AND datetime >= ?\n"
		args = append(args, filter.DateFrom)
	}
	if !filter.DateTo.IsZero() {
		query += "AND datetime <= ?\n"
		args = append(args, filter.DateTo)
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

	transactions := make([]TransactionDBEntry, 0)
	for rows.Next() {
		var transaction TransactionDBEntry
		var accountID sql.NullInt64
		var categoryID sql.NullInt64
		var placeID sql.NullInt64
		var groupID sql.NullInt64

		err = rows.Scan(
			&transaction.ID,
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
INSERT INTO transactions (type, amount, description, datetime, account_id, category_id, place_id, group_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`, transactionType, input.Amount, input.Description, input.Datetime, input.AccountID, input.CategoryID, input.PlaceID, input.GroupID)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}
