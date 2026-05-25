package db

import (
	"database/sql"
	"time"
)

type TransactionGroup struct {
	ID        int64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TransactionGroupListFilter struct {
	ID     *int64
	Limit  int
	Offset int
}

type CreateTransactionGroupInput struct {
	Name        string
	Description string
}

func GetTransactionGroupByID(db *sql.DB, id int64) (TransactionGroup, error) {
	var transactionGroup TransactionGroup

	err := db.QueryRow(`
SELECT id, name, created_at, updated_at
FROM transaction_groups
WHERE id = ?
`, id).Scan(
		&transactionGroup.ID,
		&transactionGroup.Name,
		&transactionGroup.CreatedAt,
		&transactionGroup.UpdatedAt,
	)
	if err != nil {
		return transactionGroup, err
	}

	return transactionGroup, nil
}

func GetTransactionGroupByName(db *sql.DB, name string) (TransactionGroup, error) {
	var transactionGroup TransactionGroup

	err := db.QueryRow(`
SELECT id, name, created_at, updated_at
FROM transaction_groups
WHERE name = ?
`, name).Scan(
		&transactionGroup.ID,
		&transactionGroup.Name,
		&transactionGroup.CreatedAt,
		&transactionGroup.UpdatedAt,
	)
	if err != nil {
		return transactionGroup, err
	}

	return transactionGroup, nil
}

func ListTransactionGroups(db *sql.DB, filter TransactionGroupListFilter) ([]TransactionGroup, error) {
	query := `
SELECT id, name, created_at, updated_at
FROM transaction_groups
WHERE 1 = 1
`
	args := make([]interface{}, 0)

	if filter.ID != nil {
		query += "AND id = ?\n"
		args = append(args, *filter.ID)
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

	transactionGroups := make([]TransactionGroup, 0)
	for rows.Next() {
		var transactionGroup TransactionGroup
		err = rows.Scan(
			&transactionGroup.ID,
			&transactionGroup.Name,
			&transactionGroup.CreatedAt,
			&transactionGroup.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactionGroups = append(transactionGroups, transactionGroup)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return transactionGroups, nil
}

func TransactionGroupExists(db *sql.DB, name string) (bool, error) {
	var exists bool
	err := db.QueryRow(`
SELECT EXISTS(
	SELECT 1
	FROM transaction_groups
	WHERE name = ?
)
`, name).Scan(&exists)
	return exists, err
}

func InsertTransactionGroup(db *sql.DB, input CreateTransactionGroupInput) (int64, error) {
	result, err := db.Exec(`
INSERT INTO transaction_groups (name)
VALUES (?)
`, input.Name)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
