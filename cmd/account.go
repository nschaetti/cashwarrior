package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/output"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func accountBalance(parsed parser.ParsedCmdLine, config config.Config, cashDb db.DBTX) error {
	accountName, err := getAccountCommandName(parsed)
	if err != nil {
		return err
	}

	account, err := db.GetAccountByName(cashDb, accountName)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("account %s does not exist", accountName)
	}
	if err != nil {
		return err
	}

	dbFilters, runFilters, err := createFilters(parsed, cashDb, config)
	if err != nil {
		return err
	}

	accountFilter := db.TransactionAccountNameFilter{Name: account.Name}
	historyTransactions, err := db.ListTransactions(cashDb, []db.SQLFilter{accountFilter}, []db.Filter[db.Transaction]{}, false)
	if err != nil {
		return err
	}
	sortTransactionsForRunningBalance(historyTransactions)
	balanceByTransactionID := buildRunningBalanceByTransactionID(historyTransactions, account.InitialBalance)

	dbFilters = append(dbFilters, accountFilter)
	transactions, err := db.ListTransactions(cashDb, dbFilters, runFilters, false)
	if err != nil {
		return err
	}
	sortTransactionsForRunningBalance(transactions)
	format, err := commandOutputFormat(parsed)
	if err != nil {
		return err
	}
	if format == output.FormatJSON {
		data, err := getAccountBalanceData(cashDb, account, transactions, balanceByTransactionID)
		if err != nil {
			return err
		}
		return renderJSON("account_balance", data, len(data.Transactions))
	}

	return printAccountTransactionTable(cashDb, account, transactions, balanceByTransactionID, config)
}

func getAccountCommandName(parsed parser.ParsedCmdLine) (string, error) {
	arg := parsed.Args[0]
	attr, ok := arg.(parser.ArgAttribute)
	if ok {
		if attr.Key != "account" || attr.Value.IsEmpty() {
			return "", fmt.Errorf("account command requires an account name")
		}
		return attr.Value.Raw, nil
	}
	if arg.RawString() == "" {
		return "", fmt.Errorf("account command requires an account name")
	}
	return arg.RawString(), nil
}

func sortTransactionsForRunningBalance(transactions []db.Transaction) {
	sort.Slice(transactions, func(i, j int) bool {
		if transactions[i].Datetime.Equal(transactions[j].Datetime) {
			return transactions[i].ID < transactions[j].ID
		}
		return transactions[i].Datetime.Before(transactions[j].Datetime)
	})
}

func buildRunningBalanceByTransactionID(transactions []db.Transaction, initialBalance float64) map[int64]float64 {
	balanceByTransactionID := make(map[int64]float64, len(transactions))
	balance := initialBalance
	for _, transaction := range transactions {
		balance += transaction.Amount
		balanceByTransactionID[transaction.ID] = balance
	}
	return balanceByTransactionID
}

func printAccountTransactionTable(
	cashDb db.DBTX,
	account db.Account,
	transactions []db.Transaction,
	balanceByTransactionID map[int64]float64,
	config config.Config,
) error {
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

		vendor, err := transaction.GetPlace(cashDb)
		if err != nil {
			return err
		}
		vendorName := ""
		if vendor != nil {
			vendorName = vendor.Name
		}

		rows = append(rows, []string{
			transaction.Identifier,
			strconv.FormatFloat(transaction.Amount, 'f', 2, 64),
			account.Currency,
			vendorName,
			transaction.Description,
			transaction.Datetime.Format("2006-01-02"),
			categoryName,
			strconv.FormatFloat(balanceByTransactionID[transaction.ID], 'f', 2, 64),
		})
		types = append(types, transaction.Type)
	}

	theme := gui.CurrentTheme()
	t := gui.NewTable().
		WithTitle("Account", theme.AccountsTitleBackground).
		WithSubtitle(account.Name).
		WithHeaderBackground(theme.AccountsHeaderBackground).
		WithHeaders("ID", "Amount", "Currency", "Vendor", "Description", "Date", "Category", "Balance")

	for i, row := range rows {
		t.AddRowWithMetadata(row, map[string]string{"type": types[i]})
	}

	fmt.Println(t.Render())
	fmt.Printf(" Returned %d transactions\n\n\n", len(transactions))
	return nil
}

func transactionSortKey(t db.Transaction) time.Time {
	return t.Datetime
}
