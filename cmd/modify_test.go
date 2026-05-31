package cmd

import (
	"database/sql"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
	_ "modernc.org/sqlite"
)

func TestModifyUpdatesMatchingTransactionsAfterConfirmation(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	mainAccount, err := db.GetAccountByName(cashDB, cfg.Default.Account)
	if err != nil {
		t.Fatalf("GetAccountByName returned error: %v", err)
	}

	migrosID, err := db.InsertPlace(cashDB, db.CreatePlaceInput{Name: "Migros"})
	if err != nil {
		t.Fatalf("InsertPlace(Migros) returned error: %v", err)
	}
	coopID, err := db.InsertPlace(cashDB, db.CreatePlaceInput{Name: "Coop"})
	if err != nil {
		t.Fatalf("InsertPlace(Coop) returned error: %v", err)
	}

	_, err = db.InsertTransaction(cashDB, db.CreateTransactionInput{
		Identifier:  "2026.05.1",
		Amount:      -10,
		Description: "Lunch",
		Datetime:    time.Date(2026, time.May, 27, 12, 0, 0, 0, time.UTC),
		AccountID:   mainAccount.ID,
		PlaceID:     &migrosID,
	})
	if err != nil {
		t.Fatalf("InsertTransaction(first) returned error: %v", err)
	}
	_, err = db.InsertTransaction(cashDB, db.CreateTransactionInput{
		Identifier:  "2026.05.2",
		Amount:      -20,
		Description: "Dinner",
		Datetime:    time.Date(2026, time.May, 28, 12, 0, 0, 0, time.UTC),
		AccountID:   mainAccount.ID,
		PlaceID:     &migrosID,
	})
	if err != nil {
		t.Fatalf("InsertTransaction(second) returned error: %v", err)
	}

	parsed := parser.ParsedCmdLine{
		Command: "modify",
		Filters: []parser.Token{{Kind: parser.TokenAttribute, Key: "date", Value: "2026-05-27", Raw: "date:2026-05-27"}},
		Args:    []parser.Token{{Kind: parser.TokenAttribute, Key: "store", Value: "Coop", Raw: "store:Coop"}},
	}

	withInput(t, "y\n", func() {
		if err := Modify(parsed, cfg, cashDB); err != nil {
			t.Fatalf("Modify returned error: %v", err)
		}
	})

	first, err := db.GetTransactionByIdentifier(cashDB, "2026.05.1")
	if err != nil {
		t.Fatalf("GetTransactionByIdentifier(first) returned error: %v", err)
	}
	second, err := db.GetTransactionByIdentifier(cashDB, "2026.05.2")
	if err != nil {
		t.Fatalf("GetTransactionByIdentifier(second) returned error: %v", err)
	}

	if first.PlaceID == nil || *first.PlaceID != coopID {
		t.Fatalf("first.PlaceID = %v, want %d", first.PlaceID, coopID)
	}
	if second.PlaceID == nil || *second.PlaceID != migrosID {
		t.Fatalf("second.PlaceID = %v, want %d", second.PlaceID, migrosID)
	}
}

func TestParseTransactionModificationsRejectsIdentifier(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	parsed := parser.ParsedCmdLine{
		Command: "modify",
		Args:    []parser.Token{{Kind: parser.TokenAttribute, Key: "identifier", Value: "2026.05.99", Raw: "identifier:2026.05.99"}},
	}

	_, err := parseTransactionModifications(parsed, cfg, cashDB)
	if err == nil || err.Error() != "identifier is not modifiable" {
		t.Fatalf("err = %v, want identifier is not modifiable", err)
	}
}

func TestValidateTransactionModificationTargetsRejectsTransferAccountUpdate(t *testing.T) {
	accountID := int64(5)
	err := validateTransactionModificationTargets([]db.Transaction{{Type: "transfer_out"}}, transactionModifications{AccountID: &accountID})
	if err == nil || err.Error() != "account cannot be modified for transfer transactions" {
		t.Fatalf("err = %v, want account cannot be modified for transfer transactions", err)
	}
}

func TestMergeTransactionDatetimeKeepsUnchangedParts(t *testing.T) {
	original := time.Date(2026, time.May, 27, 14, 35, 42, 0, time.UTC)
	newDate := time.Date(2026, time.June, 2, 0, 0, 0, 0, time.UTC)
	newTime := time.Date(0, time.January, 1, 8, 15, 0, 0, time.UTC)

	updated, err := mergeTransactionDatetime(original, transactionModifications{Date: &newDate, Time: &newTime})
	if err != nil {
		t.Fatalf("mergeTransactionDatetime returned error: %v", err)
	}

	want := time.Date(2026, time.June, 2, 8, 15, 0, 0, time.UTC)
	if !updated.Equal(want) {
		t.Fatalf("updated = %v, want %v", updated, want)
	}
}

