package cmd

import (
	"testing"
	"time"

	"github.com/nschaetti/cashwarrior/internal/db"
)

func accountForBalance(id int64, name, currency string, initial float64) db.Account {
	return db.Account{ID: id, Name: name, Currency: currency, InitialBalance: initial}
}

func txForBalance(id int64, accountID int64, txType string, amount float64, when time.Time) db.Transaction {
	return db.Transaction{
		ID:        id,
		Type:      txType,
		Amount:    amount,
		Datetime:  when,
		AccountID: &accountID,
	}
}

func TestCalculateBalance_NoPeriodAccounts(t *testing.T) {
	accounts := []db.Account{
		accountForBalance(1, "main", "CHF", 100),
		accountForBalance(2, "savings", "CHF", 200),
	}
	transactions := []db.Transaction{
		txForBalance(1, 1, "expense", -30, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)),
		txForBalance(2, 2, "income", 50, time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC)),
	}
	now := time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC)

	data := calculateBalance(accounts, transactions, balanceInterval{End: now})

	if len(data.Accounts) != 2 {
		t.Fatalf("Accounts length = %d, want 2", len(data.Accounts))
	}
	if data.Accounts[0].Closing != 70 || data.Accounts[1].Closing != 250 {
		t.Fatalf("closings = (%f, %f), want (70, 250)", data.Accounts[0].Closing, data.Accounts[1].Closing)
	}
	if len(data.Currencies) != 1 {
		t.Fatalf("Currencies length = %d, want 1", len(data.Currencies))
	}
	if data.Currencies[0].Closing != 320 {
		t.Fatalf("currency closing = %f, want 320", data.Currencies[0].Closing)
	}
}

func TestCalculateBalance_PeriodOpeningIncludesPriorTransactions(t *testing.T) {
	accounts := []db.Account{accountForBalance(1, "main", "CHF", 100)}
	start := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 20, 23, 59, 59, 0, time.UTC)
	transactions := []db.Transaction{
		txForBalance(1, 1, "expense", -10, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)),
		txForBalance(2, 1, "income", 25, time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC)),
		txForBalance(3, 1, "expense", -5, time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)),
		txForBalance(4, 1, "expense", -50, time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)),
	}

	data := calculateBalance(accounts, transactions, balanceInterval{Start: &start, End: end})

	item := data.Accounts[0]
	if item.Opening != 90 {
		t.Fatalf("opening = %f, want 90", item.Opening)
	}
	if item.Income != 25 || item.Expenses != -5 {
		t.Fatalf("income/expenses = (%f, %f), want (25, -5)", item.Income, item.Expenses)
	}
	if item.Net != 20 {
		t.Fatalf("net = %f, want 20", item.Net)
	}
	if item.Closing != 110 {
		t.Fatalf("closing = %f, want 110", item.Closing)
	}
	if item.Operations != 2 {
		t.Fatalf("operations = %d, want 2", item.Operations)
	}
}

func TestCalculateBalance_ExcludesTransactionsAfterEnd(t *testing.T) {
	accounts := []db.Account{accountForBalance(1, "main", "CHF", 0)}
	now := time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)
	transactions := []db.Transaction{
		txForBalance(1, 1, "expense", -10, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)),
		txForBalance(2, 1, "expense", -20, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)),
	}

	data := calculateBalance(accounts, transactions, balanceInterval{End: now})

	if data.Accounts[0].Closing != -10 {
		t.Fatalf("closing = %f, want -10 (future transaction excluded)", data.Accounts[0].Closing)
	}
}

func TestCalculateBalance_TransferClassification(t *testing.T) {
	accounts := []db.Account{
		accountForBalance(1, "main", "CHF", 0),
		accountForBalance(2, "savings", "CHF", 0),
	}
	now := time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)
	transactions := []db.Transaction{
		txForBalance(1, 1, "transfer_out", -100, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)),
		txForBalance(2, 2, "transfer_in", 100, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)),
	}

	data := calculateBalance(accounts, transactions, balanceInterval{End: now})

	if data.Accounts[0].Transfers != -100 {
		t.Fatalf("main transfers = %f, want -100", data.Accounts[0].Transfers)
	}
	if data.Accounts[1].Transfers != 100 {
		t.Fatalf("savings transfers = %f, want 100", data.Accounts[1].Transfers)
	}
	// Transfers cancel out across accounts in the currency total.
	if data.Currencies[0].Transfers != 0 {
		t.Fatalf("currency transfers = %f, want 0 (cancelled)", data.Currencies[0].Transfers)
	}
}

func TestCalculateBalance_MultiCurrencyKeepsSeparateTotals(t *testing.T) {
	accounts := []db.Account{
		accountForBalance(1, "chf", "CHF", 100),
		accountForBalance(2, "eur", "EUR", 0),
	}
	now := time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)
	transactions := []db.Transaction{
		txForBalance(1, 1, "income", 10, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)),
		txForBalance(2, 2, "income", 5, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)),
	}

	data := calculateBalance(accounts, transactions, balanceInterval{End: now})

	if len(data.Currencies) != 2 {
		t.Fatalf("currencies length = %d, want 2", len(data.Currencies))
	}
	byCurrency := map[string]float64{}
	for _, c := range data.Currencies {
		byCurrency[c.Currency] = c.Closing
	}
	if byCurrency["CHF"] != 110 || byCurrency["EUR"] != 5 {
		t.Fatalf("currency closings = %v, want CHF=110, EUR=5", byCurrency)
	}
}

func TestCalculateBalance_IgnoresTransactionsWithoutAccount(t *testing.T) {
	accounts := []db.Account{accountForBalance(1, "main", "CHF", 0)}
	var missing *int64
	now := time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)
	transactions := []db.Transaction{
		{ID: 1, Type: "expense", Amount: -99, Datetime: now, AccountID: missing},
	}

	data := calculateBalance(accounts, transactions, balanceInterval{End: now})

	if data.Accounts[0].Closing != 0 {
		t.Fatalf("closing = %f, want 0 (no-account transaction ignored)", data.Accounts[0].Closing)
	}
}
