package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

type listSortOptions struct {
	Field string
	Desc  bool
}

func defaultListSortOptions() listSortOptions {
	return listSortOptions{Desc: true}
}

func printTransactionTable(cashDb db.DBTX, transactions []db.Transaction, config config.Config) error {
	rows := make([][]string, 0, len(transactions))
	types := make([]string, 0, len(transactions))

	for _, transaction := range transactions {
		var categoryName string
		if transaction.CategoryID != nil {
			category, err := transaction.GetCategory(cashDb)
			if err != nil {
				return err
			}
			categoryName = category.Name
		} else {
			categoryName = "none"
		}
		groupName := ""
		if transaction.GroupID != nil {
			group, err := transaction.GetGroup(cashDb)
			if err != nil {
				return err
			}
			if group != nil {
				groupName = group.Name
			}
		}
		var vendorName string
		vendor, err := transaction.GetPlace(cashDb)
		if err != nil {
			return err
		}
		vendorName = vendor.Name
		rows = append(rows, []string{
			transaction.Identifier,
			transaction.AccountName,
			strconv.FormatFloat(transaction.Amount, 'f', 2, 64),
			transaction.Currency,
			vendorName,
			transaction.Description,
			transaction.Datetime.Format("2006-01-02"),
			categoryName,
			groupName,
		})
		types = append(types, transaction.Type)
	}

	theme := gui.CurrentTheme()

	t := gui.NewTable().
		WithTitle("Transactions", theme.TransactionListTitleBackground).
		WithSubtitle("Configured transactions").
		WithHeaderBackground(theme.TransactionListHeaderBackground).
		WithHeaders("ID", "Account", "Amount", "Currency", "Vendor", "Description", "Date", "Category", "Group")

	for i, row := range rows {
		t.AddRowWithMetadata(row, map[string]string{"type": types[i]})
	}

	fmt.Println(t.Render())
	fmt.Printf(" Returned %d transactions\n\n\n", len(transactions))

	return nil
}

