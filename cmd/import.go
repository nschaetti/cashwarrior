package cmd

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/domain"
	"github.com/nschaetti/cashwarrior/internal/parser"
	"github.com/pterm/pterm"
)

type csvSchema struct {
	Table    string
	Required map[string]bool
	Optional map[string]bool
	Ignored  map[string]bool
}

type csvImportData struct {
	Headers []string
	Rows    []map[string]string
}

var csvSchemas = []csvSchema{
	newCSVSchema("transactions", []string{"type", "amount", "description", "datetime", "account", "place"}, []string{"identifier", "category", "group", "deleted"}, []string{"created_at", "updated_at"}),
	newCSVSchema("accounts", []string{"name", "currency"}, []string{"initial_balance"}, []string{"created_at", "updated_at"}),
	newCSVSchema("budgets", []string{"category", "name", "description", "amount", "currency", "period"}, nil, []string{"created_at", "updated_at"}),
	newCSVSchema("categories", []string{"name"}, []string{"parent"}, []string{"created_at", "updated_at"}),
	newCSVSchema("places", []string{"place_name"}, nil, []string{"created_at", "updated_at"}),
	newCSVSchema("tags", []string{"tag_name"}, nil, []string{"created_at", "updated_at"}),
	newCSVSchema("transaction_groups", []string{"group_name"}, nil, []string{"created_at", "updated_at"}),
	newCSVSchema("transaction_tags", []string{"transaction", "tag"}, nil, nil),
	newCSVSchema("transfers", []string{"from_transaction", "to_transaction", "from_account", "to_account", "amount"}, nil, []string{"created_at", "updated_at"}),
}

var allowedBudgetPeriods = map[string]bool{
	"day": true, "week": true, "month": true, "2months": true, "3months": true, "4months": true, "6months": true, "year": true,
}

func newCSVSchema(table string, required []string, optional []string, ignored []string) csvSchema {
	return csvSchema{
		Table:    table,
		Required: toSet(required),
		Optional: toSet(optional),
		Ignored:  toSet(ignored),
	}
}

func toSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[item] = true
	}
	return set
}

func normalizeHeader(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	return strings.TrimPrefix(value, "\ufeff")
}

func schemaMatches(schema csvSchema, headers []string) bool {
	seen := make(map[string]bool, len(headers))
	for _, header := range headers {
		if header == "" {
			return false
		}
		if seen[header] {
			return false
		}
		seen[header] = true
		if !schema.Required[header] && !schema.Optional[header] && !schema.Ignored[header] {
			return false
		}
	}
	for header := range schema.Required {
		if !seen[header] {
			return false
		}
	}
	return true
}

func detectCSVSchema(headers []string) (csvSchema, error) {
	matches := make([]csvSchema, 0)
	for _, schema := range csvSchemas {
		if schemaMatches(schema, headers) {
			matches = append(matches, schema)
		}
	}
	if len(matches) == 0 {
		return csvSchema{}, fmt.Errorf("unknown CSV schema")
	}
	if len(matches) > 1 {
		names := make([]string, 0, len(matches))
		for _, match := range matches {
			names = append(names, match.Table)
		}
		return csvSchema{}, fmt.Errorf("ambiguous CSV schema: %s", strings.Join(names, ", "))
	}
	return matches[0], nil
}

func loadCSV(path string) (csvImportData, error) {
	file, err := os.Open(path)
	if err != nil {
		return csvImportData{}, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return csvImportData{}, err
	}
	if len(records) == 0 {
		return csvImportData{}, fmt.Errorf("empty CSV file")
	}

	headers := make([]string, len(records[0]))
	for i, header := range records[0] {
		headers[i] = normalizeHeader(header)
	}

	rows := make([]map[string]string, 0, len(records)-1)
	for rowIndex, record := range records[1:] {
		if len(record) != len(headers) {
			return csvImportData{}, fmt.Errorf("csv line %d: expected %d columns, got %d", rowIndex+2, len(headers), len(record))
		}
		row := make(map[string]string, len(headers))
		for colIndex, header := range headers {
			row[header] = strings.TrimSpace(record[colIndex])
		}
		row["__line"] = strconv.Itoa(rowIndex + 2)
		rows = append(rows, row)
	}

	return csvImportData{Headers: headers, Rows: rows}, nil
}

