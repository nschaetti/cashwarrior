package db

func UpdateTransactionGroupID(db DBTX, transactionID int64, groupID *int64) error {
	_, err := db.Exec(`
UPDATE transactions
SET group_id = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, groupID, transactionID)
	return err
}
