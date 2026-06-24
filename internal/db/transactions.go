package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Transaction struct {
	ID int64 // primary key

	Identifier  string    // Unique identifier for the transaction
	Type        string    // expense, income, transfer_out, transfer_in
	Amount      float64   // positive for income, negative for expense
	Description string    // free-form text
	Datetime    time.Time // date and time of the transaction

	AccountID   *int64 // foreign key to accounts
	AccountName string
	Currency    string
	CategoryID  *int64 // foreign key to categories
	PlaceID     *int64 // foreign key to places
	GroupID     *int64 // foreign key to groups
	Deleted     bool

	Notes *string // free-form text

	CreatedAt time.Time // creation datetime of the DB entry
	UpdatedAt time.Time // last update datetime of the DB entry
}

type TransactionIDFilter struct {
	ID string
}

func (f TransactionIDFilter) GenerateSQL() (string, []any) {
	return "transactions.identifier = ?", []any{f.ID}
}

func (f TransactionIDFilter) String() string {
	return fmt.Sprintf("<TransactionIDFilter: %s>", f.ID)
}

type TransactionIDListFilter struct {
	IDs []string
}

func (f TransactionIDListFilter) GenerateSQL() (string, []any) {
	placeholders := strings.Repeat("?, ", len(f.IDs))
	placeholders = strings.TrimRight(placeholders, ", ")
	args := make([]any, 0, len(f.IDs))
	for _, id := range f.IDs {
		args = append(args, id)
	}
	return fmt.Sprintf("transactions.identifier IN (%s)", placeholders), args
}
func (f TransactionIDListFilter) String() string {
	return fmt.Sprintf("<TransactionIDListFilter: %v>", f.IDs)
}

type TransactionIDRangeFilter struct {
	From string
	To   string
}

func (f TransactionIDRangeFilter) GenerateSQL() (string, []any) {
	return `(
		(
			CAST(substr(transactions.identifier, 1, 4) AS INTEGER) > CAST(substr(?, 1, 4) AS INTEGER)
			OR (
				CAST(substr(transactions.identifier, 1, 4) AS INTEGER) = CAST(substr(?, 1, 4) AS INTEGER)
				AND CAST(substr(transactions.identifier, 6, 2) AS INTEGER) > CAST(substr(?, 6, 2) AS INTEGER)
			)
			OR (
				CAST(substr(transactions.identifier, 1, 4) AS INTEGER) = CAST(substr(?, 1, 4) AS INTEGER)
				AND CAST(substr(transactions.identifier, 6, 2) AS INTEGER) = CAST(substr(?, 6, 2) AS INTEGER)
				AND CAST(substr(transactions.identifier, 9) AS INTEGER) >= CAST(substr(?, 9) AS INTEGER)
			)
		)
		AND
		(
			CAST(substr(transactions.identifier, 1, 4) AS INTEGER) < CAST(substr(?, 1, 4) AS INTEGER)
			OR (
				CAST(substr(transactions.identifier, 1, 4) AS INTEGER) = CAST(substr(?, 1, 4) AS INTEGER)
				AND CAST(substr(transactions.identifier, 6, 2) AS INTEGER) < CAST(substr(?, 6, 2) AS INTEGER)
			)
			OR (
				CAST(substr(transactions.identifier, 1, 4) AS INTEGER) = CAST(substr(?, 1, 4) AS INTEGER)
				AND CAST(substr(transactions.identifier, 6, 2) AS INTEGER) = CAST(substr(?, 6, 2) AS INTEGER)
				AND CAST(substr(transactions.identifier, 9) AS INTEGER) <= CAST(substr(?, 9) AS INTEGER)
			)
		)
	)`, []any{f.From, f.From, f.From, f.From, f.From, f.From, f.To, f.To, f.To, f.To, f.To, f.To}
}

