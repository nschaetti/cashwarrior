package cmd

import (
	"testing"
	"time"

	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestAddAccount(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	err := addAccount(parser.ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "add",
		Args: []parser.Token{
			{Kind: parser.TokenText, Raw: "savings"},
			{Kind: parser.TokenAttribute, Key: "currency", Value: "EUR", Raw: "currency:EUR"},
		},
	}, cfg, cashDB)
	if err != nil {
		t.Fatalf("addAccount returned error: %v", err)
	}

	account, err := db.GetAccountByName(cashDB, "savings")
	if err != nil {
		t.Fatalf("GetAccountByName returned error: %v", err)
	}
	if account.Currency != "EUR" {
		t.Fatalf("currency = %q, want EUR", account.Currency)
	}
}

func TestModifyAccount(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	_, err := db.InsertAccount(cashDB, db.CreateAccountInput{Name: "savings", Currency: "CHF"})
	if err != nil {
		t.Fatalf("InsertAccount returned error: %v", err)
	}

	err = modifyAccount(parser.ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "modify",
		Args: []parser.Token{
			{Kind: parser.TokenText, Raw: "savings"},
			{Kind: parser.TokenAttribute, Key: "account", Value: "brokerage", Raw: "account:brokerage"},
			{Kind: parser.TokenAttribute, Key: "currency", Value: "USD", Raw: "currency:USD"},
		},
	}, cfg, cashDB)
	if err != nil {
		t.Fatalf("modifyAccount returned error: %v", err)
	}

	account, err := db.GetAccountByName(cashDB, "brokerage")
	if err != nil {
		t.Fatalf("GetAccountByName returned error: %v", err)
	}
	if account.Currency != "USD" {
		t.Fatalf("currency = %q, want USD", account.Currency)
	}
}

func TestDeleteAccountRejectsLinkedTransactions(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	accountID, err := db.InsertAccount(cashDB, db.CreateAccountInput{Name: "savings", Currency: cfg.Default.Currency})
	if err != nil {
		t.Fatalf("InsertAccount returned error: %v", err)
	}
	placeID, err := db.InsertPlace(cashDB, db.CreatePlaceInput{Name: "Delete Account Test"})
	if err != nil {
		t.Fatalf("InsertPlace returned error: %v", err)
	}
	_, err = db.InsertTransaction(cashDB, db.CreateTransactionInput{Identifier: "2026.05.1", Amount: -5, Description: "x", Datetime: testTime(), AccountID: accountID, PlaceID: &placeID})
	if err != nil {
		t.Fatalf("InsertTransaction returned error: %v", err)
	}

	err = deleteAccount(parser.ParsedCmdLine{Command: "accounts", Subcommand: "delete", Args: []parser.Token{{Kind: parser.TokenText, Raw: "savings"}}}, cfg, cashDB)
	if err == nil {
		t.Fatal("deleteAccount expected error, got nil")
	}
}

func TestDeleteAccountDeletesEmptyAccountAfterConfirmation(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	_, err := db.InsertAccount(cashDB, db.CreateAccountInput{Name: "savings", Currency: cfg.Default.Currency})
	if err != nil {
		t.Fatalf("InsertAccount returned error: %v", err)
	}

	withInput(t, "y\n", func() {
		err = deleteAccount(parser.ParsedCmdLine{Command: "accounts", Subcommand: "delete", Args: []parser.Token{{Kind: parser.TokenText, Raw: "savings"}}}, cfg, cashDB)
	})
	if err != nil {
		t.Fatalf("deleteAccount returned error: %v", err)
	}

	_, err = db.GetAccountByName(cashDB, "savings")
	if err == nil {
		t.Fatal("GetAccountByName expected error, got nil")
	}
}

func testTime() (t time.Time) { return }
