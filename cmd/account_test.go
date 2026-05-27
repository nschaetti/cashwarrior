package cmd

import (
	"testing"
	"time"

	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestGetAccountCommandName(t *testing.T) {
	parsed := parser.ParsedCmdLine{
		Command: "account",
		Args:    []parser.Token{{Kind: parser.TokenText, Raw: "ayasdi-wallet"}},
	}

	name, err := getAccountCommandName(parsed)
	if err != nil {
		t.Fatalf("getAccountCommandName returned error: %v", err)
	}
	if name != "ayasdi-wallet" {
		t.Fatalf("name = %q, want %q", name, "ayasdi-wallet")
	}
}

func TestBuildRunningBalanceByTransactionID(t *testing.T) {
	transactions := []db.Transaction{
		{ID: 3, Amount: -5, Datetime: time.Date(2026, time.May, 2, 9, 0, 0, 0, time.UTC)},
		{ID: 1, Amount: 20, Datetime: time.Date(2026, time.May, 1, 9, 0, 0, 0, time.UTC)},
		{ID: 2, Amount: -7, Datetime: time.Date(2026, time.May, 1, 9, 0, 0, 0, time.UTC)},
	}

	sortTransactionsForRunningBalance(transactions)
	balanceByTransactionID := buildRunningBalanceByTransactionID(transactions)

	if balanceByTransactionID[1] != 20 {
		t.Fatalf("balance for tx 1 = %v, want 20", balanceByTransactionID[1])
	}
	if balanceByTransactionID[2] != 13 {
		t.Fatalf("balance for tx 2 = %v, want 13", balanceByTransactionID[2])
	}
	if balanceByTransactionID[3] != 8 {
		t.Fatalf("balance for tx 3 = %v, want 8", balanceByTransactionID[3])
	}
}

func TestHasAccountFilter(t *testing.T) {
	filters := []parser.Token{{Kind: parser.TokenAttribute, Key: "account", Value: "main"}}
	if !hasAccountFilter(filters) {
		t.Fatal("hasAccountFilter returned false, want true")
	}
}