func PrintExpensesIncomeByCurrency(cashDb db.DBTX, transactions []db.Transaction) error {
	type currencySummary struct {
		Income   float64
		Expenses float64
	}

	type accountSummary struct {
		Name     string
		Currency string
		Income   float64
		Expenses float64
	}

	accounts, err := db.ListAccounts(cashDb, db.AccountListFilter{})
	if err != nil {
		return err
	}

	accountByID := make(map[int64]db.Account, len(accounts))
	accountTotals := make(map[int64]*accountSummary, len(accounts))
	for _, account := range accounts {
		accountByID[account.ID] = account
		accountTotals[account.ID] = &accountSummary{
			Name:     account.Name,
			Currency: account.Currency,
		}
	}

	currencyTotals := make(map[string]*currencySummary)
	for _, transaction := range transactions {
		if transaction.Type != "income" && transaction.Type != "expense" {
			continue
		}

		if transaction.AccountID == nil {
			continue
		}

		account, ok := accountByID[*transaction.AccountID]
		if !ok {
			continue
		}

		currencyTotal, ok := currencyTotals[account.Currency]
		if !ok {
			currencyTotal = &currencySummary{}
			currencyTotals[account.Currency] = currencyTotal
		}

		accountTotal := accountTotals[account.ID]
		if transaction.Amount >= 0 {
			currencyTotal.Income += transaction.Amount
			accountTotal.Income += transaction.Amount
		} else {
			currencyTotal.Expenses += transaction.Amount
			accountTotal.Expenses += transaction.Amount
		}
	}

	currencies := make([]string, 0, len(currencyTotals))
	for currency := range currencyTotals {
		currencies = append(currencies, currency)
	}
	sort.Strings(currencies)

	formatAmount := func(amount float64, forceSign bool) string {
		if forceSign {
			return fmt.Sprintf("%+.2f", amount)
		}
		if amount == 0 {
			return "0.00"
		}
		return fmt.Sprintf("%.2f", amount)
	}

	headers := []string{""}
	incomeRow := []string{"Income"}
	expensesRow := []string{"Expenses"}
	netRow := []string{"Net"}
	for _, currency := range currencies {
		totals := currencyTotals[currency]
		headers = append(headers, currency)
		incomeRow = append(incomeRow, fmt.Sprintf("%s %s", formatAmount(totals.Income, true), currency))
		expensesRow = append(expensesRow, fmt.Sprintf("%s %s", formatAmount(totals.Expenses, false), currency))
		netRow = append(netRow, fmt.Sprintf("%s %s", formatAmount(totals.Income+totals.Expenses, true), currency))
	}

	theme := gui.CurrentTheme()
	summaryTable := gui.NewTable().
		WithType(gui.TableTypeSummary).
		WithTitle(" Transaction summary", theme.TransactionListTitleBackground).
		WithHeaderBackground(theme.TransactionListHeaderBackground).
		WithHeaders(headers...).
		AddRow(incomeRow...).
		AddRow(expensesRow...).
		AddRow(netRow...)

	fmt.Println(summaryTable.Render())
	fmt.Println()

	accountRows := make([][]string, 0, len(accountTotals))
	accountKeys := make([]int64, 0, len(accountTotals))
	for accountID := range accountTotals {
		accountKeys = append(accountKeys, accountID)
	}
	sort.Slice(accountKeys, func(i, j int) bool {
		return accountTotals[accountKeys[i]].Name < accountTotals[accountKeys[j]].Name
	})

	for _, accountID := range accountKeys {
		totals := accountTotals[accountID]
		accountRows = append(accountRows, []string{
			totals.Name,
			formatAmount(totals.Income, true),
			formatAmount(totals.Expenses, false),
			formatAmount(totals.Income+totals.Expenses, true),
		})
	}

	accountTable := gui.NewTable().
		WithType(gui.TableTypeSummary).
		WithTitle(" By account", theme.TransactionListTitleBackground).
		WithHeaderBackground(theme.TransactionListHeaderBackground).
		WithHeaders("Account", "Income", "Expenses", "Net").
		AddRows(accountRows)

	fmt.Println(accountTable.Render())
	fmt.Println()
	fmt.Println()

	return nil
}

func parseListSortOptions(parsed parser.ParsedCmdLine) (parser.ParsedCmdLine, listSortOptions, error) {
	sortOptions := defaultListSortOptions()
	filteredFilters := make([]parser.Token, 0, len(parsed.Filters))
	filteredArgs := make([]parser.Token, 0, len(parsed.Args))

	consumeSortAttribute := func(token parser.Token) (bool, error) {
		if token.Kind != parser.TokenAttribute {
			return false, nil
		}

		switch token.Attribute.Key {
		case "order":
			if sortOptions.Field != "" {
				return false, fmt.Errorf("order specified multiple times")
			}
			if !isSupportedListOrderField(token.Attribute.Value.Raw) {
				return false, fmt.Errorf("unsupported order field %s", token.Attribute.Value)
			}
			sortOptions.Field = token.Attribute.Value.Raw
			return true, nil
		case "desc":
			desc, err := strconv.ParseBool(token.Attribute.Value.Raw)
			if err != nil {
				return false, fmt.Errorf("invalid desc value %s", token.Attribute.Value)
			}
			sortOptions.Desc = desc
			return true, nil
		default:
			return false, nil
		}
	}

	for _, filter := range parsed.Filters {
		consumed, err := consumeSortAttribute(filter)
		if err != nil {
			return parsed, sortOptions, err
		}
		if consumed {
			continue
		}
		filteredFilters = append(filteredFilters, filter)
	}

	for _, arg := range parsed.Args {
		consumed, err := consumeSortAttribute(arg)
		if err != nil {
			return parsed, sortOptions, err
		}
		if consumed {
			continue
		}
		filteredArgs = append(filteredArgs, arg)
	}

	parsed.Filters = filteredFilters
	parsed.Args = filteredArgs
	return parsed, sortOptions, nil
}

