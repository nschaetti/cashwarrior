package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/domain"
	"github.com/nschaetti/cashwarrior/internal/parser"
	"github.com/pterm/pterm"
)

var CommandValidations = []func(
	parser.ParsedCmdLine,
	config.Config,
	*sql.DB,
	map[parser.TokenKind]int,
) (parser.ParsedCmdLine, error){
	validateAtLeastTwoArgs,
	validateNoFilters,
	validateAmount,
	validateAttributes,
}

func validateAtLeastTwoArgs(
	parsed parser.ParsedCmdLine,
	config config.Config,
	db *sql.DB,
	counts map[parser.TokenKind]int,
) (parser.ParsedCmdLine, error) {
	if len(parsed.Args) < 2 {
		return parsed, fmt.Errorf("we need at least an amount and a description")
	}
	return parsed, nil
}

func validateNoFilters(
	parsed parser.ParsedCmdLine,
	config config.Config,
	db *sql.DB,
	counts map[parser.TokenKind]int,
) (parser.ParsedCmdLine, error) {
	if len(parsed.Filters) != 0 {
		return parsed, fmt.Errorf("no filters allowed")
	}
	return parsed, nil
}

func validateAmount(
	parsed parser.ParsedCmdLine,
	config config.Config,
	db *sql.DB,
	counts map[parser.TokenKind]int,
) (parser.ParsedCmdLine, error) {
	// Zero or multiple amounts => problem
	if counts[parser.TokenAmount] == 0 {
		return parsed, fmt.Errorf("no amount specified")
	} else if counts[parser.TokenAmount] > 1 {
		name, err := pterm.DefaultInteractiveTextInput.
			WithDefaultText("y").
			Show("You specified multiple amounts. Would you like to sum them up?:")
		if err != nil {
			return parsed, err
		}
		if name != "y" {
			return parsed, fmt.Errorf("multiple amounts specified, only one allowed if not summed up")
		}

		// Sum up amounts
		var sumAmount float32
		for _, amount := range parsed.GetAmounts() {
			sumAmount += amount.Amount
		}

		// No zero amount allowed
		if sumAmount == 0 {
			return parsed, fmt.Errorf("summed up amount is zero")
		}

		// Remove amounts from parsed command line
		parsed.RemoveByKind(parser.TokenAmount)

		// Add summed up amount
		parsed.Append(parser.Token{
			Raw:    strconv.FormatFloat(float64(sumAmount), 'f', 2, 32),
			Amount: sumAmount,
			Kind:   parser.TokenAmount,
		}, false)

		return parsed, nil
	}

	// We have an amount, check if it's valid
	amount := parsed.GetAmounts()[0]

	// No zero amount allowed
	if amount.Amount == 0 {
		return parsed, fmt.Errorf("amount cannot be zero")
	}

	return parsed, nil
}

func validateAttributes(
	parsed parser.ParsedCmdLine,
	config config.Config,
	db *sql.DB,
	counts map[parser.TokenKind]int,
) (parser.ParsedCmdLine, error) {
	// Get count of attributes
	attrsCount := parsed.GetAttributesCount(false)

	for attr, count := range attrsCount {
		if count > 1 {
			return parsed, fmt.Errorf("attribute %s specified multiple times", attr)
		}
	}

	return parsed, nil
}

func runCommandLineValidation(parsed parser.ParsedCmdLine, config config.Config, db *sql.DB, counts map[parser.TokenKind]int) (parser.ParsedCmdLine, error) {
	for _, check := range CommandValidations {
		var err error
		parsed, err = check(parsed, config, db, counts)
		if err != nil {
			return parsed, err
		}
	}
	return parsed, nil
}

func getTransactionDescription(parsed parser.ParsedCmdLine, counts map[parser.TokenKind]int) string {
	if counts[parser.TokenText] == 1 {
		return parsed.Args[0].Raw
	}
	var desc string
	for _, arg := range parsed.Args {
		if arg.Kind != parser.TokenText {
			continue
		}
		desc += arg.Raw + " "
	}
	return desc[0 : len(desc)-1]
}

