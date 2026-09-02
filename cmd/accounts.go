package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/parser"
	"github.com/nschaetti/cashwarrior/internal/utils"
)

func Accounts(parsed parser.ParsedCmdLine, config config.Config, cashDb db.DBTX) error {
	switch parsed.Subcommand {
	case "list":
		return accountListJSON(parsed, config, cashDb)
	case "balance":
		return accountBalance(parsed, config, cashDb)
	case "add":
		return addAccount(parsed, config, cashDb)
	case "modify":
		return modifyAccount(parsed, config, cashDb)
	case "rename":
		return renameAccount(parsed, config, cashDb)
	case "delete":
		return deleteAccount(parsed, config, cashDb)
	case "initial-balance":
		return setAccountInitialBalance(parsed, cashDb)
	default:
		return fmt.Errorf("unknown accounts subcommand %s", parsed.Subcommand)
	}
}

func setAccountInitialBalance(parsed parser.ParsedCmdLine, cashDb db.DBTX) error {
	var accountName string
	var amountValue string

	for _, arg := range parsed.Args {
		attr, ok := arg.(parser.ArgAttribute)
		if ok {
			switch attr.Key {
			case "account":
				accountName = strings.TrimSpace(attr.Value.Raw)
			case "amount", "initial_balance":
				amountValue = strings.TrimSpace(attr.Value.Raw)
			}
		}
	}

	textArgs := make([]string, 0, 2)
	for _, arg := range parsed.Args {
		text, ok := arg.(parser.ArgText)
		if ok {
			textArgs = append(textArgs, strings.TrimSpace(text.Text))
		}
	}

	if accountName == "" || amountValue == "" {
		if len(textArgs) != 2 {
			return fmt.Errorf("accounts initial-balance requires an account and an amount")
		}
		firstAmount, firstErr := strconv.ParseFloat(textArgs[0], 64)
		secondAmount, secondErr := strconv.ParseFloat(textArgs[1], 64)
		_ = firstAmount
		_ = secondAmount

		switch {
		case firstErr == nil && secondErr == nil:
			return fmt.Errorf("accounts initial-balance is ambiguous: both arguments look like amounts")
		case firstErr != nil && secondErr != nil:
			return fmt.Errorf("accounts initial-balance requires one numeric amount")
		case firstErr == nil:
			if amountValue == "" {
				amountValue = textArgs[0]
			}
			if accountName == "" {
				accountName = textArgs[1]
			}
		default:
			if amountValue == "" {
				amountValue = textArgs[1]
			}
			if accountName == "" {
				accountName = textArgs[0]
			}
		}
	}

	if accountName == "" || amountValue == "" {
		return fmt.Errorf("accounts initial-balance requires an account and an amount")
	}

	amount, err := strconv.ParseFloat(amountValue, 64)
	if err != nil {
		return fmt.Errorf("invalid amount %q", amountValue)
	}

	account, err := db.GetAccountByName(cashDb, accountName)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("account %s does not exist", accountName)
	}
	if err != nil {
		return err
	}

	if err := db.UpdateAccountInitialBalance(cashDb, account.ID, amount); err != nil {
		return err
	}
	if isJSONOutput(parsed) {
		return renderJSON("account", map[string]any{"action": "initial_balance_updated", "account": accountName, "amount": amount}, 1)
	}
	fmt.Printf("Account %s initial balance set to %.2f\n", accountName, amount)
	return nil
}

func getAccountNameArg(arg parser.Arg) (string, error) {
	text, ok := arg.(parser.ArgText)
	if ok {
		if text.Text == "" {
			return "", fmt.Errorf("account name cannot be empty")
		}
		return text.Text, nil
	}
	attr, ok := arg.(parser.ArgAttribute)
	if ok {
		if attr.Key == "account" && !attr.Value.IsEmpty() {
			return attr.Value.Raw, nil
		}
	}
	return "", fmt.Errorf("account name is required")
}

func getAttributesFromTokens(args []parser.Arg) map[string]string {
	attributes := make(map[string]string)
	for _, arg := range args {
		attr, ok := arg.(parser.ArgAttribute)
		if !ok {
			continue
		}
		attributes[attr.Key] = attr.Value.Raw
	}
	return attributes
}

func addAccount(parsed parser.ParsedCmdLine, cfg config.Config, cashDb db.DBTX) error {
	name, err := getAccountNameArg(parsed.Args[0])
	if err != nil {
		return err
	}
	attributes := getAttributesFromTokens(parsed.Args)
	currency := cfg.Default.Currency
	if value, ok := attributes["currency"]; ok {
		currency = strings.TrimSpace(value)
	}
	initialBalance := 0.0
	if value, ok := attributes["initial_balance"]; ok {
		initialBalance, err = strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return fmt.Errorf("invalid initial_balance %q", value)
		}
	}
	if strings.TrimSpace(currency) == "" {
		return fmt.Errorf("currency cannot be empty")
	}
	exists, err := db.AccountExists(cashDb, name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("account %s already exists", name)
	}
	_, err = db.InsertAccount(cashDb, db.CreateAccountInput{Name: name, Currency: currency, InitialBalance: initialBalance})
	if err != nil {
		return err
	}
	if isJSONOutput(parsed) {
		return renderJSON("account", map[string]any{"action": "created", "name": name}, 1)
	}
	fmt.Printf("Account %s created\n", name)
	return nil
}

