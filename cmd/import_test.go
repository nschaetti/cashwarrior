package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func writeCSVFile(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "import.csv")
	if err := os.WriteFile(path, []byte(strings.TrimSpace(contents)+"\n"), 0644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	return path
}

func TestImportTransactionsCSV(t *testing.T) {
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
	groupID, err := db.InsertTransactionGroup(cashDB, db.CreateTransactionGroupInput{Name: "weekly"})
	if err != nil {
		t.Fatalf("InsertTransactionGroup returned error: %v", err)
	}

	_ = mainAccount
	_ = placeID
	_ = categoryID
	_ = groupID

	path := writeCSVFile(t, `identifier,type,amount,description,datetime,account,category,place,group,deleted
2026.05.1,expense,-12.50,Lunch,2026-05-27,main,groceries,Migros,weekly,false`)

	if err := Import(parser.ParsedCmdLine{Command: "import", Subcommand: "default", Args: []parser.Token{{Kind: parser.TokenText, Raw: path}}}, cfg, cashDB); err != nil {
		t.Fatalf("Import returned error: %v", err)
	}

	transaction, err := db.GetTransactionByIdentifier(cashDB, "2026.05.1")
	if err != nil {
		t.Fatalf("GetTransactionByIdentifier returned error: %v", err)
	}
	if transaction.Description != "Lunch" {
		t.Fatalf("transaction.Description = %q, want Lunch", transaction.Description)
	}
}

func TestImportTransactionsRollbackOnMissingAccount(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	placeID, err := db.InsertPlace(cashDB, db.CreatePlaceInput{Name: "Migros"})
	if err != nil {
		t.Fatalf("InsertPlace returned error: %v", err)
	}
	_ = placeID

	path := writeCSVFile(t, `identifier,type,amount,description,datetime,account,place,deleted
2026.05.1,expense,-12.50,Lunch,2026-05-27,missing,Migros,false`)

	err = Import(parser.ParsedCmdLine{Command: "import", Subcommand: "default", Args: []parser.Token{{Kind: parser.TokenText, Raw: path}}}, cfg, cashDB)
	if err == nil {
		t.Fatal("Import expected error, got nil")
	}

	transactions, err := db.ListTransactions(cashDB, nil, nil)
	if err != nil {
		t.Fatalf("ListTransactions returned error: %v", err)
	}
	if len(transactions) != 0 {
		t.Fatalf("len(transactions) = %d, want 0", len(transactions))
	}
}

func TestImportTransactionTagsAndTransfersCSV(t *testing.T) {
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
	placeID, err := db.InsertPlace(cashDB, db.CreatePlaceInput{Name: "transfer-import"})
	if err != nil {
		t.Fatalf("InsertPlace returned error: %v", err)
	}
	_, err = db.InsertTag(cashDB, db.CreateTagInput{Name: "food"})
	if err != nil {
		t.Fatalf("InsertTag returned error: %v", err)
	}
	_, err = db.InsertTransaction(cashDB, db.CreateTransactionInput{Identifier: "2026.05.1", Type: "transfer_out", Amount: -20, Description: "Transfer", Datetime: time.Date(2026, time.May, 27, 12, 0, 0, 0, time.UTC), AccountID: fromAccount.ID, PlaceID: &placeID})
	if err != nil {
		t.Fatalf("InsertTransaction(from) returned error: %v", err)
	}
	_, err = db.InsertTransaction(cashDB, db.CreateTransactionInput{Identifier: "2026.05.2", Type: "transfer_in", Amount: 20, Description: "Transfer", Datetime: time.Date(2026, time.May, 27, 12, 0, 0, 0, time.UTC), AccountID: toAccountID, PlaceID: &placeID})
	if err != nil {
		t.Fatalf("InsertTransaction(to) returned error: %v", err)
	}

	tagPath := writeCSVFile(t, `transaction,tag
2026.05.1,food`)
	if err := Import(parser.ParsedCmdLine{Command: "import", Subcommand: "default", Args: []parser.Token{{Kind: parser.TokenText, Raw: tagPath}}}, cfg, cashDB); err != nil {
		t.Fatalf("Import(transaction_tags) returned error: %v", err)
	}

	transaction, err := db.GetTransactionByIdentifier(cashDB, "2026.05.1")
	if err != nil {
		t.Fatalf("GetTransactionByIdentifier returned error: %v", err)
	}
	tags, err := db.ListTagsByTransactionID(cashDB, transaction.ID)
	if err != nil {
		t.Fatalf("ListTagsByTransactionID returned error: %v", err)
	}
	if len(tags) != 1 || tags[0].Name != "food" {
		t.Fatalf("tags = %v, want [food]", tags)
	}

	transferPath := writeCSVFile(t, `from_transaction,to_transaction,from_account,to_account,amount
2026.05.1,2026.05.2,main,savings,20`)
	if err := Import(parser.ParsedCmdLine{Command: "import", Subcommand: "default", Args: []parser.Token{{Kind: parser.TokenText, Raw: transferPath}}}, cfg, cashDB); err != nil {
		t.Fatalf("Import(transfers) returned error: %v", err)
	}

	transfers, err := db.ListTransfers(cashDB, db.TransferListFilter{})
	if err != nil {
		t.Fatalf("ListTransfers returned error: %v", err)
	}
	if len(transfers) != 1 {
		t.Fatalf("len(transfers) = %d, want 1", len(transfers))
	}
}
