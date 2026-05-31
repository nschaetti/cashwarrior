package cmd

import (
	"testing"
	"time"

	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestDeleteAndRestoreTransaction(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	mainAccount, err := db.GetAccountByName(cashDB, cfg.Default.Account)
	if err != nil {
		t.Fatalf("GetAccountByName returned error: %v", err)
	}

	placeID, err := db.InsertPlace(cashDB, db.CreatePlaceInput{Name: "Delete Test"})
	if err != nil {
		t.Fatalf("InsertPlace returned error: %v", err)
	}

	_, err = db.InsertTransaction(cashDB, db.CreateTransactionInput{
		Identifier:  "2026.05.1",
		Amount:      -10,
		Description: "Lunch",
		Datetime:    time.Date(2026, time.May, 27, 12, 0, 0, 0, time.UTC),
		AccountID:   mainAccount.ID,
		PlaceID:     &placeID,
	})
	if err != nil {
		t.Fatalf("InsertTransaction returned error: %v", err)
	}

	if err := Delete(parser.ParsedCmdLine{
		Command:    "delete",
		Subcommand: "default",
		Args:       []parser.Token{{Raw: "2026.05.1", Kind: parser.TokenID}},
	}, cfg, cashDB); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	active, err := db.ListTransactions(cashDB, nil, nil)
	if err != nil {
		t.Fatalf("ListTransactions returned error: %v", err)
	}
	if len(active) != 0 {
		t.Fatalf("len(active) = %d, want 0", len(active))
	}

	deleted, err := db.ListDeletedTransactions(cashDB, nil, nil)
	if err != nil {
		t.Fatalf("ListDeletedTransactions returned error: %v", err)
	}
	if len(deleted) != 1 {
		t.Fatalf("len(deleted) = %d, want 1", len(deleted))
	}

	if err := Restore(parser.ParsedCmdLine{
		Command:    "restore",
		Subcommand: "default",
		Args:       []parser.Token{{Raw: "2026.05.1", Kind: parser.TokenID}},
	}, cfg, cashDB); err != nil {
		t.Fatalf("Restore returned error: %v", err)
	}

	active, err = db.ListTransactions(cashDB, nil, nil)
	if err != nil {
		t.Fatalf("ListTransactions returned error: %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("len(active) = %d, want 1", len(active))
	}
}

func TestPurgeTransactionDeletesRegularTransaction(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	mainAccount, err := db.GetAccountByName(cashDB, cfg.Default.Account)
	if err != nil {
		t.Fatalf("GetAccountByName returned error: %v", err)
	}

	placeID, err := db.InsertPlace(cashDB, db.CreatePlaceInput{Name: "Purge Test"})
	if err != nil {
		t.Fatalf("InsertPlace returned error: %v", err)
	}

	_, err = db.InsertTransaction(cashDB, db.CreateTransactionInput{
		Identifier:  "2026.05.1",
		Amount:      -10,
		Description: "Lunch",
		Datetime:    time.Date(2026, time.May, 27, 12, 0, 0, 0, time.UTC),
		AccountID:   mainAccount.ID,
		PlaceID:     &placeID,
	})
	if err != nil {
		t.Fatalf("InsertTransaction returned error: %v", err)
	}

	if err := Purge(parser.ParsedCmdLine{
		Command:    "purge",
		Subcommand: "default",
		Args:       []parser.Token{{Raw: "2026.05.1", Kind: parser.TokenID}},
	}, cfg, cashDB); err != nil {
		t.Fatalf("Purge returned error: %v", err)
	}

	active, err := db.ListTransactions(cashDB, nil, nil)
	if err != nil {
		t.Fatalf("ListTransactions returned error: %v", err)
	}
	if len(active) != 0 {
		t.Fatalf("len(active) = %d, want 0", len(active))
	}
}

func TestPurgeTransactionDeletesEntireTransfer(t *testing.T) {
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
	transferPlaceID, err := db.InsertPlace(cashDB, db.CreatePlaceInput{Name: "transfer-purge"})
	if err != nil {
		t.Fatalf("InsertPlace returned error: %v", err)
	}

	fromTxID, err := db.InsertTransaction(cashDB, db.CreateTransactionInput{
		Identifier:  "2026.05.1",
		Type:        "transfer_out",
		Amount:      -20,
		Description: "Transfer",
		Datetime:    time.Date(2026, time.May, 27, 12, 0, 0, 0, time.UTC),
		AccountID:   fromAccount.ID,
		PlaceID:     &transferPlaceID,
	})
	if err != nil {
		t.Fatalf("InsertTransaction(from) returned error: %v", err)
	}
	toTxID, err := db.InsertTransaction(cashDB, db.CreateTransactionInput{
		Identifier:  "2026.05.2",
		Type:        "transfer_in",
		Amount:      20,
		Description: "Transfer",
		Datetime:    time.Date(2026, time.May, 27, 12, 0, 0, 0, time.UTC),
		AccountID:   toAccountID,
		PlaceID:     &transferPlaceID,
	})
	if err != nil {
		t.Fatalf("InsertTransaction(to) returned error: %v", err)
	}
	_, err = db.InsertTransfer(cashDB, db.CreateTransferInput{
		FromTransactionID: fromTxID,
		ToTransactionID:   toTxID,
		FromAccountID:     fromAccount.ID,
		ToAccountID:       toAccountID,
		Amount:            20,
	})
	if err != nil {
		t.Fatalf("InsertTransfer returned error: %v", err)
	}

	if err := Purge(parser.ParsedCmdLine{
		Command:    "purge",
		Subcommand: "default",
		Args:       []parser.Token{{Raw: "2026.05.1", Kind: parser.TokenID}},
	}, cfg, cashDB); err != nil {
		t.Fatalf("Purge returned error: %v", err)
	}

	active, err := db.ListTransactions(cashDB, nil, nil)
	if err != nil {
		t.Fatalf("ListTransactions returned error: %v", err)
	}
	if len(active) != 0 {
		t.Fatalf("len(active) = %d, want 0", len(active))
	}

	transfers, err := db.ListTransfers(cashDB, db.TransferListFilter{})
	if err != nil {
		t.Fatalf("ListTransfers returned error: %v", err)
	}
	if len(transfers) != 0 {
		t.Fatalf("len(transfers) = %d, want 0", len(transfers))
	}
}

func TestDeleteTransferMarksBothTransactionsAndTransferDeleted(t *testing.T) {
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
	transferPlaceID, err := db.InsertPlace(cashDB, db.CreatePlaceInput{Name: "transfer-delete"})
	if err != nil {
		t.Fatalf("InsertPlace returned error: %v", err)
	}

	fromTxID, err := db.InsertTransaction(cashDB, db.CreateTransactionInput{Identifier: "2026.05.1", Type: "transfer_out", Amount: -20, Description: "Transfer", Datetime: time.Date(2026, time.May, 27, 12, 0, 0, 0, time.UTC), AccountID: fromAccount.ID, PlaceID: &transferPlaceID})
	if err != nil {
		t.Fatalf("InsertTransaction(from) returned error: %v", err)
	}
	toTxID, err := db.InsertTransaction(cashDB, db.CreateTransactionInput{Identifier: "2026.05.2", Type: "transfer_in", Amount: 20, Description: "Transfer", Datetime: time.Date(2026, time.May, 27, 12, 0, 0, 0, time.UTC), AccountID: toAccountID, PlaceID: &transferPlaceID})
	if err != nil {
		t.Fatalf("InsertTransaction(to) returned error: %v", err)
	}
	transferID, err := db.InsertTransfer(cashDB, db.CreateTransferInput{FromTransactionID: fromTxID, ToTransactionID: toTxID, FromAccountID: fromAccount.ID, ToAccountID: toAccountID, Amount: 20})
	if err != nil {
		t.Fatalf("InsertTransfer returned error: %v", err)
	}

	if err := Delete(parser.ParsedCmdLine{Command: "delete", Subcommand: "default", Args: []parser.Token{{Raw: "2026.05.1", Kind: parser.TokenID}}}, cfg, cashDB); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	fromTx, err := db.GetTransactionByID(cashDB, fromTxID)
	if err != nil {
		t.Fatalf("GetTransactionByID(from) returned error: %v", err)
	}
	if !fromTx.Deleted {
		t.Fatal("from transfer transaction should be deleted")
	}
	toTx, err := db.GetTransactionByID(cashDB, toTxID)
	if err != nil {
		t.Fatalf("GetTransactionByID(to) returned error: %v", err)
	}
	if !toTx.Deleted {
		t.Fatal("to transfer transaction should be deleted")
	}

	transfer, err := db.GetTransferByID(cashDB, transferID)
	if err != nil {
		t.Fatalf("GetTransferByID returned error: %v", err)
	}
	if !transfer.Deleted {
		t.Fatal("transfer should be marked deleted")
	}
}

func TestRestoreTransferRestoresBothTransactionsAndTransfer(t *testing.T) {
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
	transferPlaceID, err := db.InsertPlace(cashDB, db.CreatePlaceInput{Name: "transfer-restore"})
	if err != nil {
		t.Fatalf("InsertPlace returned error: %v", err)
	}

	fromTxID, err := db.InsertTransaction(cashDB, db.CreateTransactionInput{Identifier: "2026.05.1", Type: "transfer_out", Amount: -20, Description: "Transfer", Datetime: time.Date(2026, time.May, 27, 12, 0, 0, 0, time.UTC), AccountID: fromAccount.ID, PlaceID: &transferPlaceID})
	if err != nil {
		t.Fatalf("InsertTransaction(from) returned error: %v", err)
	}
	toTxID, err := db.InsertTransaction(cashDB, db.CreateTransactionInput{Identifier: "2026.05.2", Type: "transfer_in", Amount: 20, Description: "Transfer", Datetime: time.Date(2026, time.May, 27, 12, 0, 0, 0, time.UTC), AccountID: toAccountID, PlaceID: &transferPlaceID})
	if err != nil {
		t.Fatalf("InsertTransaction(to) returned error: %v", err)
	}
	_, err = db.InsertTransfer(cashDB, db.CreateTransferInput{FromTransactionID: fromTxID, ToTransactionID: toTxID, FromAccountID: fromAccount.ID, ToAccountID: toAccountID, Amount: 20})
	if err != nil {
		t.Fatalf("InsertTransfer returned error: %v", err)
	}

	if err := Delete(parser.ParsedCmdLine{Command: "delete", Subcommand: "default", Args: []parser.Token{{Raw: "2026.05.1", Kind: parser.TokenID}}}, cfg, cashDB); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if err := Restore(parser.ParsedCmdLine{Command: "restore", Subcommand: "default", Args: []parser.Token{{Raw: "2026.05.1", Kind: parser.TokenID}}}, cfg, cashDB); err != nil {
		t.Fatalf("Restore returned error: %v", err)
	}

	fromTx, err := db.GetTransactionByID(cashDB, fromTxID)
	if err != nil {
		t.Fatalf("GetTransactionByID(from) returned error: %v", err)
	}
	if fromTx.Deleted {
		t.Fatal("from transfer transaction should be restored")
	}
	toTx, err := db.GetTransactionByID(cashDB, toTxID)
	if err != nil {
		t.Fatalf("GetTransactionByID(to) returned error: %v", err)
	}
	if toTx.Deleted {
		t.Fatal("to transfer transaction should be restored")
	}

	transfer, err := db.GetTransferByTransactionID(cashDB, fromTxID)
	if err != nil {
		t.Fatalf("GetTransferByTransactionID returned error: %v", err)
	}
	if transfer.Deleted {
		t.Fatal("transfer should be restored")
	}
}