func modifyAccount(parsed parser.ParsedCmdLine, _ config.Config, cashDb db.DBTX) error {
	name, err := getAccountNameArg(parsed.Args[0])
	if err != nil {
		return err
	}
	account, err := db.GetAccountByName(cashDb, name)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("account %s does not exist", name)
	}
	if err != nil {
		return err
	}
	attributes := getAttributesFromTokens(parsed.Args[1:])
	if newName, ok := attributes["account"]; ok {
		newName = strings.TrimSpace(newName)
		if newName == "" {
			return fmt.Errorf("account name cannot be empty")
		}
		if newName != account.Name {
			exists, err := db.AccountExists(cashDb, newName)
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("account %s already exists", newName)
			}
			if err := db.UpdateAccountName(cashDb, account.ID, newName); err != nil {
				return err
			}
		}
	}
	if currency, ok := attributes["currency"]; ok {
		currency = strings.TrimSpace(currency)
		if currency == "" {
			return fmt.Errorf("currency cannot be empty")
		}
		if err := db.UpdateAccountCurrency(cashDb, account.ID, currency); err != nil {
			return err
		}
	}
	if value, ok := attributes["initial_balance"]; ok {
		initialBalance, parseErr := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if parseErr != nil {
			return fmt.Errorf("invalid initial_balance %q", value)
		}
		if err := db.UpdateAccountInitialBalance(cashDb, account.ID, initialBalance); err != nil {
			return err
		}
	}
	if isJSONOutput(parsed) {
		return renderJSON("account", map[string]any{"action": "updated", "name": name}, 1)
	}
	fmt.Printf("Account %s updated\n", name)
	return nil
}

func renameAccount(parsed parser.ParsedCmdLine, cfg config.Config, cashDb db.DBTX) error {
	if err := requireYesForJSON(parsed); err != nil {
		return err
	}
	name, err := getAccountNameArg(parsed.Args[0])
	if err != nil {
		return err
	}
	newName, err := getAccountNameArg(parsed.Args[1])
	if err != nil {
		return err
	}
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return fmt.Errorf("account name cannot be empty")
	}
	if name == newName {
		return nil
	}
	if name == cfg.Default.Account {
		if !parsed.HasFlag("yes") && !isJSONOutput(parsed) {
			fmt.Printf("Account %s is your default account.\n", name)
		}
		if !parsed.HasFlag("yes") && !utils.AskYesNo("Rename it and update default.account in config?") {
			return nil
		}
		cfg.Default.Account = newName
		configPath := utils.ExpandPath(config.DefaultConfigFile)
		if err := config.SaveConfig(configPath, cfg); err != nil {
			return err
		}
	}
	account, err := db.GetAccountByName(cashDb, name)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("account %s does not exist", name)
	}
	if err != nil {
		return err
	}
	exists, err := db.AccountExists(cashDb, newName)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("account %s already exists", newName)
	}
	if err := db.UpdateAccountName(cashDb, account.ID, newName); err != nil {
		return err
	}
	if isJSONOutput(parsed) {
		return renderJSON("account", map[string]any{"action": "renamed", "name": newName}, 1)
	}
	fmt.Printf("Account %s renamed to %s\n", name, newName)
	return nil
}

func deleteAccount(parsed parser.ParsedCmdLine, cfg config.Config, cashDb db.DBTX) error {
	if err := requireYesForJSON(parsed); err != nil {
		return err
	}
	name, err := getAccountNameArg(parsed.Args[0])
	if err != nil {
		return err
	}
	if name == cfg.Default.Account {
		return fmt.Errorf("cannot delete default account %s", name)
	}
	account, err := db.GetAccountByName(cashDb, name)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("account %s does not exist", name)
	}
	if err != nil {
		return err
	}
	count, err := db.CountTransactionsByAccountID(cashDb, account.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("account %s has linked transactions", name)
	}
	if !parsed.HasFlag("yes") && !utils.AskYesNo(fmt.Sprintf("Delete account %s?", name)) {
		return nil
	}
	if err := db.DeleteAccountByID(cashDb, account.ID); err != nil {
		return err
	}
	if isJSONOutput(parsed) {
		return renderJSON("account", map[string]any{"action": "deleted", "name": name}, 1)
	}
	fmt.Printf("Account %s deleted\n", name)
	return nil
}

func listAccounts(config config.Config, cashDb db.DBTX) error {
	accounts, err := db.ListAccounts(cashDb, []db.SQLFilter{}, []db.Filter[db.Account]{}, []string{})
	if err != nil {
		return err
	}

	transactions, err := db.ListTransactions(cashDb, []db.SQLFilter{}, []db.Filter[db.Transaction]{}, false)
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
		totalsByAccount[account.ID] = &accountTotals{Balance: account.InitialBalance}
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
			formatAmount(account.InitialBalance, true),
			formatAmount(totals.Balance, true),
			strconv.Itoa(totals.Operations),
		})
	}

	theme := gui.CurrentTheme()

	t := gui.NewTable().
		WithTitle("Accounts", theme.AccountsTitleBackground).
		WithSubtitle("Configured cash accounts").
		WithHeaderBackground(theme.AccountsHeaderBackground).
		WithHeaders("ID", "Name", "Currency", "Initial Balance", "Balance", "Operations").
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
