package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
	"github.com/pterm/pterm"
)

type transactionModifications struct {
	Amount      *float64
	Description *string
	Date        *time.Time
	Time        *time.Time
	Datetime    *time.Time
	AccountID   *int64
	CategoryID  **int64
	PlaceID     *int64
	GroupID     **int64
	AddTagIDs   []int64
	DropTagIDs  []int64
}

func getOrCreateTagID(cashDb db.DBTX, name string) (int64, error) {
	tag, err := db.GetTagByName(cashDb, name)
	if errors.Is(err, sql.ErrNoRows) {
		return db.InsertTag(cashDb, db.CreateTagInput{Name: name})
	}
	if err != nil {
		return 0, err
	}
	return tag.ID, nil
}

func getExistingTagID(cashDb db.DBTX, name string) (*int64, error) {
	tag, err := db.GetTagByName(cashDb, name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tag.ID, nil
}

func confirmModify(title string, desc string) bool {
	pterm.FgWhite.Println(title)
	pterm.FgWhite.Println("========================")
	pterm.FgWhite.Println(desc)
	ok, err := pterm.DefaultInteractiveConfirm.
		WithDefaultText("Confirm transaction (N/y) ?").
		Show()

	if err != nil {
		panic(err)
	}
	return ok
}

func getModifyTransactionAccount(
	cashDb db.DBTX,
	attributes map[string]parser.AttributeValue,
	config config.Config,
) (string, int64, error) {
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
	return accountName, transactionAccount.ID, nil
}

func getModifyTransactionCategory(cashDb db.DBTX, attributes map[string]parser.AttributeValue) (string, *int64, error) {
	var categoryName string
	var categoryID int64
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
				return "", nil, cErr
			}
			categoryID = newCategoryID
			return categoryName, &categoryID, nil
		} else if err != nil {
			return "", nil, err
		}
		categoryID = category.ID
		return categoryName, nil, nil
	}
	return categoryName, &categoryID, nil
}

func getModifyTransactionStore(cashDb db.DBTX, attributes map[string]parser.AttributeValue) (string, error) {
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
			return storeName, nil
		} else if err != nil {
			// SQL Error
			return "", err
		}
		storeID = sql.NullInt64{Int64: store.ID, Valid: true}
		return storeName, nil
	}
	return "", fmt.Errorf("no store specified")
}

func ModifyTransactions(parsed parser.ParsedCmdLine, cfg config.Config, cashDb db.DBTX) error {
	if parsed.Subcommand != "transactions" {
		panic("ModifyTransactions received an invalid command")
	}

	// Create the transaction filters from arguments
	dbFilters, runFilters, err := createTransactionFilters(parsed, cfg)
	if err != nil {
		return err
	}
	fmt.Printf("dbFilters: %+v\n", dbFilters)
	fmt.Printf("runFilters: %+v\n", runFilters)
	fmt.Printf("parsed: %+v\n", parsed)
	// Parse the modifications from arguments
	modifications, err := parseTransactionModifications(parsed, cfg, cashDb)
	if err != nil {
		return err
	}
	fmt.Printf("modifications: %+v\n", modifications)
	// Get transactions to modify
	transactions, err := db.ListTransactions(cashDb, dbFilters, runFilters)
	if err != nil {
		return err
	}

	// There are no transactions to modify
	if len(transactions) == 0 {
		return fmt.Errorf("no transactions match the given filters")
	}

	// ???
	if err = validateTransactionModificationTargets(transactions, modifications); err != nil {
		return err
	}

	// Check with the user (confirmation)
	confirmed := confirmModify("Modify transactions", fmt.Sprintf("Modifying %d transactions", len(transactions)))
	if !confirmed {
		pterm.Warning.Println("Modification(s) cancelled by user, no transaction added.")
		return nil
	}

	// Modify each transactions
	updatedCount := 0
	for _, transaction := range transactions {
		if err = applyTransactionModifications(cashDb, transaction, modifications); err != nil {
			return err
		}
		updatedCount++
	}

	pterm.Success.Printf("Updated %d transactions\n", updatedCount)
	return nil
}

func Modify(parsed parser.ParsedCmdLine, cfg config.Config, cashDb db.DBTX) error {
	if parsed.Command != "modify" {
		panic("Modify received an invalid command")
	}

	if parsed.Subcommand == "transactions" {
		return ModifyTransactions(parsed, cfg, cashDb)
	}

	return fmt.Errorf("unknown modify subcommand %s", parsed.Subcommand)
}

