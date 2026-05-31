package db

import "time"

type Transfer struct {
	ID int64

	FromTransactionID int64
	ToTransactionID   int64

	FromAccountID int64
	ToAccountID   int64

	Amount  float64
	Deleted bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

type TransferListFilter struct {
	ID                *int64
	FromTransactionID *int64
	ToTransactionID   *int64
	FromAccountID     *int64
	ToAccountID       *int64
	Limit             int
	Offset            int
}

type CreateTransferInput struct {
	FromTransactionID int64
	ToTransactionID   int64
	FromAccountID     int64
	ToAccountID       int64
	Amount            float64
}

func GetTransferByID(db DBTX, id int64) (Transfer, error) {
	var transfer Transfer

	err := db.QueryRow(`
SELECT id, from_transaction_id, to_transaction_id, from_account_id, to_account_id, amount, deleted, created_at, updated_at
FROM transfers
WHERE id = ?
`, id).Scan(
		&transfer.ID,
		&transfer.FromTransactionID,
		&transfer.ToTransactionID,
		&transfer.FromAccountID,
		&transfer.ToAccountID,
		&transfer.Amount,
		&transfer.Deleted,
		&transfer.CreatedAt,
		&transfer.UpdatedAt,
	)
	if err != nil {
		return transfer, err
	}

	return transfer, nil
}

func ListTransfers(db DBTX, filter TransferListFilter) ([]Transfer, error) {
	query := `
SELECT id, from_transaction_id, to_transaction_id, from_account_id, to_account_id, amount, deleted, created_at, updated_at
FROM transfers
WHERE deleted = FALSE
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

	transfers := make([]Transfer, 0)
	for rows.Next() {
		var transfer Transfer
		err = rows.Scan(
			&transfer.ID,
			&transfer.FromTransactionID,
			&transfer.ToTransactionID,
			&transfer.FromAccountID,
			&transfer.ToAccountID,
			&transfer.Amount,
			&transfer.Deleted,
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

func TransferExists(db DBTX, id int64) (bool, error) {
	var exists bool
	err := db.QueryRow(`
SELECT EXISTS(
	SELECT 1 FROM transfers WHERE id = ?
)
`, id).Scan(&exists)
	return exists, err
}

func GetTransferByTransactionID(db DBTX, transactionID int64) (Transfer, error) {
	return getTransferByTransactionID(db, transactionID, false)
}

func GetTransferByTransactionIDIncludingDeleted(db DBTX, transactionID int64) (Transfer, error) {
	return getTransferByTransactionID(db, transactionID, true)
}

func getTransferByTransactionID(db DBTX, transactionID int64, includeDeleted bool) (Transfer, error) {
	var transfer Transfer
	whereDeleted := "AND deleted = FALSE"
	if includeDeleted {
		whereDeleted = ""
	}

	query := `
SELECT id, from_transaction_id, to_transaction_id, from_account_id, to_account_id, amount, deleted, created_at, updated_at
FROM transfers
WHERE (from_transaction_id = ? OR to_transaction_id = ?)
` + whereDeleted + `
LIMIT 1
`
	err := db.QueryRow(query, transactionID, transactionID).Scan(
		&transfer.ID,
		&transfer.FromTransactionID,
		&transfer.ToTransactionID,
		&transfer.FromAccountID,
		&transfer.ToAccountID,
		&transfer.Amount,
		&transfer.Deleted,
		&transfer.CreatedAt,
		&transfer.UpdatedAt,
	)
	if err != nil {
		return transfer, err
	}

	return transfer, nil
}

func InsertTransfer(db DBTX, input CreateTransferInput) (int64, error) {
	result, err := db.Exec(`
INSERT INTO transfers (from_transaction_id, to_transaction_id, from_account_id, to_account_id, amount, deleted)
VALUES (?, ?, ?, ?, ?, FALSE)
`, input.FromTransactionID, input.ToTransactionID, input.FromAccountID, input.ToAccountID, input.Amount)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func UpdateTransferDeleted(db DBTX, transferID int64, deleted bool) error {
	_, err := db.Exec(`
UPDATE transfers
SET deleted = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, deleted, transferID)
	return err
}

func DeleteTransferByID(db DBTX, id int64) error {
	_, err := db.Exec(`
DELETE FROM transfers
WHERE id = ?
`, id)
	return err
}

func (t Transfer) GetFromTransaction(db DBTX) (*Transaction, error) {
	transaction, err := GetTransactionByID(db, t.FromTransactionID)
	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

func (t Transfer) GetToTransaction(db DBTX) (*Transaction, error) {
	transaction, err := GetTransactionByID(db, t.ToTransactionID)
	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

func (t Transfer) GetFromAccount(db DBTX) (*Account, error) {
	account, err := GetAccountByID(db, t.FromAccountID)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (t Transfer) GetToAccount(db DBTX) (*Account, error) {
	account, err := GetAccountByID(db, t.ToAccountID)
	if err != nil {
		return nil, err
	}

	return &account, nil
}
