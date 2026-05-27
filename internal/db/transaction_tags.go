package db

func InsertTransactionTag(db DBTX, transactionID int64, tagID int64) error {
	_, err := db.Exec(`
INSERT INTO transaction_tags (transaction_id, tag_id)
VALUES (?, ?)
`, transactionID, tagID)
	return err
}
