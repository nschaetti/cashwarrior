package cmd

import (
	"testing"
	"time"

	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestAddCategory(t *testing.T) {
	_, cashDB := openTestDB(t)
	defer cashDB.Close()

	err := addCategory(parser.ParsedCmdLine{
		Command:    "categories",
		Subcommand: "add",
		Args: []parser.Token{
			{Kind: parser.TokenText, Raw: "travel"},
		},
	}, cashDB)
	if err != nil {
		t.Fatalf("addCategory returned error: %v", err)
	}

	category, err := db.GetCategoryByName(cashDB, "travel")
	if err != nil {
		t.Fatalf("GetCategoryByName returned error: %v", err)
	}
	if category.Name != "travel" {
		t.Fatalf("name = %q, want travel", category.Name)
	}
}

func TestModifyCategory(t *testing.T) {
	_, cashDB := openTestDB(t)
	defer cashDB.Close()

	_, err := db.InsertCategory(cashDB, db.CreateCategoryInput{Name: "travel"})
	if err != nil {
		t.Fatalf("InsertCategory returned error: %v", err)
	}
	_, err = db.InsertCategory(cashDB, db.CreateCategoryInput{Name: "lifestyle"})
	if err != nil {
		t.Fatalf("InsertCategory returned error: %v", err)
	}

	err = modifyCategory(parser.ParsedCmdLine{
		Command:    "categories",
		Subcommand: "modify",
		Args: []parser.Token{
			{Kind: parser.TokenText, Raw: "travel"},
			{Kind: parser.TokenAttribute, Key: "category", Value: "vacation", Raw: "category:vacation"},
			{Kind: parser.TokenAttribute, Key: "parent", Value: "lifestyle", Raw: "parent:lifestyle"},
		},
	}, cashDB)
	if err != nil {
		t.Fatalf("modifyCategory returned error: %v", err)
	}

	category, err := db.GetCategoryByName(cashDB, "vacation")
	if err != nil {
		t.Fatalf("GetCategoryByName returned error: %v", err)
	}
	parent, err := db.GetCategoryByName(cashDB, "lifestyle")
	if err != nil {
		t.Fatalf("GetCategoryByName(parent) returned error: %v", err)
	}
	if category.ParentID == nil || *category.ParentID != parent.ID {
		t.Fatalf("parentID = %v, want %d", category.ParentID, parent.ID)
	}
}

func TestDeleteCategoryRejectsLinkedTransactions(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	categoryID, err := db.InsertCategory(cashDB, db.CreateCategoryInput{Name: "travel"})
	if err != nil {
		t.Fatalf("InsertCategory returned error: %v", err)
	}
	mainAccount, err := db.GetAccountByName(cashDB, cfg.Default.Account)
	if err != nil {
		t.Fatalf("GetAccountByName returned error: %v", err)
	}
	placeID, err := db.InsertStore(cashDB, db.CreatePlaceInput{Name: "Category Delete Test"})
	if err != nil {
		t.Fatalf("InsertPlace returned error: %v", err)
	}
	_, err = db.InsertTransaction(cashDB, db.CreateTransactionInput{Identifier: "2026.05.1", Amount: -5, Description: "x", Date: time.Date(2026, time.May, 27, 12, 0, 0, 0, time.UTC), AccountID: mainAccount.ID, CategoryID: &categoryID, PlaceID: &placeID})
	if err != nil {
		t.Fatalf("InsertTransaction returned error: %v", err)
	}

	err = deleteCategory(parser.ParsedCmdLine{Command: "categories", Subcommand: "delete", Args: []parser.Token{{Kind: parser.TokenText, Raw: "travel"}}}, cashDB)
	if err == nil {
		t.Fatal("deleteCategory expected error, got nil")
	}
}

func TestDeleteCategoryDeletesEmptyCategoryAfterConfirmation(t *testing.T) {
	_, cashDB := openTestDB(t)
	defer cashDB.Close()

	_, err := db.InsertCategory(cashDB, db.CreateCategoryInput{Name: "travel"})
	if err != nil {
		t.Fatalf("InsertCategory returned error: %v", err)
	}

	withInput(t, "y\n", func() {
		err = deleteCategory(parser.ParsedCmdLine{Command: "categories", Subcommand: "delete", Args: []parser.Token{{Kind: parser.TokenText, Raw: "travel"}}}, cashDB)
	})
	if err != nil {
		t.Fatalf("deleteCategory returned error: %v", err)
	}

	_, err = db.GetCategoryByName(cashDB, "travel")
	if err == nil {
		t.Fatal("GetCategoryByName expected error, got nil")
	}
}
