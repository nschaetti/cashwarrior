package db

import "database/sql"

func InsertTransactionTag(db DBTX, transactionID int64, tagID int64) error {
	_, err := db.Exec(`
INSERT INTO transaction_tags (transaction_id, tag_id)
VALUES (?, ?)
`, transactionID, tagID)
	return err
}

func ListTagsByTransactionID(db DBTX, transactionID int64) ([]Tag, error) {
	rows, err := db.Query(`
SELECT tags.id, tags.name, tags.created_at, tags.updated_at
FROM transaction_tags
JOIN tags ON tags.id = transaction_tags.tag_id
WHERE transaction_tags.transaction_id = ?
ORDER BY tags.name
`, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]Tag, 0)
	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.CreatedAt, &tag.UpdatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

func DeleteTransactionTag(db DBTX, transactionID int64, tagID int64) error {
	_, err := db.Exec(`
DELETE FROM transaction_tags
WHERE transaction_id = ? AND tag_id = ?
`, transactionID, tagID)
	return err
}

func TransactionTagExists(db DBTX, transactionID int64, tagID int64) (bool, error) {
	var exists bool
	err := db.QueryRow(`
SELECT EXISTS(
	SELECT 1
	FROM transaction_tags
	WHERE transaction_id = ? AND tag_id = ?
)
`, transactionID, tagID).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return exists, nil
}