func Import(parsed parser.ParsedCmdLine, cfg config.Config, query db.DBTX) error {
	data, err := loadCSV(parsed.Args[0].RawString())
	if err != nil {
		return err
	}

	schema, err := detectCSVSchema(data.Headers)
	if err != nil {
		return err
	}

	switch schema.Table {
	case "transactions":
		err = importTransactions(data.Rows, cfg, query)
	case "accounts":
		err = importAccounts(data.Rows, query)
	case "budgets":
		err = importBudgets(data.Rows, query)
	case "categories":
		err = importCategories(data.Rows, query)
	case "places":
		err = importPlaces(data.Rows, query)
	case "tags":
		err = importTags(data.Rows, query)
	case "transaction_groups":
		err = importTransactionGroups(data.Rows, query)
	case "transaction_tags":
		err = importTransactionTags(data.Rows, query)
	case "transfers":
		err = importTransfers(data.Rows, query)
	default:
		err = fmt.Errorf("unsupported CSV schema %s", schema.Table)
	}
	if err != nil {
		return err
	}

	pterm.Success.Printf("Imported %d rows into %s\n", len(data.Rows), schema.Table)
	return nil
}

func importTransactions(rows []map[string]string, cfg config.Config, query db.DBTX) error {
	nextNumberByMonth := make(map[string]int)
	for i, row := range rows {
		amount, err := strconv.ParseFloat(row["amount"], 64)
		if err != nil {
			return fmt.Errorf("csv line %d: invalid amount %q", i+2, row["amount"])
		}
		datetime, err := parseDateTime(row["datetime"], cfg)
		if err != nil {
			return fmt.Errorf("csv line %d: invalid datetime %q", i+2, row["datetime"])
		}

		identifier := row["identifier"]
		if identifier == "" {
			identifier, err = generateImportTransactionIdentifier(query, datetime, nextNumberByMonth)
			if err != nil {
				return fmt.Errorf("csv line %d: %w", i+2, err)
			}
		} else {
			if _, err := db.GetTransactionByIdentifier(query, identifier); err == nil {
				return fmt.Errorf("csv line %d: transaction %s already exists", i+2, identifier)
			} else if !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("csv line %d: %w", i+2, err)
			}
			reserveImportTransactionIdentifier(identifier, nextNumberByMonth)
		}

		deleted := false
		if row["deleted"] != "" {
			deleted, err = strconv.ParseBool(row["deleted"])
			if err != nil {
				return fmt.Errorf("csv line %d: invalid deleted value %q", i+2, row["deleted"])
			}
		}

		account, err := db.GetAccountByName(query, row["account"])
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("csv line %d: account %s does not exist", i+2, row["account"])
		}
		if err != nil {
			return fmt.Errorf("csv line %d: %w", i+2, err)
		}

		placeID, err := resolveOrCreateOptionalPlaceID(query, row["place"], i+2)
		if err != nil {
			return err
		}

		categoryID, err := resolveOrCreateOptionalCategoryID(query, row["category"], i+2)
		if err != nil {
			return err
		}
		groupID, err := resolveOrCreateOptionalGroupID(query, row["group"], i+2)
		if err != nil {
			return err
		}

		transactionID, err := db.InsertTransaction(query, db.CreateTransactionInput{
			Identifier:  identifier,
			Type:        row["type"],
			Amount:      amount,
			Description: row["description"],
			Datetime:    datetime,
			AccountID:   account.ID,
			CategoryID:  categoryID,
			PlaceID:     placeID,
			GroupID:     groupID,
		})
		if err != nil {
			return fmt.Errorf("csv line %d: %w", i+2, err)
		}
		if deleted {
			if err := db.UpdateTransactionDeleted(query, transactionID, true); err != nil {
				return fmt.Errorf("csv line %d: %w", i+2, err)
			}
		}
	}
	return nil
}

func importAccounts(rows []map[string]string, query db.DBTX) error {
	for i, row := range rows {
		initialBalance := 0.0
		if raw := strings.TrimSpace(row["initial_balance"]); raw != "" {
			parsed, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				return fmt.Errorf("csv line %d: invalid initial_balance %q", i+2, row["initial_balance"])
			}
			initialBalance = parsed
		}
		if _, err := db.InsertAccount(query, db.CreateAccountInput{Name: row["name"], Currency: row["currency"], InitialBalance: initialBalance}); err != nil {
			return fmt.Errorf("csv line %d: %w", i+2, err)
		}
	}
	return nil
}

func importBudgets(rows []map[string]string, query db.DBTX) error {
	for i, row := range rows {
		category, err := db.GetCategoryByName(query, row["category"])
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("csv line %d: category %s does not exist", i+2, row["category"])
		}
		if err != nil {
			return fmt.Errorf("csv line %d: %w", i+2, err)
		}
		amount, err := strconv.ParseFloat(row["amount"], 64)
		if err != nil {
			return fmt.Errorf("csv line %d: invalid amount %q", i+2, row["amount"])
		}
		if !allowedBudgetPeriods[row["period"]] {
			return fmt.Errorf("csv line %d: invalid period %q", i+2, row["period"])
		}
		if _, err := db.InsertBudget(query, db.CreateBudgetInput{CategoryID: category.ID, Name: row["name"], Description: row["description"], Amount: amount, Currency: row["currency"], Period: row["period"]}); err != nil {
			return fmt.Errorf("csv line %d: %w", i+2, err)
		}
	}
	return nil
}