func TestModifyAddsAndRemovesTransactionTags(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	mainAccount, err := db.GetAccountByName(cashDB, cfg.Default.Account)
	if err != nil {
		t.Fatalf("GetAccountByName returned error: %v", err)
	}

	placeID, err := db.InsertPlace(cashDB, db.CreatePlaceInput{Name: "Tag Modify Test"})
	if err != nil {
		t.Fatalf("InsertPlace returned error: %v", err)
	}
	foodTagID, err := db.InsertTag(cashDB, db.CreateTagInput{Name: "food"})
	if err != nil {
		t.Fatalf("InsertTag(food) returned error: %v", err)
	}
	travelTagID, err := db.InsertTag(cashDB, db.CreateTagInput{Name: "travel"})
	if err != nil {
		t.Fatalf("InsertTag(travel) returned error: %v", err)
	}

	transactionID, err := db.InsertTransaction(cashDB, db.CreateTransactionInput{
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
	if err := db.InsertTransactionTag(cashDB, transactionID, travelTagID); err != nil {
		t.Fatalf("InsertTransactionTag returned error: %v", err)
	}

	parsed := parser.ParsedCmdLine{
		Command:    "modify",
		Subcommand: "default",
		Filters:    []parser.Token{{Kind: parser.TokenText, Raw: "T2026.05.1"}},
		Args: []parser.Token{
			{Kind: parser.TokenTag, Raw: "@food"},
			{Kind: parser.TokenTagNegative, Raw: "-@travel"},
		},
	}

	withInput(t, "y\n", func() {
		if err := Modify(parsed, cfg, cashDB); err != nil {
			t.Fatalf("Modify returned error: %v", err)
		}
	})

	tags, err := db.ListTagsByTransactionID(cashDB, transactionID)
	if err != nil {
		t.Fatalf("ListTagsByTransactionID returned error: %v", err)
	}

	tagNames := make([]string, 0, len(tags))
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}
	if !slices.Equal(tagNames, []string{"food"}) {
		t.Fatalf("tagNames = %v, want [food]", tagNames)
	}

	exists, err := db.TransactionTagExists(cashDB, transactionID, foodTagID)
	if err != nil {
		t.Fatalf("TransactionTagExists(food) returned error: %v", err)
	}
	if !exists {
		t.Fatal("food tag link missing after modify")
	}
}

func TestParseTransactionModificationsRejectsUnknownTag(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	parsed := parser.ParsedCmdLine{
		Command: "modify",
		Args:    []parser.Token{{Kind: parser.TokenTagNegative, Raw: "-@missing"}},
	}

	modifications, err := parseTransactionModifications(parsed, cfg, cashDB)
	if err != nil {
		t.Fatalf("parseTransactionModifications returned error: %v", err)
	}
	if len(modifications.DropTagIDs) != 0 {
		t.Fatalf("DropTagIDs = %v, want empty", modifications.DropTagIDs)
	}
}

func TestParseTransactionModificationsCreatesMissingAddedTag(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	parsed := parser.ParsedCmdLine{
		Command: "modify",
		Args:    []parser.Token{{Kind: parser.TokenTag, Raw: "@missing"}},
	}

	modifications, err := parseTransactionModifications(parsed, cfg, cashDB)
	if err != nil {
		t.Fatalf("parseTransactionModifications returned error: %v", err)
	}
	if len(modifications.AddTagIDs) != 1 {
		t.Fatalf("len(AddTagIDs) = %d, want 1", len(modifications.AddTagIDs))
	}

	tag, err := db.GetTagByName(cashDB, "missing")
	if err != nil {
		t.Fatalf("GetTagByName returned error: %v", err)
	}
	if tag.ID != modifications.AddTagIDs[0] {
		t.Fatalf("tag.ID = %d, want %d", tag.ID, modifications.AddTagIDs[0])
	}
}

func TestModifyUpdatesAmountWithIDFilter(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	mainAccount, err := db.GetAccountByName(cashDB, cfg.Default.Account)
	if err != nil {
		t.Fatalf("GetAccountByName returned error: %v", err)
	}
	placeID, err := db.InsertPlace(cashDB, db.CreatePlaceInput{Name: "Amount Modify Test"})
	if err != nil {
		t.Fatalf("InsertPlace returned error: %v", err)
	}

	_, err = db.InsertTransaction(cashDB, db.CreateTransactionInput{
		Identifier:  "2026.05.1",
		Amount:      -10,
		Description: "Lunch",
		Datetime:    time.Date(2026, time.May, 27, 0, 0, 0, 0, time.UTC),
		AccountID:   mainAccount.ID,
		PlaceID:     &placeID,
	})
	if err != nil {
		t.Fatalf("InsertTransaction returned error: %v", err)
	}

	parsed := parser.ParsedCmdLine{
		Command: "modify",
		Filters: []parser.Token{{Kind: parser.TokenID, Raw: "2026.05.1"}},
		Args:    []parser.Token{{Kind: parser.TokenAttribute, Key: "amount", Value: "-3.60", Raw: "amount:-3.60"}},
	}

	withInput(t, "y\n", func() {
		if err := Modify(parsed, cfg, cashDB); err != nil {
			t.Fatalf("Modify returned error: %v", err)
		}
	})

	transaction, err := db.GetTransactionByIdentifier(cashDB, "2026.05.1")
	if err != nil {
		t.Fatalf("GetTransactionByIdentifier returned error: %v", err)
	}
	if transaction.Amount != -3.60 {
		t.Fatalf("transaction.Amount = %v, want -3.60", transaction.Amount)
	}
}

func openTestDB(t *testing.T) (config.Config, *sql.DB) {
	t.Helper()

	cfg := config.GetDefaultConfig()
	dbConn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open returned error: %v", err)
	}
	if err := db.Init(dbConn, cfg); err != nil {
		_ = dbConn.Close()
		t.Fatalf("db.Init returned error: %v", err)
	}

	return cfg, dbConn
}

func withInput(t *testing.T, input string, fn func()) {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe returned error: %v", err)
	}

	oldStdin := os.Stdin
	os.Stdin = reader
	defer func() {
		os.Stdin = oldStdin
		_ = reader.Close()
	}()

	if _, err := writer.WriteString(input); err != nil {
		_ = writer.Close()
		t.Fatalf("WriteString returned error: %v", err)
	}
	_ = writer.Close()

	fn()
}
