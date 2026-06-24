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
	"github.com/pterm/pterm"
)

// transactionModifications groups the modifications that can be applied to a transaction
type transactionModifications struct {
	Amount      *float64
	Description *string
	Date        *time.Time
	AccountID   *int64
	CategoryID  **int64
	PlaceID     *int64
	GroupID     **int64
	AddTagIDs   []int64
	DropTagIDs  []int64
}

// accountModifications groups the modifications that can be applied to an account
type accountModifications struct {
	Name           *string
	Currency       *string
	InitialBalance *float64
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
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &tag.ID, nil
}

func confirmModifyTransaction(cashDb db.DBTX, modification transactionModifications, title string, desc string, config config.Config) bool {
	pterm.FgWhite.Println(title)
	pterm.FgWhite.Println("========================")
	if modification.Amount != nil {
		pterm.FgWhite.Println(fmt.Sprintf("Amount: %f", *modification.Amount))
	}
	if modification.Description != nil {
		pterm.FgWhite.Println(fmt.Sprintf("Description: %s", *modification.Description))
	}
	if modification.Date != nil {
		pterm.FgWhite.Println(fmt.Sprintf("Date: %s", (*modification.Date).Format(config.Display.DateFormat)))
	}
	if modification.AccountID != nil {
		account, err := db.GetAccountByID(cashDb, *modification.AccountID)
		if err != nil {
			panic(err)
		}
		pterm.FgWhite.Println(fmt.Sprintf("Account: %s", account.Name))
	}
	if modification.CategoryID != nil {
		if *modification.CategoryID == nil {
			// Clear category
			pterm.FgWhite.Println("Clear category")
		} else {
			// Category given
			category, err := db.GetCategoryByID(cashDb, **modification.CategoryID)
			if err != nil {
				panic(err)
			}
			pterm.FgWhite.Println(fmt.Sprintf("Category: %s", category.Name))
		}
	}
	if modification.PlaceID != nil {
		store, err := db.GetStoreByID(cashDb, *modification.PlaceID)
		if err != nil {
			panic(err)
		}
		pterm.FgWhite.Println(fmt.Sprintf("Store: %s", store.Name))
	}
	if modification.GroupID != nil {
		group, err := db.GetGroupByID(cashDb, **modification.GroupID)
		if err != nil {
			panic(err)
		}
		pterm.FgWhite.Println(fmt.Sprintf("Group: %s", group.Name))
	}
	if len(modification.AddTagIDs) > 0 {
		for _, tagID := range modification.AddTagIDs {
			tag, err := db.GetTagByID(cashDb, tagID)
			if err != nil {
				panic(err)
			}
			pterm.FgWhite.Println(fmt.Sprintf("Add tag: %s", tag.Name))
		}
	}
	if len(modification.DropTagIDs) > 0 {
		for _, tagID := range modification.DropTagIDs {
			tag, err := db.GetTagByID(cashDb, tagID)
			if err != nil {
				panic(err)
			}
			pterm.FgWhite.Println(fmt.Sprintf("Drop tag: %s", tag.Name))
		}
	}
	pterm.FgWhite.Println(desc)
	ok, err := pterm.DefaultInteractiveConfirm.
		WithDefaultText("Confirm transaction (N/y) ?").
		Show()

	if err != nil {
		panic(err)
	}
	return ok
}

func confirmModifyAccount(
	modification accountModifications,
	title string,
	desc string,
) bool {
	fmt.Printf("modifications: %+v\n", modification)
	pterm.FgWhite.Println(title)
	pterm.FgWhite.Println("========================")
	if modification.Name != nil {
		pterm.FgWhite.Println(fmt.Sprintf("Name: %s", *modification.Name))
	}
	if modification.Currency != nil {
		pterm.FgWhite.Println(fmt.Sprintf("Currency: %s", *modification.Currency))
	}
	if modification.InitialBalance != nil {
		pterm.FgWhite.Println(fmt.Sprintf("Initial balance: %f", *modification.InitialBalance))
	}
	pterm.FgWhite.Println(desc)
	ok, err := pterm.DefaultInteractiveConfirm.
		WithDefaultText("Confirm account modification (N/y) ?").
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
	// Category given
	if categoryAttrValue, ok := attributes["category"]; ok {
		if categoryAttrValue.ValueShape != parser.AttributeValueShapeSingle {
			panic(fmt.Errorf("expected 'category' attribute value to be single valued, got %q", categoryAttrValue.ValueShape))
		}
		var categoryID int64
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
		return categoryName, &categoryID, nil
	}
	return categoryName, nil, nil
}