func parseTransactionModifications(parsed parser.ParsedCmdLine, cfg config.Config, cashDb db.DBTX) (transactionModifications, error) {
	var modifications transactionModifications
	attributes := getAttributes(parsed)
	dateExists := false
	timeExists := false
	datetimeExists := false

	// For each argument, check if it's a tag or a modification'
	for _, arg := range parsed.Args {
		// Tag and negative tag are handled separately
		switch v := arg.(type) {
		case parser.ArgTag:
			if v.Negative {
				tagID, err := getExistingTagID(cashDb, v.Tag)
				if err != nil {
					return transactionModifications{}, err
				}
				if tagID == nil {
					continue
				}
				modifications.DropTagIDs = append(modifications.DropTagIDs, *tagID)
				continue
			} else {
				tagID, err := getOrCreateTagID(cashDb, v.Tag)
				if err != nil {
					return transactionModifications{}, err
				}
				modifications.AddTagIDs = append(modifications.AddTagIDs, tagID)
				continue
			}
		case parser.ArgAttribute:
			switch v.Key {
			case "identifier":
				return transactionModifications{}, fmt.Errorf("identifier is not modifiable")
			case "description":
				var description string
				if v.IsAttributeClear() {
					description = ""
				} else {
					description = v.Value.Value.(parser.StringItem).Value
				}
				modifications.Description = &description
			case "amount":
				if v.IsAttributeClear() {
					return transactionModifications{}, fmt.Errorf("you cannot clear the amount")
				}
				amount := v.Value.Value.(parser.FloatItem).Value
				modifications.Amount = &amount
			case "date":
				dateExists = true
			case "time":
				timeExists = true
			case "datetime":
				datetimeExists = true
			case "account":
				if v.IsAttributeClear() {
					return transactionModifications{}, fmt.Errorf("you cannot clear the account")
				}
				_, accountID, err := getModifyTransactionAccount(cashDb, attributes, cfg)
				if err != nil {
					return transactionModifications{}, err
				}
				modifications.AccountID = &accountID
			case "category":
				if v.IsAttributeClear() {
					var categoryID *int64
					modifications.CategoryID = &categoryID
					continue
				}
				_, categoryID, err := getModifyTransactionCategory(cashDb, attributes)
				if err != nil {
					return transactionModifications{}, err
				}
				if categoryID == nil {
					return transactionModifications{}, fmt.Errorf("no category specified")
				}
				value := categoryID
				modifications.CategoryID = &value
			case "store":
				storeID, err := getTransactionStore(cashDb, attributes)
				if err != nil {
					return transactionModifications{}, err
				}
				if !storeID.Valid {
					return transactionModifications{}, fmt.Errorf("no store specified")
				}
				modifications.PlaceID = &storeID.Int64
			case "group":
				if arg.Kind == parser.TokenAttributeClear {
					var groupID *int64
					modifications.GroupID = &groupID
					continue
				}
				groupID, err := getTransactionGroup(cashDb, attributes)
				if err != nil {
					return transactionModifications{}, err
				}
				if !groupID.Valid {
					return transactionModifications{}, fmt.Errorf("no group specified")
				}
				value := groupID.Int64
				ptr := &value
				modifications.GroupID = &ptr
			default:
				return transactionModifications{}, fmt.Errorf("attribute %s is not modifiable", arg.Key)
			}
		}
	}

	if datetimeExists && (dateExists || timeExists) {
		return transactionModifications{}, fmt.Errorf("datetime and date/time specified, only one allowed")
	}

	modifiedDatetime, err := buildModifiedDatetime(attributes, cfg, datetimeExists, dateExists, timeExists)
	if err != nil {
		return transactionModifications{}, err
	}
	modifications.Date = modifiedDatetime.Date
	modifications.Time = modifiedDatetime.Time
	modifications.Datetime = modifiedDatetime.Datetime

	return modifications, nil
}

type modifiedDatetimeParts struct {
	Date     *time.Time
	Time     *time.Time
	Datetime *time.Time
}

