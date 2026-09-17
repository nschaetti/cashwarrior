package cmd

import (
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/output"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func setupTransferTestData(t *testing.T) (config.Config, *sql.DB) {
	t.Helper()

	cfg, cashDB := openTestDB(t)

	fromAccount, err := db.GetAccountByName(cashDB, cfg.Default.Account)
	if err != nil {
		t.Fatalf("GetAccountByName returned error: %v", err)
	}
	toAccountID, err := db.InsertAccount(cashDB, db.CreateAccountInput{Name: "savings", Currency: cfg.Default.Currency})
	if err != nil {
		t.Fatalf("InsertAccount(savings) returned error: %v", err)
	}
	transferPlace, err := db.GetStoreByName(cashDB, "transfer")
	if err != nil {
		t.Fatalf("GetStoreByName(transfer) returned error: %v", err)
	}

	fromTxID, err := db.InsertTransaction(cashDB, db.CreateTransactionInput{
		Identifier:  "2026.05.1",
		Type:        "transfer_out",
		Amount:      -20,
		Description: "Transfer",
		Date:        time.Date(2026, time.May, 27, 12, 0, 0, 0, time.UTC),
		AccountID:   fromAccount.ID,
		PlaceID:     &transferPlace.ID,
	})
	if err != nil {
		t.Fatalf("InsertTransaction(from) returned error: %v", err)
	}
	toTxID, err := db.InsertTransaction(cashDB, db.CreateTransactionInput{
		Identifier:  "2026.05.2",
		Type:        "transfer_in",
		Amount:      20,
		Description: "Transfer",
		Date:        time.Date(2026, time.May, 27, 12, 0, 0, 0, time.UTC),
		AccountID:   toAccountID,
		PlaceID:     &transferPlace.ID,
	})
	if err != nil {
		t.Fatalf("InsertTransaction(to) returned error: %v", err)
	}
	if _, err := db.InsertTransfer(cashDB, db.CreateTransferInput{
		FromTransactionID: fromTxID,
		ToTransactionID:   toTxID,
		FromAccountID:     fromAccount.ID,
		ToAccountID:       toAccountID,
		Amount:            20,
	}); err != nil {
		t.Fatalf("InsertTransfer returned error: %v", err)
	}

	if _, err := db.InsertTransaction(cashDB, db.CreateTransactionInput{
		Identifier:  "2026.04.1",
		Type:        "expense",
		Amount:      -5,
		Description: "Unrelated",
		Date:        time.Date(2026, time.April, 20, 12, 0, 0, 0, time.UTC),
		AccountID:   fromAccount.ID,
		PlaceID:     &transferPlace.ID,
	}); err != nil {
		t.Fatalf("InsertTransaction(unrelated) returned error: %v", err)
	}

	return cfg, cashDB
}

func TestTransferDeleteMarksPairDeleted(t *testing.T) {
	cfg, cashDB := setupTransferTestData(t)
	defer cashDB.Close()

	got := captureListOutput(t, func() error {
		return Transfer(parser.ParsedCmdLine{
			Command:    "transfer",
			Subcommand: "delete",
			Flags:      yesFlag(),
			Args:       []parser.Arg{testArg(t, "identifier:2026.05.1")},
		}, cfg, cashDB)
	})
	if !strings.Contains(got, "Transfer 2026.05.1 deleted") {
		t.Fatalf("output missing success message: %s", got)
	}

	fromTx, err := db.GetTransactionByIdentifier(cashDB, "2026.05.1")
	if err != nil {
		t.Fatalf("GetTransactionByIdentifier(from) returned error: %v", err)
	}
	if !fromTx.Deleted {
		t.Fatal("from transfer transaction should be deleted")
	}
	toTx, err := db.GetTransactionByIdentifier(cashDB, "2026.05.2")
	if err != nil {
		t.Fatalf("GetTransactionByIdentifier(to) returned error: %v", err)
	}
	if !toTx.Deleted {
		t.Fatal("to transfer transaction should be deleted")
	}

	transfer, err := db.GetTransferByTransactionIDIncludingDeleted(cashDB, fromTx.ID)
	if err != nil {
		t.Fatalf("GetTransferByTransactionIDIncludingDeleted returned error: %v", err)
	}
	if !transfer.Deleted {
		t.Fatal("transfer should be marked deleted")
	}

	unrelated, err := db.GetTransactionByIdentifier(cashDB, "2026.04.1")
	if err != nil {
		t.Fatalf("GetTransactionByIdentifier(unrelated) returned error: %v", err)
	}
	if unrelated.Deleted {
		t.Fatal("unrelated transaction should remain untouched")
	}
}