func getModifyTransactionStore(cashDb db.DBTX, attributes map[string]parser.AttributeValue) (string, *int64, error) {
	var storeID int64
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
				return "", nil, cErr
			}
			storeID = newStoreID
			return storeName, &storeID, nil
		} else if err != nil {
			// SQL Error
			return "", nil, err
		}
		storeID = store.ID
		return storeName, &storeID, nil
	}
	return "", nil, fmt.Errorf("no store specified")
}

func getModifyTransactionGroup(cashDb db.DBTX, attributes map[string]parser.AttributeValue) (string, *int64, error) {
	var groupID int64
	if _, ok := attributes["group"]; ok {
		attrGroup, _ := attributes["group"]
		if attrGroup.ValueShape != parser.AttributeValueShapeSingle {
			panic(fmt.Errorf("expected 'group' attribute value to be single or list valued, got %q", attrGroup.ValueShape))
		}
		groupValue, _ := attrGroup.Value.(parser.StringItem)
		group, err := db.GetGroupByName(cashDb, groupValue.Value)
		if errors.Is(err, sql.ErrNoRows) {
			newGroupID, cErr := db.InsertTransactionGroup(cashDb, db.CreateTransactionGroupInput{Name: groupValue.Value})
			if cErr != nil {
				return "", nil, cErr
			}
			groupID = newGroupID
			return groupValue.Value, &groupID, nil
		} else if err != nil {
			return "", nil, err
		}
		groupID = group.ID
		return groupValue.Value, &groupID, nil
	}
	return "", nil, nil
}

func parseAccountModifications(parsed parser.ParsedCmdLine, cfg config.Config, cashDb db.DBTX) (accountModifications, error) {
	var modifications accountModifications
	for _, arg := range parsed.Args {
		switch v := arg.(type) {
		case parser.ArgAttribute:
			switch v.Key {
			case "name":
				if v.IsAttributeClear() {
					return accountModifications{}, fmt.Errorf("you cannot clear the name of the account")
				}
				name := v.Value.Value.(parser.StringItem).Value
				modifications.Name = &name
			case "currency":
				if v.IsAttributeClear() {
					return accountModifications{}, fmt.Errorf("you cannot clear the currency (only modify)")
				}
				currency := v.Value.Value.(parser.StringItem).Value
				modifications.Currency = &currency
			case "initial-balance":
				var initialAmount float64
				if v.IsAttributeClear() {
					initialAmount = 0
				}
				modifications.InitialBalance = &initialAmount
			default:
				return accountModifications{}, fmt.Errorf("attribute %s is not modifiable", v.Key)
			}
		}
	}
	return modifications, nil
}

func ModifyAccount(parsed parser.ParsedCmdLine, cfg config.Config, cashDb db.DBTX) error {
	if parsed.Subcommand != "accounts" {
		panic("ModifyAccount received an invalid command")
	}

	// Create the account filters from arguments
	dbFilters, runFilters, err := createAccountFilters(parsed, cfg)
	if err != nil {
		return err
	}

	// Parse the modifications from arguments
	modifications, err := parseAccountModifications(parsed, cfg, cashDb)
	if err != nil {
		return err
	}

	// Get accounts to modify
	accounts, err := db.ListAccounts(cashDb, dbFilters, runFilters, []string{})
	if err != nil {
		return err
	}

	// If there are no accounts to modify, return
	if len(accounts) == 0 {
		pterm.Warning.Println("No accounts match the given filters")
		return nil
	}

	// Check with the user (confirmation)
	confirmDesc := fmt.Sprintf("Modifying %d accounts", len(accounts))
	confirmed := confirmModifyAccount(modifications, "Modify accounts", confirmDesc)
	if !confirmed {
		pterm.Warning.Println("Modification(s) cancelled by user, no account added.")
		return nil
	}

	// Modify each account
	updatedCount := 0
	for _, account := range accounts {
		if err = applyAccountModifications(cashDb, account, modifications); err != nil {
			return err
		}
		updatedCount++
	}

	pterm.Success.Printf("Updated %d accounts\n", updatedCount)
	return nil
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

	// Parse the modifications from arguments
	modifications, err := parseTransactionModifications(parsed, cfg, cashDb)
	if err != nil {
		return err
	}

	// Get transactions to modify
	transactions, err := db.ListTransactions(cashDb, dbFilters, runFilters, false)
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
	confirmDesc := fmt.Sprintf("Modifying %d transactions", len(transactions))
	confirmed := confirmModifyTransaction(cashDb, modifications, "Modify transactions", confirmDesc, cfg)
	if !confirmed {
		pterm.Warning.Println("Modification(s) cancelled by user, no transaction added.")
		return nil
	}

	// Modify each transaction
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
	} else if parsed.Subcommand == "accounts" {
		return ModifyAccount(parsed, cfg, cashDb)
	}

	return fmt.Errorf("unknown modify subcommand %s", parsed.Subcommand)
}