func (f TransactionIDRangeFilter) String() string {
	return fmt.Sprintf("<TransactionIDRangeFilter: %s..%s>", f.From, f.To)
}

type TransactionAccountNameFilter struct {
	Name string
}

func (f TransactionAccountNameFilter) GenerateSQL() (string, []any) {
	return "accounts.name = ?", []any{f.Name}
}

func (f TransactionAccountNameFilter) String() string {
	return fmt.Sprintf("<TransactionAccountNameFilter: %s>", f.Name)
}

type TransactionCurrencyFilter struct {
	Currency string
}

func (f TransactionCurrencyFilter) GenerateSQL() (string, []any) {
	return "UPPER(accounts.currency) = UPPER(?)", []any{f.Currency}
}

func (f TransactionCurrencyFilter) String() string {
	return fmt.Sprintf("<TransactionCurrencyFilter: %s>", f.Currency)
}

type TransactionStoreNameFilter struct {
	Store string
}

func (f TransactionStoreNameFilter) GenerateSQL() (string, []any) {
	return "places.name LIKE ?", []any{f.Store}
}

func (f TransactionStoreNameFilter) String() string {
	return fmt.Sprintf("<TransactionStoreNameFilter: %s>", f.Store)
}

type TransactionDescriptionFilter struct {
	Description string
}

func (f TransactionDescriptionFilter) GenerateSQL() (string, []any) {
	return "transactions.description LIKE ?", []any{f.Description}
}

func (f TransactionDescriptionFilter) String() string {
	return fmt.Sprintf("<TransactionDescriptionFilter: %s>", f.Description)
}

type TransactionDatetimeFilter struct {
	From time.Time
	To   time.Time
}

func (f TransactionDatetimeFilter) GenerateSQL() (string, []any) {
	return "transactions.datetime BETWEEN ? AND ?", []any{f.From, f.To}
}

func (f TransactionDatetimeFilter) String() string {
	return fmt.Sprintf("<TransactionDatetimeFilter: %s - %s>", f.From, f.To)
}

type TransactionDateFilter struct {
	From string
	To   string
}

func (f TransactionDateFilter) GenerateSQL() (string, []any) {
	return "SUBSTR(transactions.datetime, 1, 10) BETWEEN ? AND ?", []any{f.From, f.To}
}

func (f TransactionDateFilter) String() string {
	return fmt.Sprintf("<TransactionDateFilter: %s - %s>", f.From, f.To)
}

type TransactionGroupNameFilter struct {
	Name string
}

func (f TransactionGroupNameFilter) GenerateSQL() (string, []any) {
	return "transaction_groups.name = ?", []any{f.Name}
}

func (f TransactionGroupNameFilter) String() string {
	return fmt.Sprintf("<TransactionGroupNameFilter: %s>", f.Name)
}

type CreateTransactionInput struct {
	Identifier  string
	Type        string
	Amount      float64
	Description string
	Date        time.Time
	AccountID   int64
	CategoryID  *int64
	PlaceID     *int64
	GroupID     *int64
}

func GetLastTransaction(db DBTX) (Transaction, error) {
	var transaction Transaction
	return getTransactionFromQueryRow(db.QueryRow(`
SELECT transactions.id, transactions.identifier, transactions.type, transactions.amount, transactions.description, transactions.datetime,
       transactions.account_id, COALESCE(accounts.name, ''), COALESCE(accounts.currency, ''),
       transactions.category_id, transactions.place_id, transactions.group_id, transactions.deleted,
       transactions.created_at, transactions.updated_at
FROM transactions
LEFT JOIN accounts ON accounts.id = transactions.account_id
ORDER BY transactions.id DESC
LIMIT 1
`), transaction)
}