func TestTransferDeleteByToIdentifier(t *testing.T) {
	cfg, cashDB := setupTransferTestData(t)
	defer cashDB.Close()

	if err := Transfer(parser.ParsedCmdLine{
		Command:    "transfer",
		Subcommand: "delete",
		Flags:      yesFlag(),
		Args:       []parser.Arg{testArg(t, "identifier:2026.05.2")},
	}, cfg, cashDB); err != nil {
		t.Fatalf("Transfer returned error: %v", err)
	}

	fromTx, err := db.GetTransactionByIdentifier(cashDB, "2026.05.1")
	if err != nil {
		t.Fatalf("GetTransactionByIdentifier(from) returned error: %v", err)
	}
	if !fromTx.Deleted {
		t.Fatal("from transfer transaction should be deleted")
	}
	toTx, err := db.GetTransactionByIdentifier(cashDB, "2026.05.2")
	if err != nil {
		t.Fatalf("GetTransactionByIdentifier(to) returned error: %v", err)
	}
	if !toTx.Deleted {
		t.Fatal("to transfer transaction should be deleted")
	}
}

func TestTransferDeleteJSONRequiresYes(t *testing.T) {
	cfg, cashDB := setupTransferTestData(t)
	defer cashDB.Close()

	err := Transfer(parser.ParsedCmdLine{
		Command:    "transfer",
		Subcommand: "delete",
		Flags:      []parser.Arg{parser.ArgFlag{Key: "json", Value: parser.BoolItem{Raw: "true", Value: true}}},
		Args:       []parser.Arg{testArg(t, "identifier:2026.05.1")},
	}, cfg, cashDB)
	if err == nil || err.Error() != "--yes is required for mutating commands in JSON mode" {
		t.Fatalf("err = %v, want --yes is required for mutating commands in JSON mode", err)
	}

	fromTx, err := db.GetTransactionByIdentifier(cashDB, "2026.05.1")
	if err != nil {
		t.Fatalf("GetTransactionByIdentifier(from) returned error: %v", err)
	}
	if fromTx.Deleted {
		t.Fatal("transaction should not be deleted without --yes in JSON mode")
	}
}

func TestTransferDeleteJSON(t *testing.T) {
	cfg, cashDB := setupTransferTestData(t)
	defer cashDB.Close()

	got := captureListOutput(t, func() error {
		return Transfer(parser.ParsedCmdLine{
			Command:    "transfer",
			Subcommand: "delete",
			Flags: []parser.Arg{
				parser.ArgFlag{Key: "json", Value: parser.BoolItem{Raw: "true", Value: true}},
				parser.ArgFlag{Raw: "--yes", Key: "yes"},
			},
			Args: []parser.Arg{testArg(t, "identifier:2026.05.1")},
		}, cfg, cashDB)
	})

	var envelope struct {
		Success bool   `json:"success"`
		Type    string `json:"type"`
		Count   int    `json:"count"`
	}
	if err := json.Unmarshal([]byte(got), &envelope); err != nil {
		t.Fatalf("JSON output is invalid: %v", err)
	}
	if !envelope.Success || envelope.Type != "transfer" || envelope.Count != 1 {
		t.Fatalf("JSON envelope = %#v", envelope)
	}
	if !strings.Contains(got, `"action":"deleted"`) || !strings.Contains(got, `"identifier":"2026.05.1"`) {
		t.Fatalf("JSON output missing deleted info: %s", got)
	}
}