func importCategories(rows []map[string]string, query db.DBTX) error {
	pending := make([]map[string]string, len(rows))
	copy(pending, rows)

	for len(pending) > 0 {
		progress := false
		next := make([]map[string]string, 0)
		for i, row := range pending {
			line := csvRowLine(row, i+2)
			var parentID *int64
			if parentName := row["parent"]; parentName != "" {
				parent, err := db.GetCategoryByName(query, parentName)
				if errors.Is(err, sql.ErrNoRows) {
					next = append(next, row)
					continue
				}
				if err != nil {
					return fmt.Errorf("csv line %d: %w", line, err)
				}
				parentID = &parent.ID
			}
			if _, err := db.InsertCategory(query, db.CreateCategoryInput{Name: row["name"], ParentID: parentID}); err != nil {
				return fmt.Errorf("csv line %d: %w", line, err)
			}
			progress = true
		}
		if !progress {
			return fmt.Errorf("csv line %d: unresolved category parent dependencies", csvRowLine(pending[0], 2))
		}
		pending = next
	}
	return nil
}

func csvRowLine(row map[string]string, fallback int) int {
	if raw, ok := row["__line"]; ok {
		if line, err := strconv.Atoi(raw); err == nil {
			return line
		}
	}
	return fallback
}

func importPlaces(rows []map[string]string, query db.DBTX) error {
	for i, row := range rows {
		if _, err := db.InsertPlace(query, db.CreatePlaceInput{Name: row["place_name"]}); err != nil {
			return fmt.Errorf("csv line %d: %w", i+2, err)
		}
	}
	return nil
}

func importTags(rows []map[string]string, query db.DBTX) error {
	for i, row := range rows {
		if _, err := db.InsertTag(query, db.CreateTagInput{Name: row["tag_name"]}); err != nil {
			return fmt.Errorf("csv line %d: %w", i+2, err)
		}
	}
	return nil
}

func importTransactionGroups(rows []map[string]string, query db.DBTX) error {
	for i, row := range rows {
		if _, err := db.InsertTransactionGroup(query, db.CreateTransactionGroupInput{Name: row["group_name"]}); err != nil {
			return fmt.Errorf("csv line %d: %w", i+2, err)
		}
	}
	return nil
}

func importTransactionTags(rows []map[string]string, query db.DBTX) error {
	for i, row := range rows {
		transaction, err := db.GetTransactionByIdentifier(query, row["transaction"])
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("csv line %d: transaction %s does not exist", i+2, row["transaction"])
		}
		if err != nil {
			return fmt.Errorf("csv line %d: %w", i+2, err)
		}
		tag, err := db.GetTagByName(query, row["tag"])
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("csv line %d: tag %s does not exist", i+2, row["tag"])
		}
		if err != nil {
			return fmt.Errorf("csv line %d: %w", i+2, err)
		}
		if err := db.InsertTransactionTag(query, transaction.ID, tag.ID); err != nil {
			return fmt.Errorf("csv line %d: %w", i+2, err)
		}
	}
	return nil
}

func importTransfers(rows []map[string]string, query db.DBTX) error {
	for i, row := range rows {
		fromTransaction, err := db.GetTransactionByIdentifier(query, row["from_transaction"])
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("csv line %d: transaction %s does not exist", i+2, row["from_transaction"])
		}
		if err != nil {
			return fmt.Errorf("csv line %d: %w", i+2, err)
		}
		toTransaction, err := db.GetTransactionByIdentifier(query, row["to_transaction"])
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("csv line %d: transaction %s does not exist", i+2, row["to_transaction"])
		}
		if err != nil {
			return fmt.Errorf("csv line %d: %w", i+2, err)
		}
		fromAccount, err := db.GetAccountByName(query, row["from_account"])
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("csv line %d: account %s does not exist", i+2, row["from_account"])
		}
		if err != nil {
			return fmt.Errorf("csv line %d: %w", i+2, err)
		}
		toAccount, err := db.GetAccountByName(query, row["to_account"])
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("csv line %d: account %s does not exist", i+2, row["to_account"])
		}
		if err != nil {
			return fmt.Errorf("csv line %d: %w", i+2, err)
		}
		amount, err := strconv.ParseFloat(row["amount"], 64)
		if err != nil {
			return fmt.Errorf("csv line %d: invalid amount %q", i+2, row["amount"])
		}
		if _, err := db.InsertTransfer(query, db.CreateTransferInput{FromTransactionID: fromTransaction.ID, ToTransactionID: toTransaction.ID, FromAccountID: fromAccount.ID, ToAccountID: toAccount.ID, Amount: amount}); err != nil {
			return fmt.Errorf("csv line %d: %w", i+2, err)
		}
	}
	return nil
}