func buildModifiedDatetime(attributes map[string]string, cfg config.Config, datetimeExists bool, dateExists bool, timeExists bool) (modifiedDatetimeParts, error) {
	if !datetimeExists && !dateExists && !timeExists {
		return modifiedDatetimeParts{}, nil
	}

	if datetimeExists {
		datetime, err := time.Parse(cfg.Display.DateFormat, attributes["datetime"])
		if err != nil {
			return modifiedDatetimeParts{}, err
		}
		return modifiedDatetimeParts{Datetime: &datetime}, nil
	}

	parts := modifiedDatetimeParts{}
	if dateExists {
		dateValue, err := parseDateOnly(attributes["date"], cfg)
		if err != nil {
			return modifiedDatetimeParts{}, err
		}
		parts.Date = &dateValue
	}
	if timeExists {
		timeValue, err := parseTimeOnly(attributes["time"], cfg)
		if err != nil {
			return modifiedDatetimeParts{}, err
		}
		parts.Time = &timeValue
	}

	return parts, nil
}

func validateTransactionModificationTargets(transactions []db.Transaction, modifications transactionModifications) error {
	if modifications.AccountID == nil {
		return nil
	}

	for _, transaction := range transactions {
		if transaction.Type == "transfer_in" || transaction.Type == "transfer_out" {
			return fmt.Errorf("account cannot be modified for transfer transactions")
		}
	}

	return nil
}

func applyTransactionModifications(cashDb db.DBTX, transaction db.Transaction, modifications transactionModifications) error {
	if modifications.Amount != nil {
		if err := db.UpdateTransactionAmount(cashDb, transaction.ID, *modifications.Amount); err != nil {
			return err
		}
	}

	if modifications.Description != nil {
		if err := db.UpdateTransactionDescription(cashDb, transaction.ID, *modifications.Description); err != nil {
			return err
		}
	}

	if modifications.Date != nil || modifications.Time != nil || modifications.Datetime != nil {
		updatedDatetime, err := mergeTransactionDatetime(transaction.Datetime, modifications)
		if err != nil {
			return err
		}
		if err := db.UpdateTransactionDatetime(cashDb, transaction.ID, updatedDatetime); err != nil {
			return err
		}
	}

	if modifications.AccountID != nil {
		if err := db.UpdateTransactionAccountID(cashDb, transaction.ID, modifications.AccountID); err != nil {
			return err
		}
	}

	if modifications.CategoryID != nil {
		if err := db.UpdateTransactionCategoryID(cashDb, transaction.ID, *modifications.CategoryID); err != nil {
			return err
		}
	}

	if modifications.PlaceID != nil {
		if err := db.UpdateTransactionPlaceID(cashDb, transaction.ID, modifications.PlaceID); err != nil {
			return err
		}
	}

	if modifications.GroupID != nil {
		if err := db.UpdateTransactionGroupID(cashDb, transaction.ID, *modifications.GroupID); err != nil {
			return err
		}
	}

	for _, tagID := range modifications.AddTagIDs {
		exists, err := db.TransactionTagExists(cashDb, transaction.ID, tagID)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		if err := db.InsertTransactionTag(cashDb, transaction.ID, tagID); err != nil {
			return err
		}
	}

	for _, tagID := range modifications.DropTagIDs {
		if err := db.DeleteTransactionTag(cashDb, transaction.ID, tagID); err != nil {
			return err
		}
	}

	return nil
}

func parseTimeOnly(value string, cfg config.Config) (time.Time, error) {
	parts := strings.Split(cfg.Display.DateFormat, " ")
	if len(parts) > 1 {
		parsed, err := time.Parse(parts[1], value)
		if err == nil {
			return parsed, nil
		}
	}

	for _, layout := range []string{"15:04:05", "15:04"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid time format: %s", value)
}

func mergeTransactionDatetime(original time.Time, modifications transactionModifications) (time.Time, error) {
	if modifications.Datetime != nil {
		return *modifications.Datetime, nil
	}

	updated := original
	if modifications.Date != nil {
		updated = time.Date(
			modifications.Date.Year(),
			modifications.Date.Month(),
			modifications.Date.Day(),
			updated.Hour(),
			updated.Minute(),
			updated.Second(),
			updated.Nanosecond(),
			updated.Location(),
		)
	}
	if modifications.Time != nil {
		updated = time.Date(
			updated.Year(),
			updated.Month(),
			updated.Day(),
			modifications.Time.Hour(),
			modifications.Time.Minute(),
			modifications.Time.Second(),
			modifications.Time.Nanosecond(),
			updated.Location(),
		)
	}

	return updated, nil
}
