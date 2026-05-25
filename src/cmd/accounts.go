package cmd

import (
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Accounts(parsed parser.ParsedCmdLine, config config.Config, cashDb *sql.DB) error {
	accounts, err := db.ListAccounts(cashDb, db.AccountListFilter{})
	if err != nil {
		return err
	}

	transactions, err := db.ListTransactions(cashDb, []db.Filter{})
	if err != nil {
		return err
	}

	type accountTotals struct {
		Balance       float64
		Operations    int
		MonthIncome   float64
		MonthExpenses float64
		MonthTransfer float64
	}

	totalsByAccount := make(map[int64]*accountTotals, len(accounts))
	for _, account := range accounts {
		totalsByAccount[account.ID] = &accountTotals{}
	}

	now := time.Now()
	currentYear, currentMonth, _ := now.Date()

	for _, transaction := range transactions {
		if transaction.AccountID == nil {
			continue
		}
		totals, ok := totalsByAccount[*transaction.AccountID]
		if !ok {
			continue
		}

		totals.Balance += transaction.Amount
		totals.Operations++

		ty, tm, _ := transaction.Datetime.Date()
		if ty != currentYear || tm != currentMonth {
			continue
		}

		switch transaction.Type {
		case "income":
			totals.MonthIncome += transaction.Amount
		case "expense":
			totals.MonthExpenses += transaction.Amount
		case "transfer_in", "transfer_out":
			totals.MonthTransfer += transaction.Amount
		}
	}

	rows := make([][]string, 0, len(accounts))
	formatAmount := func(amount float64, forceSign bool) string {
		if forceSign {
			return fmt.Sprintf("%+.2f", amount)
		}
		if amount == 0 {
			return "0.00"
		}
		return fmt.Sprintf("%.2f", amount)
	}

	for _, account := range accounts {
		totals := totalsByAccount[account.ID]
		rows = append(rows, []string{
			strconv.FormatInt(account.ID, 10),
			account.Name,
			account.Currency,
			formatAmount(totals.Balance, true),
			strconv.Itoa(totals.Operations),
		})
	}

	theme := gui.CurrentTheme()

	t := gui.NewTable().
		WithTitle("Accounts", theme.AccountsTitleBackground).
		WithSubtitle("Configured cash accounts").
		WithHeaderBackground(theme.AccountsHeaderBackground).
		WithHeaders("ID", "Name", "Currency", "Balance", "Operations").
		AddRows(rows)

	fmt.Println(t.Render())
	fmt.Println()
	fmt.Println()

	summaryRows := make([][]string, 0, len(accounts))
	sortedAccounts := make([]db.Account, len(accounts))
	copy(sortedAccounts, accounts)
	sort.Slice(sortedAccounts, func(i, j int) bool {
		return sortedAccounts[i].Name < sortedAccounts[j].Name
	})

	for _, account := range sortedAccounts {
		totals := totalsByAccount[account.ID]
		net := totals.MonthIncome + totals.MonthExpenses
		summaryRows = append(summaryRows, []string{
			account.Name,
			account.Currency,
			formatAmount(totals.MonthIncome, true),
			formatAmount(totals.MonthExpenses, false),
			formatAmount(net, true),
			formatAmount(totals.MonthTransfer, true),
		})
	}

	summary := gui.NewTable().
		WithType(gui.TableTypeSummary).
		WithTitle("Current month summary", theme.AccountsTitleBackground).
		WithHeaderBackground(theme.AccountsHeaderBackground).
		WithHeaders("Account", "Currency", "Income", "Expenses", "Net", "Transfers").
		AddRows(summaryRows)

	fmt.Println(summary.Render())
	fmt.Println()
	fmt.Println()

	return nil
}