func isSupportedListOrderField(field string) bool {
	switch field {
	case "id", "datetime", "description", "amount", "account", "currency", "type", "category", "vendor":
		return true
	case "date":
		return true
	default:
		return false
	}
}

func sortTransactionsForList(cashDb db.DBTX, transactions []db.Transaction, options listSortOptions) error {
	if options.Field == "" {
		return nil
	}

	type sortValues struct {
		category string
		vendor   string
	}

	valuesByID := make(map[int64]sortValues, len(transactions))
	for _, transaction := range transactions {
		values := sortValues{}
		if options.Field == "category" && transaction.CategoryID != nil {
			category, err := transaction.GetCategory(cashDb)
			if err != nil {
				return err
			}
			values.category = category.Name
		}
		if options.Field == "vendor" && transaction.PlaceID != nil {
			vendor, err := transaction.GetPlace(cashDb)
			if err != nil {
				return err
			}
			if vendor != nil {
				values.vendor = vendor.Name
			}
		}
		valuesByID[transaction.ID] = values
	}

	compareText := func(left string, right string) int {
		return strings.Compare(strings.ToLower(left), strings.ToLower(right))
	}

	sort.SliceStable(transactions, func(i, j int) bool {
		left := transactions[i]
		right := transactions[j]

		cmp := 0
		switch options.Field {
		case "id":
			cmp = compareOrdered(left.ID, right.ID)
		case "datetime", "date":
			cmp = compareTime(left.Datetime, right.Datetime)
		case "description":
			cmp = compareText(left.Description, right.Description)
		case "amount":
			cmp = compareOrdered(left.Amount, right.Amount)
		case "account":
			cmp = compareText(left.AccountName, right.AccountName)
		case "currency":
			cmp = compareText(left.Currency, right.Currency)
		case "type":
			cmp = compareText(left.Type, right.Type)
		case "category":
			cmp = compareText(valuesByID[left.ID].category, valuesByID[right.ID].category)
		case "vendor":
			cmp = compareText(valuesByID[left.ID].vendor, valuesByID[right.ID].vendor)
		}

		if cmp == 0 {
			cmp = compareOrdered(left.ID, right.ID)
		}

		if options.Desc {
			return cmp > 0
		}
		return cmp < 0
	})

	return nil
}

func compareOrdered[T ~int64 | ~float64](left T, right T) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}

func compareTime(left time.Time, right time.Time) int {
	if left.Before(right) {
		return -1
	}
	if left.After(right) {
		return 1
	}
	return 0
}

func List(parsed parser.ParsedCmdLine, config config.Config, cashDb db.DBTX) error {
	parsed, sortOptions, err := parseListSortOptions(parsed)
	if err != nil {
		return err
	}

	dbFilters, runFilters, err := createTransactionFilters(parsed, config)
	if err != nil {
		return err
	}

	// Get list of transactions
	transactions, err := db.ListTransactions(cashDb, dbFilters, runFilters)
	if err != nil {
		return err
	}

	err = sortTransactionsForList(cashDb, transactions, sortOptions)
	if err != nil {
		return err
	}

	// Print transaction table
	err = printTransactionTable(cashDb, transactions, config)
	if err != nil {
		return err
	}

	//  Print expenses and income summary
	err = PrintExpensesIncomeByCurrency(cashDb, transactions)
	if err != nil {
		return err
	}

	return nil
}

func listDeletedTransactions(config config.Config, cashDb db.DBTX) error {
	transactions, err := db.ListDeletedTransactions(cashDb, nil, nil)
	if err != nil {
		return err
	}

	if err := printTransactionTable(cashDb, transactions, config); err != nil {
		return err
	}

	return nil
}