func getNextIdentifier(cashDb *sql.DB) (domain.TransactionID, error) {
	// Get next transaction identifier
	lastTransaction, err := db.GetLastTransaction(cashDb)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.CurrentTransactionID(1), nil
	} else if err != nil {
		return domain.CurrentTransactionID(0), err
	}

	// Parse identifier
	id, err := domain.ParseTransactionID(lastTransaction.Identifier)
	if err != nil {
		return domain.TransactionID{}, err
	}

	// Same year-month -> increment sequence number
	if id.Month == int(time.Now().Month()) && id.Year == int(time.Now().Year()) {
		return domain.CurrentTransactionID(id.Num + 1), nil
	}

	return domain.CurrentTransactionID(0), nil
}

func getTransactionAmount(parsed parser.ParsedCmdLine, counts map[parser.TokenKind]int) float32 {
	return parsed.GetAmounts()[0].Amount
}

func getTransactionDatetime(
	attributes map[string]string,
	config config.Config,
) (time.Time, error) {
	// Check attributes exists
	var dateTimeExists bool
	var dateExists bool
	var timeExists bool
	for key, _ := range attributes {
		switch key {
		case "datetime":
			dateTimeExists = true
		case "date":
			dateExists = true
		case "time":
			timeExists = true
		}
	}

	// Can be specified by datetime, or date alone (time 00:00:00), or date+time
	if dateTimeExists && dateExists {
		return time.Time{}, fmt.Errorf("datetime and date specified, only one allowed")
	}
	if dateTimeExists && timeExists {
		return time.Time{}, fmt.Errorf("datetime and time specified, only one allowed")

	}

	// Datetime given
	if dateTimeExists {
		return time.Parse(config.Display.DateFormat, attributes["datetime"])
	}

	// Date and time given
	if dateExists && timeExists {
		return time.Parse(
			config.Display.DateFormat,
			attributes["date"]+" "+attributes["time"],
		)
	} else if dateExists {
		return time.Parse(strings.Split(config.Display.DateFormat, " ")[0], attributes["date"])
	} else if timeExists {
		return time.Parse(strings.Split(config.Display.DateFormat, " ")[1], attributes["time"])
	}

	return time.Now(), nil
}

func getAttributes(parsed parser.ParsedCmdLine) map[string]string {
	var attributes map[string]string = make(map[string]string)
	for _, arg := range parsed.Args {
		if arg.Kind != parser.TokenAttribute {
			continue
		}
		attributes[arg.Key] = arg.Value
	}
	return attributes
}

func getTransactionStore(cashDb *sql.DB, attributes map[string]string) (sql.NullInt64, error) {
	var storeID sql.NullInt64
	if _, ok := attributes["store"]; ok {
		store, err := db.GetPlaceByName(cashDb, attributes["store"])
		if errors.Is(err, sql.ErrNoRows) {
			// Create new store
			newStoreID, cErr := db.InsertPlace(cashDb, db.CreatePlaceInput{Name: attributes["store"]})
			if cErr != nil {
				return storeID, cErr
			}
			storeID = sql.NullInt64{Int64: newStoreID, Valid: true}
			return storeID, nil
		} else if err != nil {
			// SQL Error
			return storeID, err
		}
		storeID = sql.NullInt64{Int64: store.ID, Valid: true}
	} else {
		return storeID, fmt.Errorf("no store specified")
	}
	return storeID, nil
}

func getTransactionAccount(cashDb *sql.DB, attributes map[string]string, config config.Config) (int64, error) {
	var accountName string
	if _, ok := attributes["account"]; ok {
		accountName = attributes["account"]
	} else {
		accountName = config.Default.Account
	}
	// Get account
	transactionAccount, err := db.GetAccountByName(cashDb, accountName)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("account %s does not exist, create it first", accountName)
	} else if err != nil {
		return 0, err
	}
	return transactionAccount.ID, nil
}