func TestTransferDeleteNotFound(t *testing.T) {
	cfg, cashDB := setupTransferTestData(t)
	defer cashDB.Close()

	err := Transfer(parser.ParsedCmdLine{
		Command:    "transfer",
		Subcommand: "delete",
		Flags:      yesFlag(),
		Args:       []parser.Arg{testArg(t, "identifier:2026.99.9")},
	}, cfg, cashDB)
	if err == nil || err.Error() != "transaction 2026.99.9 does not exist" {
		t.Fatalf("err = %v, want transaction 2026.99.9 does not exist", err)
	}
}

func TestTransferDeleteNotATransfer(t *testing.T) {
	cfg, cashDB := setupTransferTestData(t)
	defer cashDB.Close()

	err := Transfer(parser.ParsedCmdLine{
		Command:    "transfer",
		Subcommand: "delete",
		Flags:      yesFlag(),
		Args:       []parser.Arg{testArg(t, "identifier:2026.04.1")},
	}, cfg, cashDB)
	if err == nil || err.Error() != "transaction 2026.04.1 is not a transfer" {
		t.Fatalf("err = %v, want transaction 2026.04.1 is not a transfer", err)
	}
}

func TestTransferList(t *testing.T) {
	cfg, cashDB := setupTransferTestData(t)
	defer cashDB.Close()

	got := captureListOutput(t, func() error {
		return Transfer(parser.ParsedCmdLine{Command: "transfer", Subcommand: "list"}, cfg, cashDB)
	})

	for _, value := range []string{"2026.05.1", "2026.05.2", "20.00", cfg.Default.Account, "savings"} {
		if !strings.Contains(got, value) {
			t.Fatalf("output missing %q:\n%s", value, got)
		}
	}
	if strings.Contains(got, "2026.04.1") {
		t.Fatalf("output contains unrelated transaction:\n%s", got)
	}
}

func TestTransferListJSON(t *testing.T) {
	cfg, cashDB := setupTransferTestData(t)
	defer cashDB.Close()

	got := captureListOutput(t, func() error {
		return Transfer(parser.ParsedCmdLine{
			Command:    "transfer",
			Subcommand: "list",
			Flags:      []parser.Arg{parser.ArgFlag{Key: "json", Value: parser.BoolItem{Raw: "true", Value: true}}},
		}, cfg, cashDB)
	})

	var envelope struct {
		Success bool                 `json:"success"`
		Type    string               `json:"type"`
		Count   int                  `json:"count"`
		Data    output.TransfersData `json:"data"`
	}
	if err := json.Unmarshal([]byte(got), &envelope); err != nil {
		t.Fatalf("JSON output is invalid: %v", err)
	}
	if !envelope.Success || envelope.Type != "transfers" || envelope.Count != 1 {
		t.Fatalf("JSON envelope = %#v", envelope)
	}
	if len(envelope.Data.Transfers) != 1 {
		t.Fatalf("len(transfers) = %d, want 1", len(envelope.Data.Transfers))
	}
	item := envelope.Data.Transfers[0]
	if item.FromID != "2026.05.1" || item.ToID != "2026.05.2" || item.Amount != 20 {
		t.Fatalf("transfer item = %#v", item)
	}
}

func TestTransferListEmpty(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	got := captureListOutput(t, func() error {
		return Transfer(parser.ParsedCmdLine{
			Command:    "transfer",
			Subcommand: "list",
			Flags:      []parser.Arg{parser.ArgFlag{Key: "json", Value: parser.BoolItem{Raw: "true", Value: true}}},
		}, cfg, cashDB)
	})

	var envelope struct {
		Success bool `json:"success"`
		Count   int  `json:"count"`
	}
	if err := json.Unmarshal([]byte(got), &envelope); err != nil {
		t.Fatalf("JSON output is invalid: %v", err)
	}
	if !envelope.Success || envelope.Count != 0 {
		t.Fatalf("JSON envelope = %#v, want empty transfer list", envelope)
	}
}