func parseTransactionModifications(parsed parser.ParsedCmdLine, cfg config.Config, cashDb db.DBTX) (transactionModifications, error) {
	var modifications transactionModifications
	attributes := getAttributes(parsed)
	dateExists := false

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
				if v.IsAttributeClear() {
					return transactionModifications{}, fmt.Errorf("you cannot clear the store (only modify)")
				}
				_, storeID, err := getModifyTransactionStore(cashDb, attributes)
				if err != nil {
					return transactionModifications{}, err
				}
				if storeID == nil {
					return transactionModifications{}, fmt.Errorf("no store specified")
				}
				modifications.PlaceID = storeID
			case "group":
				if v.IsAttributeClear() {
					var groupID *int64
					modifications.GroupID = &groupID
					continue
				}
				_, groupID, err := getModifyTransactionGroup(cashDb, attributes)
				if err != nil {
					return transactionModifications{}, err
				}
				if groupID == nil {
					return transactionModifications{}, fmt.Errorf("no group specified")
				}
				value := groupID
				modifications.GroupID = &value
			default:
				return transactionModifications{}, fmt.Errorf("attribute %s is not modifiable", v.Key)
			}
		}
	}

	newDate, err := buildModifiedDatetime(attributes, dateExists)
	if err != nil {
		return transactionModifications{}, err
	}
	modifications.Date = newDate

	return modifications, nil
}

type modifiedDatetimeParts struct {
	Date     *time.Time
	Time     *time.Time
	Datetime *time.Time
}

func getModifiedDate(dateAttrValue parser.AttributeValue) (time.Time, error) {
	// We accept only single and sshortcut
	if dateAttrValue.ValueShape != parser.AttributeValueShapeSingle && dateAttrValue.ValueShape != parser.AttributeValueShapeShortcut {
		panic(
			fmt.Sprintf("expected 'date' attribute value to be single valued, got %q", dateAttrValue.ValueShape),
		)
	}
	if dateAttrValue.ValueShape == parser.AttributeValueShapeSingle {
		dateValue, _ := dateAttrValue.Value.(parser.TimeItem)
		return time.Date(dateValue.Value.Year(), dateValue.Value.Month(), dateValue.Value.Day(), 0, 0, 0, 0, dateValue.Value.Location()), nil
	} else if dateAttrValue.ValueShape == parser.AttributeValueShapeShortcut {
		if !validateShortcut(dateAttrValue.Shortcut.Name) {
			return time.Time{}, fmt.Errorf("invalid shortcut %q, accepted: %v", dateAttrValue.Shortcut.Name, acceptedShortcuts)
		}
		dateValue, err := domain.GetTimeShortcutAsSingle(dateAttrValue.Shortcut.Name)
		if err != nil {
			return time.Time{}, err
		}
		return time.Date(dateValue.Year(), dateValue.Month(), dateValue.Day(), 0, 0, 0, 0, dateValue.Location()), nil
	}
	panic(
		fmt.Sprintf("expected 'date' attribute value to be single valued or a shortcut, got %q", dateAttrValue.ValueShape),
	)
}

func buildModifiedDatetime(
	attributes map[string]parser.AttributeValue,
	dateExists bool,
) (*time.Time, error) {
	if !dateExists {
		return nil, nil
	}
	newDate, err := getModifiedDate(attributes["date"])
	if err != nil {
		return nil, err
	}
	return &newDate, nil
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

func applyAccountModifications(cashDb db.DBTX, account db.Account, modifications accountModifications) error {
	// Update account name
	if modifications.Name != nil {
		if err := db.UpdateAccountName(cashDb, account.ID, *modifications.Name); err != nil {
			return err
		}
	}

	// Update account currency
	if modifications.Currency != nil {
		if err := db.UpdateAccountCurrency(cashDb, account.ID, *modifications.Currency); err != nil {
			return err
		}
	}

	// Update account initial-balance
	if modifications.InitialBalance != nil {
		if err := db.UpdateAccountInitialBalance(cashDb, account.ID, *modifications.InitialBalance); err != nil {
			return err
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

	if modifications.Date != nil {
		if err := db.UpdateTransactionDatetime(cashDb, transaction.ID, *modifications.Date); err != nil {
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
