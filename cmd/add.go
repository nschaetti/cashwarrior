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

// CommandValidations List of command line validator
var CommandValidations = []func(parser.ParsedCmdLine, config.Config, db.DBTX, map[parser.ArgKind]int) (parser.ParsedCmdLine, error){
	validateAmount,
}

func validateAmount(
	parsed parser.ParsedCmdLine,
	config config.Config,
	db db.DBTX,
	counts map[parser.ArgKind]int,
) (parser.ParsedCmdLine, error) {
	return parsed, nil
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

// getTransactionDescription Get the text tokens for the description of the transaction
func getTransactionDescription(addInput *db.CreateTransactionInput, parsed parser.ParsedCmdLine, counts map[parser.ArgKind]int) string {
	_ = counts
	textParts := make([]string, 0)
	for _, arg := range parsed.Args {
		text, ok := arg.(parser.ArgText)
		if !ok {
			continue
		}
		textParts = append(textParts, text.Text)
	}
	addInput.Description = strings.Join(textParts, " ")
	return addInput.Description
}

func getNextIdentifier(addInput *db.CreateTransactionInput, cashDb db.DBTX, transactionTime time.Time) (domain.TransactionID, error) {
	lastNumber, err := getMaxTransactionNumberForMonth(cashDb, transactionTime.Year(), int(transactionTime.Month()))
	if err != nil {
		return domain.TransactionID{}, err
	}

	nextIdentifier := domain.TransactionID{
		Year:  transactionTime.Year(),
		Month: int(transactionTime.Month()),
		Num:   lastNumber + 1,
	}

	addInput.Identifier = fmt.Sprintf("%s", nextIdentifier)

	return nextIdentifier, nil
}

func getTransactionAmount(addInput *db.CreateTransactionInput, attributes map[string]parser.AttributeValue) float64 {
	if amountAttrValue, ok := attributes["amount"]; ok {
		// Check single value
		if amountAttrValue.ValueShape != parser.AttributeValueShapeSingle {
			panic(
				fmt.Sprintf(
					"Amount must be single valued for add %v. Expected %v.",
					amountAttrValue.ValueShape,
					parser.AttributeValueShapeSingle,
				),
			)
		}
		amountValue, _ := amountAttrValue.Value.(parser.FloatItem)
		addInput.Amount = amountValue.Value
		return amountValue.Value
	}
	panic(fmt.Sprintf("No amount attribute found for add %v", addInput.Identifier))
	return 0.0
}

func getTransactionDatetime(
	addInput *db.CreateTransactionInput,
	attributes map[string]parser.AttributeValue,
	config config.Config,
) (time.Time, error) {
	toDateOnly := func(dt time.Time) time.Time {
		return time.Date(dt.Year(), dt.Month(), dt.Day(), 0, 0, 0, 0, dt.Location())
	}

	// Check attributes exists
	dateAttrValue, dateExists := attributes["date"]

	// Date given (time is ignored on purpose)
	if dateExists {
		// Check single
		if dateAttrValue.ValueShape != parser.AttributeValueShapeSingle {
			panic(
				fmt.Sprintf("expected 'date' attribute value to be single valued, got %q", dateAttrValue.ValueShape),
			)
		}
		dateValue, _ := dateAttrValue.Value.(parser.TimeItem)
		addInput.Date = toDateOnly(dateValue.Value)
		return addInput.Date, nil
	}

	// If nothing specified, take now, just the date
	addInput.Date = toDateOnly(time.Now())
	return addInput.Date, nil
}

// getAttributes Get attributes by name with their values
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

func getTransactionStore(addInput *db.CreateTransactionInput, cashDb db.DBTX, attributes map[string]parser.AttributeValue) (string, error) {
	var storeID sql.NullInt64
	var storeName string
	if storeAttrValue, ok := attributes["store"]; ok {
		if storeAttrValue.ValueShape != parser.AttributeValueShapeSingle {
			panic(fmt.Errorf("expected 'store' attribute value to be single valued, got %q", storeAttrValue.ValueShape))
		}
		storeValue, _ := storeAttrValue.Value.(parser.StringItem)
		storeName = storeValue.Value
		store, err := db.GetStoreByName(cashDb, storeName)
		if errors.Is(err, sql.ErrNoRows) {
			// Create new store
			newStoreID, cErr := db.InsertStore(cashDb, db.CreatePlaceInput{Name: storeName})
			if cErr != nil {
				return "", cErr
			}
			storeID = sql.NullInt64{Int64: newStoreID, Valid: true}
			addInput.PlaceID = db.NullInt64ToPtr(storeID)
			return storeName, nil
		} else if err != nil {
			// SQL Error
			return "", err
		}
		storeID = sql.NullInt64{Int64: store.ID, Valid: true}
	} else {
		return "", fmt.Errorf("no store specified")
	}
	return storeName, nil
}

func getTransactionAccount(addInput *db.CreateTransactionInput, cashDb db.DBTX, attributes map[string]parser.AttributeValue, config config.Config) (string, int64, error) {
	var accountName string
	if accountAttrValue, ok := attributes["account"]; ok {
		if accountAttrValue.ValueShape != parser.AttributeValueShapeSingle {
			panic(fmt.Errorf("expected 'account' attribute value to be single, got %q", accountAttrValue.ValueShape))
		}
		accountValue, _ := accountAttrValue.Value.(parser.StringItem)
		accountName = accountValue.Value
	} else {
		accountName = config.Default.Account
	}
	// Get account
	transactionAccount, err := db.GetAccountByName(cashDb, accountName)
	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, fmt.Errorf("account %s does not exist, create it first", accountName)
	} else if err != nil {
		return "", 0, err
	}
	addInput.AccountID = transactionAccount.ID
	return accountName, transactionAccount.ID, nil
}