func getTransactionCategory(cashDb *sql.DB, attributes map[string]string) (sql.NullInt64, error) {
	var categoryID sql.NullInt64
	// Category given
	if _, ok := attributes["category"]; ok {
		attrCategory := attributes["category"]
		category, err := db.GetCategoryByName(cashDb, attrCategory)
		if errors.Is(err, sql.ErrNoRows) {
			// Create new category with root as parent
			parentID := int64(1)
			newCategoryID, cErr := db.InsertCategory(
				cashDb,
				db.CreateCategoryInput{Name: attrCategory, ParentID: &parentID},
			)
			if cErr != nil {
				return sql.NullInt64{}, cErr
			}
			return sql.NullInt64{Int64: newCategoryID, Valid: true}, nil
		} else if err != nil {
			return categoryID, err
		}
		return sql.NullInt64{Int64: category.ID, Valid: true}, nil
	}
	return categoryID, nil
}

func getTransactionGroup(cashDb *sql.DB, attributes map[string]string) (sql.NullInt64, error) {
	if _, ok := attributes["group"]; ok {
		attrGroup := attributes["group"]
		group, err := db.GetTransactionGroupByName(cashDb, attrGroup)
		if errors.Is(err, sql.ErrNoRows) {
			newGroupID, cErr := db.InsertTransactionGroup(cashDb, db.CreateTransactionGroupInput{Name: attrGroup})
			if cErr != nil {
				return sql.NullInt64{}, cErr
			}
			return sql.NullInt64{Int64: newGroupID, Valid: true}, nil
		} else if err != nil {
			return sql.NullInt64{}, err
		}
		return sql.NullInt64{Int64: group.ID, Valid: true}, nil
	}
	return sql.NullInt64{}, nil
}

func Add(parsed parser.ParsedCmdLine, config config.Config, cashDb *sql.DB) error {
	// Get count by token kind for validation
	tokenKindsCount := parsed.GetTokenKindCount(false)

	// We check the commnand line first
	parsed, err := runCommandLineValidation(parsed, config, cashDb, tokenKindsCount)
	if err != nil {
		return err
	}

	// Get attributes
	attributes := getAttributes(parsed)

	// Get transaction description (merge it if necessary) & amount
	desc := getTransactionDescription(parsed, tokenKindsCount)
	amount := getTransactionAmount(parsed, tokenKindsCount)

	// Get next transaction identifier
	nextIdentifier, err := getNextIdentifier(cashDb)
	if err != nil {
		return err
	}

	// Transaction type
	var transactionType string
	if amount < 0.0 {
		transactionType = "expense"
	} else {
		transactionType = "income"
	}

	// Get transaction datetime
	transactionTime, err := getTransactionDatetime(attributes, config)
	if err != nil {
		return err
	}

	// Get store
	transactionStore, err := getTransactionStore(cashDb, attributes)
	if err != nil {
		return err
	}

	// Get account
	transactionAccount, err := getTransactionAccount(cashDb, attributes, config)
	if err != nil {
		return err
	}

	// Get category
	transactionCategory, err := getTransactionCategory(cashDb, attributes)
	if err != nil {
		return err
	}

	// Get group
	transactionGroup, err := getTransactionGroup(cashDb, attributes)
	if err != nil {
		return err
	}

	// Insert transaction
	transactionID, err := db.InsertTransaction(
		cashDb,
		db.CreateTransactionInput{
			Identifier:  fmt.Sprintf("%s", nextIdentifier),
			Type:        transactionType,
			Amount:      float64(amount),
			Description: desc,
			Datetime:    transactionTime,
			AccountID:   transactionAccount,
			CategoryID:  db.NullInt64ToPtr(transactionCategory),
			PlaceID:     db.NullInt64ToPtr(transactionStore),
			GroupID:     db.NullInt64ToPtr(transactionGroup),
		},
	)
	if err != nil {
		return err
	}

	// Show success message
	pterm.Success.Println("Transaction added with id: " + strconv.FormatInt(transactionID, 10) + "")

	return nil
}
