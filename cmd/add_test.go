package cmd

import (
	"testing"
	"time"

	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestAddUsesProvidedDateMonthForIdentifier(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	mainAccount, err := db.GetAccountByName(cashDB, cfg.Default.Account)
	if err != nil {
		t.Fatalf("GetAccountByName returned error: %v", err)
	}
	placeID, err := db.InsertPlace(cashDB, db.CreatePlaceInput{Name: "existing-place"})
	if err != nil {
		t.Fatalf("InsertPlace returned error: %v", err)
	}

	_, err = db.InsertTransaction(cashDB, db.CreateTransactionInput{
		Identifier:  "2026.04.2",
		Type:        "expense",
		Amount:      -2,
		Description: "April existing",
		Datetime:    time.Date(2026, time.April, 10, 12, 0, 0, 0, time.UTC),
		AccountID:   mainAccount.ID,
		PlaceID:     &placeID,
	})
	if err != nil {
		t.Fatalf("InsertTransaction(april) returned error: %v", err)
	}
	_, err = db.InsertTransaction(cashDB, db.CreateTransactionInput{
		Identifier:  "2026.06.1",
		Type:        "expense",
		Amount:      -3,
		Description: "June existing",
		Datetime:    time.Date(2026, time.June, 1, 12, 0, 0, 0, time.UTC),
		AccountID:   mainAccount.ID,
		PlaceID:     &placeID,
	})
	if err != nil {
		t.Fatalf("InsertTransaction(june) returned error: %v", err)
	}

	err = Add(parser.ParsedCmdLine{
		Command:    "add",
		Subcommand: "default",
		Args: []parser.Token{
			{Kind: parser.TokenAmount, Raw: "-4", Amount: -4},
			{Kind: parser.TokenText, Raw: "hehe"},
			{Kind: parser.TokenAttribute, Key: "date", Value: "27.04.2026", Raw: "date:27.04.2026"},
			{Kind: parser.TokenAttribute, Key: "store", Value: "coop", Raw: "store:coop"},
		},
	}, cfg, cashDB)
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	transaction, err := db.GetTransactionByIdentifier(cashDB, "2026.04.3")
	if err != nil {
		t.Fatalf("GetTransactionByIdentifier returned error: %v", err)
	}
	if transaction.Description != "hehe" {
		t.Fatalf("transaction.Description = %q, want hehe", transaction.Description)
	}
	if !transaction.Datetime.Equal(time.Date(2026, time.April, 27, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("transaction.Datetime = %v, want 2026-04-27", transaction.Datetime)
	}
}

func TestTransferUsesProvidedDateMonthForIdentifiers(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	fromAccount, err := db.GetAccountByName(cashDB, cfg.Default.Account)
	if err != nil {
		t.Fatalf("GetAccountByName returned error: %v", err)
	}
	toAccountID, err := db.InsertAccount(cashDB, db.CreateAccountInput{Name: "savings", Currency: cfg.Default.Currency})
	if err != nil {
		t.Fatalf("InsertAccount returned error: %v", err)
	}
	_ = toAccountID
	transferPlace, err := db.GetPlaceByName(cashDB, "transfer")
	if err != nil {
		t.Fatalf("GetPlaceByName returned error: %v", err)
	}

	_, err = db.InsertTransaction(cashDB, db.CreateTransactionInput{
		Identifier:  "2026.04.5",
		Type:        "expense",
		Amount:      -5,
		Description: "April existing",
		Datetime:    time.Date(2026, time.April, 20, 12, 0, 0, 0, time.UTC),
		AccountID:   fromAccount.ID,
		PlaceID:     &transferPlace.ID,
	})
	if err != nil {
		t.Fatalf("InsertTransaction(april) returned error: %v", err)
	}
	_, err = db.InsertTransaction(cashDB, db.CreateTransactionInput{
		Identifier:  "2026.06.9",
		Type:        "expense",
		Amount:      -9,
		Description: "June existing",
		Datetime:    time.Date(2026, time.June, 2, 12, 0, 0, 0, time.UTC),
		AccountID:   fromAccount.ID,
		PlaceID:     &transferPlace.ID,
	})
	if err != nil {
		t.Fatalf("InsertTransaction(june) returned error: %v", err)
	}

	err = Transfer(parser.ParsedCmdLine{
		Command:    "transfer",
		Subcommand: "default",
		Args: []parser.Token{
			{Kind: parser.TokenAmount, Raw: "+20", Amount: 20},
			{Kind: parser.TokenText, Raw: "rent"},
			{Kind: parser.TokenAttribute, Key: "from", Value: cfg.Default.Account, Raw: "from:" + cfg.Default.Account},
			{Kind: parser.TokenAttribute, Key: "to", Value: "savings", Raw: "to:savings"},
			{Kind: parser.TokenAttribute, Key: "date", Value: "27.04.2026", Raw: "date:27.04.2026"},
		},
	}, cfg, cashDB)
	if err != nil {
		t.Fatalf("Transfer returned error: %v", err)
	}

	if _, err := db.GetTransactionByIdentifier(cashDB, "2026.04.6"); err != nil {
		t.Fatalf("GetTransactionByIdentifier(2026.04.6) returned error: %v", err)
	}
	if _, err := db.GetTransactionByIdentifier(cashDB, "2026.04.7"); err != nil {
		t.Fatalf("GetTransactionByIdentifier(2026.04.7) returned error: %v", err)
	}
}
