package db

import (
	"database/sql"
	"time"
)

type TransferDBEntry struct {
	ID int64

	FromTransactionID int64
	ToTransactionID   int64

	FromAccountID int64
	ToAccountID   int64

	Amount float64

	CreatedAt time.Time
	UpdatedAt time.Time
}

type TransferDBListFilter struct {
	ID                *int64
	FromTransactionID *int64
	ToTransactionID   *int64
	FromAccountID     *int64
	ToAccountID       *int64
	Limit             int
	Offset            int
}

type CreateTransferDBInput struct {
	FromTransactionID int64
	ToTransactionID   int64
	FromAccountID     int64
	ToAccountID       int64
	Amount            float64
}

func GetTransferByID(db *sql.DB, id int64) (TransferDBEntry, error) {
	var transfer TransferDBEntry

	err := db.QueryRow(`
SELECT id, from_transaction_id, to_transaction_id, from_account_id, to_account_id, amount, created_at, updated_at
FROM transfers
WHERE id = ?
`, id).Scan(
		&transfer.ID,
		&transfer.FromTransactionID,
		&transfer.ToTransactionID,
		&transfer.FromAccountID,
		&transfer.ToAccountID,
		&transfer.Amount,
		&transfer.CreatedAt,
		&transfer.UpdatedAt,
	)
	if err != nil {
		return transfer, err
	}

	return transfer, nil
}

func ListTransfers(db *sql.DB, filter TransferDBListFilter) ([]TransferDBEntry, error) {
	query := `
SELECT id, from_transaction_id, to_transaction_id, from_account_id, to_account_id, amount, created_at, updated_at
FROM transfers
WHERE 1 = 1
`
	args := make([]interface{}, 0)

	if filter.ID != nil {
		query += "AND id = ?\n"
		args = append(args, *filter.ID)
	}
	if filter.FromTransactionID != nil {
		query += "AND from_transaction_id = ?\n"
		args = append(args, *filter.FromTransactionID)
	}
	if filter.ToTransactionID != nil {
		query += "AND to_transaction_id = ?\n"
		args = append(args, *filter.ToTransactionID)
	}
	if filter.FromAccountID != nil {
		query += "AND from_account_id = ?\n"
		args = append(args, *filter.FromAccountID)
	}
	if filter.ToAccountID != nil {
		query += "AND to_account_id = ?\n"
		args = append(args, *filter.ToAccountID)
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

	transfers := make([]TransferDBEntry, 0)
	for rows.Next() {
		var transfer TransferDBEntry
		err = rows.Scan(
			&transfer.ID,
			&transfer.FromTransactionID,
			&transfer.ToTransactionID,
			&transfer.FromAccountID,
			&transfer.ToAccountID,
			&transfer.Amount,
			&transfer.CreatedAt,
			&transfer.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, transfer)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return transfers, nil
}

func TransferExists(db *sql.DB, id int64) (bool, error) {
	var exists bool
	err := db.QueryRow(`
SELECT EXISTS(
	SELECT 1 FROM transfers WHERE id = ?
)
`, id).Scan(&exists)
	return exists, err
}

func InsertTransfer(db *sql.DB, input CreateTransferDBInput) (int64, error) {
	result, err := db.Exec(`
INSERT INTO transfers (from_transaction_id, to_transaction_id, from_account_id, to_account_id, amount)
VALUES (?, ?, ?, ?, ?)
`, input.FromTransactionID, input.ToTransactionID, input.FromAccountID, input.ToAccountID, input.Amount)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}