func GetTransactionByID(db DBTX, id int64) (Transaction, error) {
	var transaction Transaction
	return getTransactionFromQueryRow(db.QueryRow(`
SELECT transactions.id, transactions.identifier, transactions.type, transactions.amount, transactions.description, transactions.datetime,
       transactions.account_id, COALESCE(accounts.name, ''), COALESCE(accounts.currency, ''),
       transactions.category_id, transactions.place_id, transactions.group_id, transactions.deleted,
       transactions.created_at, transactions.updated_at
FROM transactions
LEFT JOIN accounts ON accounts.id = transactions.account_id
WHERE transactions.id = ?
`, id), transaction)
}

func GetTransactionByIdentifier(db DBTX, identifier string) (Transaction, error) {
	var transaction Transaction
	return getTransactionFromQueryRow(db.QueryRow(`
SELECT transactions.id, transactions.identifier, transactions.type, transactions.amount, transactions.description, transactions.datetime,
       transactions.account_id, COALESCE(accounts.name, ''), COALESCE(accounts.currency, ''),
       transactions.category_id, transactions.place_id, transactions.group_id, transactions.deleted,
       transactions.created_at, transactions.updated_at
FROM transactions
LEFT JOIN accounts ON accounts.id = transactions.account_id
WHERE transactions.identifier = ?
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
		&transaction.AccountName,
		&transaction.Currency,
		&categoryID,
		&placeID,
		&groupID,
		&transaction.Deleted,
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

func GetSumOfTransactions(db DBTX, filters []SQLFilter) (float64, error) {
	query := `
SELECT SUM(amount)
FROM transactions
WHERE deleted = FALSE
`
	args := make([]interface{}, 0)

	for _, filter := range filters {
		filterSQL, filterArgs := filter.GenerateSQL()
		query += "AND " + filterSQL + "\n"
		args = append(args, filterArgs...)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	rows.Next()
	var sum float64
	err = rows.Scan(&sum)
	if err != nil {
		return 0, err
	}
	return sum, nil
}

func ListDeletedTransactions(
	db DBTX,
	dbFilters []SQLFilter,
	runFilters []Filter[Transaction],
) ([]Transaction, error) {
	return ListTransactions(db, dbFilters, runFilters, true)
}

func ListTransactions(
	db DBTX,
	dbFilters []SQLFilter,
	runFilters []Filter[Transaction],
	deleted bool,
) ([]Transaction, error) {
	query := `
SELECT transactions.id, transactions.identifier, transactions.type, transactions.amount, transactions.description, transactions.datetime,
       transactions.account_id, COALESCE(accounts.name, ''), COALESCE(accounts.currency, ''),
       transactions.category_id, transactions.place_id, transactions.group_id, transactions.deleted,
       transactions.created_at, transactions.updated_at
FROM transactions
LEFT JOIN accounts ON accounts.id = transactions.account_id
LEFT JOIN places ON places.id = transactions.place_id
LEFT JOIN transaction_groups ON transaction_groups.id = transactions.group_id
WHERE transactions.deleted = ?
`
	args := make([]interface{}, 0, 1)
	args = append(args, deleted)

	for _, filter := range dbFilters {
		filterSQL, filterArgs := filter.GenerateSQL()
		query += "AND " + filterSQL + "\n"
		args = append(args, filterArgs...)
	}

	query += "ORDER BY transactions.id\n"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := make([]Transaction, 0)
	var addItem bool
	for rows.Next() {
		var transaction Transaction
		var accountID sql.NullInt64
		var accountName string
		var currency string
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
			&accountName,
			&currency,
			&categoryID,
			&placeID,
			&groupID,
			&transaction.Deleted,
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
		transaction.AccountName = accountName
		transaction.Currency = currency
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

		addItem = true
		for _, filter := range runFilters {
			if filter.Reject(transaction) {
				addItem = false
				break
			}
		}

		if addItem {
			transactions = append(transactions, transaction)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func TransactionExists(db DBTX, id int64) (bool, error) {
	var exists bool
	err := db.QueryRow(`
SELECT EXISTS(
	SELECT 1 FROM transactions WHERE id = ?
)
`, id).Scan(&exists)
	return exists, err
}

func InsertTransaction(db DBTX, input CreateTransactionInput) (int64, error) {
	transactionType := input.Type
	if transactionType == "" {
		transactionType = "expense"
	}

	result, err := db.Exec(`
INSERT INTO transactions (identifier, type, amount, description, datetime, account_id, category_id, place_id, group_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
`, input.Identifier, transactionType, input.Amount, input.Description, input.Date, input.AccountID, input.CategoryID, input.PlaceID, input.GroupID)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func UpdateTransactionDescription(db DBTX, transactionID int64, description string) error {
	_, err := db.Exec(`
UPDATE transactions
SET description = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, description, transactionID)
	return err
}

func UpdateTransactionAmount(db DBTX, transactionID int64, amount float64) error {
	_, err := db.Exec(`
UPDATE transactions
SET amount = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, amount, transactionID)
	return err
}

func UpdateTransactionDatetime(db DBTX, transactionID int64, datetime time.Time) error {
	_, err := db.Exec(`
UPDATE transactions
SET datetime = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, datetime, transactionID)
	return err
}

func UpdateTransactionAccountID(db DBTX, transactionID int64, accountID *int64) error {
	_, err := db.Exec(`
UPDATE transactions
SET account_id = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, accountID, transactionID)
	return err
}

func UpdateTransactionCategoryID(db DBTX, transactionID int64, categoryID *int64) error {
	_, err := db.Exec(`
UPDATE transactions
SET category_id = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, categoryID, transactionID)
	return err
}

func UpdateTransactionPlaceID(db DBTX, transactionID int64, placeID *int64) error {
	_, err := db.Exec(`
UPDATE transactions
SET place_id = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, placeID, transactionID)
	return err
}

func UpdateTransactionDeleted(db DBTX, transactionID int64, deleted bool) error {
	_, err := db.Exec(`
UPDATE transactions
SET deleted = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, deleted, transactionID)
	return err
}

func DeleteTransactionByID(db DBTX, transactionID int64) error {
	_, err := db.Exec(`
DELETE FROM transactions
WHERE id = ?
`, transactionID)
	return err
}

func PurgeTransactionByIdentifier(db DBTX, identifier string) error {
	transaction, err := GetTransactionByIdentifier(db, identifier)
	if err != nil {
		return err
	}

	transfer, err := GetTransferByTransactionIDIncludingDeleted(db, transaction.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return DeleteTransactionByID(db, transaction.ID)
	}
	if err != nil {
		return err
	}

	if err := DeleteTransferByID(db, transfer.ID); err != nil {
		return err
	}
	if err := DeleteTransactionByID(db, transfer.FromTransactionID); err != nil {
		return err
	}
	if transfer.ToTransactionID != transfer.FromTransactionID {
		if err := DeleteTransactionByID(db, transfer.ToTransactionID); err != nil {
			return err
		}
	}

	return nil
}

func (t Transaction) GetAccount(db DBTX) (*Account, error) {
	if t.AccountID == nil {
		return nil, nil
	}

	account, err := GetAccountByID(db, *t.AccountID)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (t Transaction) GetCategory(db DBTX) (*Category, error) {
	if t.CategoryID == nil {
		return nil, nil
	}

	category, err := GetCategoryByID(db, *t.CategoryID)
	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (t Transaction) GetPlace(db DBTX) (*Place, error) {
	if t.PlaceID == nil {
		return nil, nil
	}

	place, err := GetStoreByID(db, *t.PlaceID)
	if err != nil {
		return nil, err
	}

	return &place, nil
}

func (t Transaction) GetGroup(db DBTX) (*TransactionGroup, error) {
	if t.GroupID == nil {
		return nil, nil
	}

	group, err := GetGroupByID(db, *t.GroupID)
	if err != nil {
		return nil, err
	}

	return &group, nil
}