func getTransactionCategory(addInput *db.CreateTransactionInput, cashDb db.DBTX, attributes map[string]parser.AttributeValue) (string, error) {
	var categoryName string
	// Category given
	if categoryAttrValue, ok := attributes["category"]; ok {
		if categoryAttrValue.ValueShape != parser.AttributeValueShapeSingle {
			panic(fmt.Errorf("expected 'category' attribute value to be single valued, got %q", categoryAttrValue.ValueShape))
		}
		categoryValue, _ := categoryAttrValue.Value.(parser.StringItem)
		categoryName = categoryValue.Value
		category, err := db.GetCategoryByName(cashDb, categoryName)
		if errors.Is(err, sql.ErrNoRows) {
			// Create new category with root as parent
			parentID := int64(1)
			newCategoryID, cErr := db.InsertCategory(
				cashDb,
				db.CreateCategoryInput{Name: categoryName, ParentID: &parentID},
			)
			if cErr != nil {
				return "", cErr
			}
			addInput.CategoryID = db.NullInt64ToPtr(sql.NullInt64{Int64: newCategoryID, Valid: true})
			return categoryName, nil
		} else if err != nil {
			return "", err
		}
		addInput.CategoryID = db.NullInt64ToPtr(sql.NullInt64{Int64: category.ID, Valid: true})
		return categoryName, nil
	}
	return categoryName, nil
}

func getTransactionGroup(addInput *db.CreateTransactionInput, cashDb db.DBTX, attributes map[string]parser.AttributeValue) (string, error) {
	if _, ok := attributes["group"]; ok {
		attrGroup, _ := attributes["group"]
		if attrGroup.ValueShape != parser.AttributeValueShapeSingle {
			panic(fmt.Errorf("expected 'group' attribute value to be single or list valued, got %q", attrGroup.ValueShape))
		}
		groupValue, _ := attrGroup.Value.(parser.StringItem)
		group, err := db.GetTransactionGroupByName(cashDb, groupValue.Value)
		if errors.Is(err, sql.ErrNoRows) {
			newGroupID, cErr := db.InsertTransactionGroup(cashDb, db.CreateTransactionGroupInput{Name: groupValue.Value})
			if cErr != nil {
				return "", cErr
			}
			addInput.GroupID = db.NullInt64ToPtr(sql.NullInt64{Int64: newGroupID, Valid: true})
			return groupValue.Value, nil
		} else if err != nil {
			return "", err
		}
		addInput.GroupID = db.NullInt64ToPtr(sql.NullInt64{Int64: group.ID, Valid: true})
		return groupValue.Value, nil
	}
	return "", nil
}

