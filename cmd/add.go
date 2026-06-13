package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/domain"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

var CommandValidations = []func(
	parser.ParsedCmdLine,
	config.Config,
	db.DBTX,
	map[parser.ArgKind]int,
) (parser.ParsedCmdLine, error){
	validateAmount,
}

func validateAmount(
	parsed parser.ParsedCmdLine,
	config config.Config,
	db db.DBTX,
	counts map[parser.ArgKind]int,
) (parser.ParsedCmdLine, error) {
	// Zero or multiple amounts => problem
	//if counts[parser.TokenAmount] > 1 {
	//	name, err := pterm.DefaultInteractiveTextInput.
	//		WithDefaultText("y").
	//		Show("You specified multiple amounts. Would you like to sum them up?:")
	//	if err != nil {
	//		return parsed, err
	//	}
	//	if name != "y" {
	//		return parsed, fmt.Errorf("multiple amounts specified, only one allowed if not summed up")
	//	}
	//
	//	// Sum up amounts
	//	var sumAmount float32
	//	for _, amount := range parsed.GetAmounts() {
	//		sumAmount += amount.Amount
	//	}
	//
	//	// No zero amount allowed
	//	if sumAmount == 0 {
	//		return parsed, fmt.Errorf("summed up amount is zero")
	//	}
	//
	//	// Remove amounts from parsed command line
	//	parsed.RemoveByKind(parser.TokenAmount)
	//
	//	// Add summed up amount
	//	parsed.Append(parser.Token{
	//		Raw:    strconv.FormatFloat(float64(sumAmount), 'f', 2, 32),
	//		Amount: sumAmount,
	//		Kind:   parser.TokenAmount,
	//	}, false)
	//
	//	return parsed, nil
	//}

	// We have an amount, check if it's valid
	//amount := parsed.GetAmounts()[0]

	// No zero amount allowed
	//if amount.Amount == 0 {
	//	return parsed, fmt.Errorf("amount cannot be zero")
	//}

	//return parsed, nil
	return parser.ParsedCmdLine{}, nil
}

func runCommandLineValidation(parsed parser.ParsedCmdLine, config config.Config, db db.DBTX, counts map[parser.ArgKind]int) (parser.ParsedCmdLine, error) {
	for _, check := range CommandValidations {
		var err error
		parsed, err = check(parsed, config, db, counts)
		if err != nil {
			return parsed, err
		}
	}
	return parsed, nil
}

func getTransactionDescription(parsed parser.ParsedCmdLine, counts map[parser.ArgKind]int) string {
	_ = counts
	textParts := make([]string, 0)
	for _, arg := range parsed.Args {
		text, ok := arg.(parser.ArgText)
		if !ok {
			continue
		}
		textParts = append(textParts, text.Text)
	}
	return strings.Join(textParts, " ")
}

func getNextIdentifier(cashDb db.DBTX, transactionTime time.Time) (domain.TransactionID, error) {
	lastNumber, err := getMaxTransactionNumberForMonth(cashDb, transactionTime.Year(), int(transactionTime.Month()))
	if err != nil {
		return domain.TransactionID{}, err
	}

	return domain.TransactionID{
		Year:  transactionTime.Year(),
		Month: int(transactionTime.Month()),
		Num:   lastNumber + 1,
	}, nil
}

func getTransactionAmount(parsed parser.ParsedCmdLine, counts map[parser.ArgKind]int) float32 {
	//return parsed.GetAmounts()[0].Amount
	return 0.0
}

