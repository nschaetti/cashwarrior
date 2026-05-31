package cmd

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe returned error: %v", err)
	}

	oldStdout := os.Stdout
	os.Stdout = writer
	defer func() {
		os.Stdout = oldStdout
	}()

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close returned error: %v", err)
	}

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("io.ReadAll returned error: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("reader.Close returned error: %v", err)
	}

	return string(output)
}

func TestShowDisplaysTransactionDetails(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	mainAccount, err := db.GetAccountByName(cashDB, cfg.Default.Account)
	if err != nil {
		t.Fatalf("GetAccountByName returned error: %v", err)
	}

	placeID, err := db.InsertPlace(cashDB, db.CreatePlaceInput{Name: "Migros"})
	if err != nil {
		t.Fatalf("InsertPlace returned error: %v", err)
	}
	categoryID, err := db.InsertCategory(cashDB, db.CreateCategoryInput{Name: "groceries"})
	if err != nil {
		t.Fatalf("InsertCategory returned error: %v", err)
	}
	groupID, err := db.InsertTransactionGroup(cashDB, db.CreateTransactionGroupInput{Name: "weekly-shop"})
	if err != nil {
		t.Fatalf("InsertTransactionGroup returned error: %v", err)
	}
	tagID, err := db.InsertTag(cashDB, db.CreateTagInput{Name: "food"})
	if err != nil {
		t.Fatalf("InsertTag returned error: %v", err)
	}

	transactionID, err := db.InsertTransaction(cashDB, db.CreateTransactionInput{
		Identifier:  "2026.05.1",
		Amount:      -10.5,
		Description: "Lunch",
		Datetime:    time.Date(2026, time.May, 27, 12, 34, 56, 0, time.UTC),
		AccountID:   mainAccount.ID,
		CategoryID:  &categoryID,
		PlaceID:     &placeID,
		GroupID:     &groupID,
	})
	if err != nil {
		t.Fatalf("InsertTransaction returned error: %v", err)
	}
	if err := db.InsertTransactionTag(cashDB, transactionID, tagID); err != nil {
		t.Fatalf("InsertTransactionTag returned error: %v", err)
	}

	output := captureStdout(t, func() {
		if err := Show(parser.ParsedCmdLine{
			Command:    "show",
			Subcommand: "default",
			Args:       []parser.Token{{Raw: "2026.05.1", Kind: parser.TokenID}},
		}, cfg, cashDB); err != nil {
			t.Fatalf("Show returned error: %v", err)
		}
	})

	for _, want := range []string{
		"2026.05.1",
		"expense",
		"-10.50",
		mainAccount.Name,
		mainAccount.Currency,
		"Migros",
		"Lunch",
		"groceries",
		"weekly-shop",
		"food",
		"2026-05-27 12:34:56",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}

	if strings.Contains(output, "Value") {
		t.Fatalf("output unexpectedly contains header row:\n%s", output)
	}
}

func TestShowDisplaysTransferDetails(t *testing.T) {
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
	placeID, err := db.InsertPlace(cashDB, db.CreatePlaceInput{Name: "transfer-show-test"})
	if err != nil {
		t.Fatalf("InsertPlace returned error: %v", err)
	}

	fromTransactionID, err := db.InsertTransaction(cashDB, db.CreateTransactionInput{
		Identifier:  "2026.05.1",
		Type:        "transfer_out",
		Amount:      -20,
		Description: "Transfer",
		Datetime:    time.Date(2026, time.May, 27, 12, 0, 0, 0, time.UTC),
		AccountID:   fromAccount.ID,
		PlaceID:     &placeID,
	})
	if err != nil {
		t.Fatalf("InsertTransaction(from) returned error: %v", err)
	}
	toTransactionID, err := db.InsertTransaction(cashDB, db.CreateTransactionInput{
		Identifier:  "2026.05.2",
		Type:        "transfer_in",
		Amount:      20,
		Description: "Transfer",
		Datetime:    time.Date(2026, time.May, 27, 12, 0, 0, 0, time.UTC),
		AccountID:   toAccountID,
		PlaceID:     &placeID,
	})
	if err != nil {
		t.Fatalf("InsertTransaction(to) returned error: %v", err)
	}
	if _, err := db.InsertTransfer(cashDB, db.CreateTransferInput{
		FromTransactionID: fromTransactionID,
		ToTransactionID:   toTransactionID,
		FromAccountID:     fromAccount.ID,
		ToAccountID:       toAccountID,
		Amount:            20,
	}); err != nil {
		t.Fatalf("InsertTransfer returned error: %v", err)
	}

	output := captureStdout(t, func() {
		if err := Show(parser.ParsedCmdLine{
			Command:    "show",
			Subcommand: "default",
			Args:       []parser.Token{{Raw: "2026.05.1", Kind: parser.TokenID}},
		}, cfg, cashDB); err != nil {
			t.Fatalf("Show returned error: %v", err)
		}
	})

	for _, want := range []string{"Transfer amount", "20.00", "Transfer from account", fromAccount.Name, "Transfer to account", "savings", "Transfer pair ID", "2026.05.2"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
}