type AddTransactionInput struct {
	Identifier domain.TransactionID
	Amount     float64
	Desc       string
	Date       time.Time
	Type       string
	StoreID    sql.NullInt64
	CategoryID sql.NullInt64
	GroupID    sql.NullInt64
}

func Add(parsed parser.ParsedCmdLine, config config.Config, cashDb db.DBTX) error {
	// Get count by token kind for validation
	tokenKindsCount := parsed.GetTokenKindCount(false)

	// Get command, and subcommand specs
	// commandSpec, _ = parser.GetCommandSpec(parsed.Command)
	// subcommandSpec, _ = parser.GetSubcommandSpec(parsed.Subcommand)

	// We check the commnand line first
	parsed, err := runCommandLineValidation(parsed, config, cashDb, tokenKindsCount)
	if err != nil {
		return err
	}

	// Get attributes
	attributes := getAttributes(parsed)

	// Define what's needed to add a transaction
	addInput := db.CreateTransactionInput{}

	// Get transaction description (text tokens, merge it if necessary) & amount
	desc := getTransactionDescription(&addInput, parsed, tokenKindsCount)
	amount := getTransactionAmount(&addInput, attributes)

	// Get transaction datetime
	transactionDate, err := getTransactionDatetime(&addInput, attributes, config)
	if err != nil {
		return err
	}

	// Get next transaction identifier for the transaction month
	nextIdentifier, err := getNextIdentifier(&addInput, cashDb, transactionDate)
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

	// Get store
	storeSqlID, err := getTransactionStore(&addInput, cashDb, attributes)
	if err != nil {
		return err
	}

	// Get account
	transactionAccount, _, err := getTransactionAccount(&addInput, cashDb, attributes, config)
	if err != nil {
		return err
	}

	// Get category
	transactionCategory, err := getTransactionCategory(&addInput, cashDb, attributes)
	if err != nil {
		return err
	}

	// Get group
	transactionGroup, err := getTransactionGroup(&addInput, cashDb, attributes)
	if err != nil {
		return err
	}

	fmt.Printf("Amount: %f \n", amount)
	fmt.Printf("Description: %v \n", desc)
	fmt.Printf("Transaction Time: %v \n", transactionDate)
	fmt.Printf("NextIdentifier: %v \n", nextIdentifier)
	fmt.Printf("Transaction Type: %v \n", transactionType)
	fmt.Printf("Store SQL ID: %v \n", storeSqlID)
	fmt.Printf("Transaction Account: %v \n", transactionAccount)
	fmt.Printf("Transaction Category: %v \n", transactionCategory)
	fmt.Printf("Transaction Group: %v \n", transactionGroup)

	//// Insert transaction
	//transactionID, err := db.InsertTransaction(
	//	cashDb,
	//	db.CreateTransactionInput{
	//		Identifier:  fmt.Sprintf("%s", nextIdentifier),
	//		Type:        transactionType,
	//		Amount:      float64(amount),
	//		Description: desc,
	//		Date:        transactionTime,
	//		AccountID:   transactionAccount,
	//		CategoryID:  db.NullInt64ToPtr(transactionCategory),
	//		PlaceID:     db.NullInt64ToPtr(transactionStore),
	//		GroupID:     db.NullInt64ToPtr(transactionGroup),
	//	},
	//)
	//if err != nil {
	//	return err
	//}
	//
	//// Show success message
	//pterm.Success.Println("Transaction added with id: " + strconv.FormatInt(transactionID, 10) + "")

	return nil
}