func getTransactionDatetime(
	attributes map[string]string,
	config config.Config,
) (time.Time, error) {
	toDateOnly := func(dt time.Time) time.Time {
		return time.Date(dt.Year(), dt.Month(), dt.Day(), 0, 0, 0, 0, dt.Location())
	}
	parseFlexibleDate := func(value string) (time.Time, error) {
		value = strings.TrimSpace(value)
		layouts := []string{
			"2006-01-02",
			"2006-01-02 15:04:05",
			"2006-01-02 15:04",
			"02.01.2006",
			"02.01.2006 15:04:05",
			"02.01.2006 15:04",
			"02/01/2006",
			"02/01/2006 15:04:05",
			"02/01/2006 15:04",
			config.Display.DateFormat,
		}
		seen := make(map[string]bool, len(layouts))
		for _, layout := range layouts {
			if layout == "" || seen[layout] {
				continue
			}
			seen[layout] = true
			parsed, err := time.Parse(layout, value)
			if err == nil {
				return parsed, nil
			}
		}
		return time.Time{}, fmt.Errorf("unsupported datetime format %q", value)
	}
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
		datetime, err := parseFlexibleDate(attributes["datetime"])
		if err != nil {
			return time.Time{}, err
		}
		return toDateOnly(datetime), nil
	}

	// Date given (time is ignored on purpose)
	if dateExists {
		datetime, err := parseFlexibleDate(attributes["date"])
		if err != nil {
			return time.Time{}, err
		}
		return toDateOnly(datetime), nil
	}

	if timeExists {
		now := time.Now()
		return toDateOnly(now), nil
	}

	return toDateOnly(time.Now()), nil
}

func getAttributes(parsed parser.ParsedCmdLine) map[string]parser.AttributeValue {
	var attributes map[string]parser.AttributeValue = make(map[string]parser.AttributeValue)
	for _, arg := range parsed.Args {
		attr, ok := arg.(parser.ArgAttribute)
		if !ok {
			continue
		}
		attributes[attr.Key] = attr.Value
	}
	return attributes
}

func getTransactionStore(cashDb db.DBTX, attributes map[string]string) (sql.NullInt64, error) {
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

func getTransactionAccount(cashDb db.DBTX, attributes map[string]string, config config.Config) (int64, error) {
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

func getTransactionCategory(cashDb db.DBTX, attributes map[string]string) (sql.NullInt64, error) {
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

func getTransactionGroup(cashDb db.DBTX, attributes map[string]string) (sql.NullInt64, error) {
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

func Add(parsed parser.ParsedCmdLine, config config.Config, cashDb db.DBTX) error {
	// Get count by token kind for validation
	tokenKindsCount := parsed.GetTokenKindCount(false)

	// We check the commnand line first
	parsed, err := runCommandLineValidation(parsed, config, cashDb, tokenKindsCount)
	if err != nil {
		return err
	}

	// Get attributes
	// attributes := getAttributes(parsed)

	// Get transaction description (merge it if necessary) & amount
	// desc := getTransactionDescription(parsed, tokenKindsCount)
	// amount := getTransactionAmount(parsed, tokenKindsCount)

	// Get transaction datetime
	//transactionTime, err := getTransactionDatetime(attributes, config)
	//if err != nil {
	//	return err
	//}
	//
	//// Get next transaction identifier for the transaction month
	//nextIdentifier, err := getNextIdentifier(cashDb, transactionTime)
	//if err != nil {
	//	return err
	//}
	//
	//// Transaction type
	//var transactionType string
	//if amount < 0.0 {
	//	transactionType = "expense"
	//} else {
	//	transactionType = "income"
	//}
	//
	//// Get store
	//transactionStore, err := getTransactionStore(cashDb, attributes)
	//if err != nil {
	//	return err
	//}
	//
	//// Get account
	//transactionAccount, err := getTransactionAccount(cashDb, attributes, config)
	//if err != nil {
	//	return err
	//}
	//
	//// Get category
	//transactionCategory, err := getTransactionCategory(cashDb, attributes)
	//if err != nil {
	//	return err
	//}
	//
	//// Get group
	//transactionGroup, err := getTransactionGroup(cashDb, attributes)
	//if err != nil {
	//	return err
	//}
	//
	//// Insert transaction
	//transactionID, err := db.InsertTransaction(
	//	cashDb,
	//	db.CreateTransactionInput{
	//		Identifier:  fmt.Sprintf("%s", nextIdentifier),
	//		Type:        transactionType,
	//		Amount:      float64(amount),
	//		Description: desc,
	//		Datetime:    transactionTime,
	//		AccountID:   transactionAccount,
	//		CategoryID:  db.NullInt64ToPtr(transactionCategory),
	//		PlaceID:     db.NullInt64ToPtr(transactionStore),
	//		GroupID:     db.NullInt64ToPtr(transactionGroup),
	//	},
	//)
	//if err != nil {
	//	return err
	//}

	// Show success message
	// pterm.Success.Println("Transaction added with id: " + strconv.FormatInt(transactionID, 10) + "")

	return nil
}
