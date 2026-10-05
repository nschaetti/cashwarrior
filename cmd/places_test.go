package cmd

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestPlacesAddAndDelete(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()
	parsed := parser.ParsedCmdLine{Command: "stores", Subcommand: "add", Args: []parser.Arg{testArg(t, "Migros")}, Flags: yesFlag()}
	if err := Places(parsed, cfg, cashDB); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetStoreByName(cashDB, "Migros"); err != nil {
		t.Fatal(err)
	}
	if err := Places(parsed, cfg, cashDB); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("duplicate: %v", err)
	}
	parsed.Subcommand = "rm"
	if err := Places(parsed, cfg, cashDB); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetStoreByName(cashDB, "Migros"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("deleted place: %v", err)
	}
	if err := Places(parsed, cfg, cashDB); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("missing place: %v", err)
	}
}

func TestPlacesAddJSONRequiresYes(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()
	err := Places(parser.ParsedCmdLine{Subcommand: "add", Args: []parser.Arg{testArg(t, "Migros")}, Flags: []parser.Arg{testArg(t, "--json")}}, cfg, cashDB)
	if err == nil || !strings.Contains(err.Error(), "--yes is required") {
		t.Fatalf("err = %v", err)
	}
	if exists, err := db.PlaceExists(cashDB, "Migros"); err != nil || exists {
		t.Fatalf("exists = %v, err = %v", exists, err)
	}
}

func TestPlacesDeleteGuards(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()
	placeID, err := db.InsertStore(cashDB, db.CreatePlaceInput{Name: "Migros"})
	if err != nil {
		t.Fatal(err)
	}
	account, err := db.GetAccountByName(cashDB, cfg.Default.Account)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.InsertTransaction(cashDB, db.CreateTransactionInput{Identifier: "2026.05.1", Amount: -5, Description: "x", Date: testTime(), AccountID: account.ID, PlaceID: &placeID}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, message string }{{"Migros", "has linked transactions"}, {"transfer", "cannot be deleted"}} {
		err := Places(parser.ParsedCmdLine{Subcommand: "delete", Args: []parser.Arg{testArg(t, tc.name)}, Flags: yesFlag()}, cfg, cashDB)
		if err == nil || !strings.Contains(err.Error(), tc.message) {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if _, err := db.GetStoreByName(cashDB, tc.name); err != nil {
			t.Fatal(err)
		}
	}
}

func TestPlacesRename(t *testing.T) {
	_, cashDB := openTestDB(t)
	defer cashDB.Close()

	if _, err := db.InsertStore(cashDB, db.CreatePlaceInput{Name: "Coop"}); err != nil {
		t.Fatalf("InsertPlace returned error: %v", err)
	}

	err := Places(parser.ParsedCmdLine{
		Command:    "places",
		Subcommand: "rename",
		Args:       []parser.Arg{testArg(t, "Coop"), testArg(t, "Migros")},
	}, config.GetDefaultConfig(), cashDB)
	if err != nil {
		t.Fatalf("Places(rename) returned error: %v", err)
	}

	if _, err := db.GetStoreByName(cashDB, "Migros"); err != nil {
		t.Fatalf("GetPlaceByName(Migros) returned error: %v", err)
	}
}