func resolveOrCreateOptionalPlaceID(query db.DBTX, name string, line int) (*int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil
	}
	place, err := db.GetPlaceByName(query, name)
	if errors.Is(err, sql.ErrNoRows) {
		id, insertErr := db.InsertPlace(query, db.CreatePlaceInput{Name: name})
		if insertErr != nil {
			return nil, fmt.Errorf("csv line %d: %w", line, insertErr)
		}
		return &id, nil
	}
	if err != nil {
		return nil, fmt.Errorf("csv line %d: %w", line, err)
	}
	return &place.ID, nil
}

func resolveOrCreateOptionalCategoryID(query db.DBTX, name string, line int) (*int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil
	}
	category, err := db.GetCategoryByName(query, name)
	if errors.Is(err, sql.ErrNoRows) {
		id, insertErr := db.InsertCategory(query, db.CreateCategoryInput{Name: name})
		if insertErr != nil {
			return nil, fmt.Errorf("csv line %d: %w", line, insertErr)
		}
		return &id, nil
	}
	if err != nil {
		return nil, fmt.Errorf("csv line %d: %w", line, err)
	}
	return &category.ID, nil
}

func resolveOrCreateOptionalGroupID(query db.DBTX, name string, line int) (*int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil
	}
	group, err := db.GetTransactionGroupByName(query, name)
	if errors.Is(err, sql.ErrNoRows) {
		id, insertErr := db.InsertTransactionGroup(query, db.CreateTransactionGroupInput{Name: name})
		if insertErr != nil {
			return nil, fmt.Errorf("csv line %d: %w", line, insertErr)
		}
		return &id, nil
	}
	if err != nil {
		return nil, fmt.Errorf("csv line %d: %w", line, err)
	}
	return &group.ID, nil
}

func parseDateTime(value string, cfg config.Config) (time.Time, error) {
	value = strings.TrimSpace(value)
	layouts := []string{
		cfg.Display.DateFormat,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
		"02.01.2006 15:04:05",
		"02.01.2006 15:04",
		"02.01.2006",
		"02/01/2006 15:04:05",
		"02/01/2006 15:04",
		"02/01/2006",
	}

	seen := make(map[string]bool, len(layouts))
	for _, layout := range layouts {
		if layout == "" || seen[layout] {
			continue
		}
		seen[layout] = true
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, parsed.Location()), nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported datetime format %q", value)
}

func generateImportTransactionIdentifier(query db.DBTX, datetime time.Time, nextNumberByMonth map[string]int) (string, error) {
	monthKey := fmt.Sprintf("%04d.%02d", datetime.Year(), int(datetime.Month()))
	nextNumber, ok := nextNumberByMonth[monthKey]
	if !ok {
		lastNumber, err := getMaxTransactionNumberForMonth(query, datetime.Year(), int(datetime.Month()))
		if err != nil {
			return "", err
		}
		nextNumber = lastNumber + 1
	}
	nextNumberByMonth[monthKey] = nextNumber + 1
	return fmt.Sprintf("%s.%d", monthKey, nextNumber), nil
}

func reserveImportTransactionIdentifier(identifier string, nextNumberByMonth map[string]int) {
	parsedIdentifier, err := domain.ParseTransactionID(identifier)
	if err != nil {
		return
	}
	monthKey := fmt.Sprintf("%04d.%02d", parsedIdentifier.Year, parsedIdentifier.Month)
	nextNumber := parsedIdentifier.Num + 1
	if current, ok := nextNumberByMonth[monthKey]; !ok || nextNumber > current {
		nextNumberByMonth[monthKey] = nextNumber
	}
}

func getMaxTransactionNumberForMonth(query db.DBTX, year int, month int) (int, error) {
	prefix := fmt.Sprintf("%04d.%02d.", year, month)
	var maxNumber sql.NullInt64
	err := query.QueryRow(`
SELECT MAX(CAST(SUBSTR(identifier, LENGTH(?) + 1) AS INTEGER))
FROM transactions
WHERE identifier LIKE ?
`, prefix, prefix+"%").Scan(&maxNumber)
	if err != nil {
		return 0, err
	}
	if !maxNumber.Valid {
		return 0, nil
	}
	return int(maxNumber.Int64), nil
}
