package db

func UpdateTransactionGroupID(db DBTX, transactionID int64, groupID *int64) error {
	_, err := db.Exec(`
UPDATE transactions
SET group_id = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, groupID, transactionID)
	return err
}

func CountTransactionsByGroupID(db DBTX, groupID int64) (int, error) {
	var count int
	err := db.QueryRow(`
SELECT COUNT(*)
FROM transactions
WHERE group_id = ?
`, groupID).Scan(&count)
	return count, err
}

func ClearTransactionsGroupID(db DBTX, groupID int64) error {
	_, err := db.Exec(`
UPDATE transactions
SET group_id = NULL, updated_at = CURRENT_TIMESTAMP
WHERE group_id = ?
`, groupID)
	return err
}
