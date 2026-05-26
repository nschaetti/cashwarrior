package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/domain"
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

const (
	FilterUnknown = iota
	FilterTypeTransactionID
	FilterTypeAccountName
	FilterTypeCurrency
	FilterTypeStore
	FilterTypeDescription
	FilterTypeDatetime
)

func printTransactionTable(cashDb db.DBTX, transactions []db.Transaction, config config.Config) error {
	rows := make([][]string, 0, len(transactions))
	types := make([]string, 0, len(transactions))

	for _, transaction := range transactions {
		account, err := transaction.GetAccount(cashDb)
		if err != nil {
			return err
		}
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
			transaction.Datetime.Format(config.Display.DateFormat),
			account.Name,
			categoryName,
		})
		types = append(types, transaction.Type)
	}

	theme := gui.CurrentTheme()

	t := gui.NewTable().
		WithTitle("Transactions", theme.TransactionListTitleBackground).
		WithSubtitle("Configured transactions").
		WithHeaderBackground(theme.TransactionListHeaderBackground).
		WithHeaders("ID", "Account", "Amount", "Currency", "Vendor", "Description", "Datetime", "Account", "Category")

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

func classifyFilter(tokenFilter parser.Token) int {
	switch tokenFilter.Kind {
	case parser.TokenText:
		if tokenFilter.Raw[0] == 'T' {
			return FilterTypeTransactionID
		}
	case parser.TokenAttribute:
		if tokenFilter.Key == "account" {
			return FilterTypeAccountName
		} else if tokenFilter.Key == "currency" {
			return FilterTypeCurrency
		} else if tokenFilter.Key == "store" {
			return FilterTypeStore
		} else if tokenFilter.Key == "desc" {
			return FilterTypeDescription
		} else if tokenFilter.Key == "date" {
			return FilterTypeDatetime
		} else if tokenFilter.Key == "period" {
			return FilterTypeDatetime
		} else if tokenFilter.Key == "time" {
			return FilterTypeDatetime
		}
	default:
		return FilterUnknown
	}
	return FilterUnknown
}

func createTransactionIDFilter(token parser.Token) (db.SQLFilter, error) {
	var transactionID string = token.Raw[1:]
	_, err := domain.ParseTransactionID(transactionID)
	if err != nil {
		return nil, err
	}
	return db.TransactionIDFilter{ID: transactionID}, nil
}

func createAccountNameFilter(token parser.Token) (db.SQLFilter, error) {
	return db.TransactionAccountNameFilter{Name: token.Value}, nil
}

func createCurrencyFilter(token parser.Token) (db.SQLFilter, error) {
	return db.TransactionCurrencyFilter{Currency: token.Value}, nil
}

func createStoreFilter(token parser.Token) (db.SQLFilter, error) {
	return db.TransactionStoreNameFilter{Store: token.Value}, nil
}

func createDescriptionFilter(token parser.Token) (db.SQLFilter, error) {
	return db.TransactionDescriptionFilter{Description: token.Value}, nil
}

func createDatetimeFilter(token parser.Token, config config.Config) (db.SQLFilter, error) {
	if domain.IsTimeShortcut(token.Value) {
		from, to, err := domain.GetTimeShortcut(token.Value)
		if err != nil {
			return nil, err
		}
		return db.TransactionDatetimeFilter{From: from, To: to}, nil
	} else if strings.Contains(token.Value, "-") {
		datetimeRange := strings.SplitN(token.Value, "-", 2)
		datetimeFrom, err := time.Parse(config.Display.DateFormat, datetimeRange[0])
		if err != nil {
			return nil, err
		}
		datetimeTo, err := time.Parse(config.Display.DateFormat, datetimeRange[1])
		if err != nil {
			return nil, err
		}
		return db.TransactionDatetimeFilter{From: datetimeFrom, To: datetimeTo}, nil
	}
	return nil, fmt.Errorf("unknown datetime format: %s. Must be given as shortcuts or from-to", token.Value)
}

func createFilters(
	parsed parser.ParsedCmdLine,
	cashDb db.DBTX,
	config config.Config,
) ([]db.SQLFilter, []db.Filter[db.Transaction], error) {
	var dbFilters []db.SQLFilter = make([]db.SQLFilter, 0, len(parsed.Filters))
	var runFilters []db.Filter[db.Transaction] = make([]db.Filter[db.Transaction], 0, len(parsed.Filters))
	for _, filter := range parsed.Filters {
		filterType := classifyFilter(filter)
		if filterType == FilterUnknown {
			return nil, nil, fmt.Errorf("unknown filter: %s", filter.Raw)
		} else if filterType == FilterTypeTransactionID {
			newFilter, err := createTransactionIDFilter(filter)
			if err != nil {
				return nil, nil, err
			}
			dbFilters = append(dbFilters, newFilter)
		} else if filterType == FilterTypeAccountName {
			newFilter, err := createAccountNameFilter(filter)
			if err != nil {
				return nil, nil, err
			}
			dbFilters = append(dbFilters, newFilter)
		} else if filterType == FilterTypeCurrency {
			newFilter, err := createCurrencyFilter(filter)
			if err != nil {
				return nil, nil, err
			}
			dbFilters = append(dbFilters, newFilter)
		} else if filterType == FilterTypeStore {
			newFilter, err := createStoreFilter(filter)
			if err != nil {
				return nil, nil, err
			}
			dbFilters = append(dbFilters, newFilter)
		} else if filterType == FilterTypeDescription {
			newFilter, err := createDescriptionFilter(filter)
			if err != nil {
				return nil, nil, err
			}
			dbFilters = append(dbFilters, newFilter)
		} else if filterType == FilterTypeDatetime {
			newFilter, err := createDatetimeFilter(filter, config)
			if err != nil {
				return nil, nil, err
			}
			dbFilters = append(dbFilters, newFilter)
		}
	}
	return dbFilters, runFilters, nil
}

func List(parsed parser.ParsedCmdLine, config config.Config, cashDb db.DBTX) error {
	dbFilters, runFilters, err := createFilters(parsed, cashDb, config)
	if err != nil {
		return err
	}

	// Get list of transactions
	transactions, err := db.ListTransactions(cashDb, dbFilters, runFilters)
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
