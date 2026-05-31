package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestGroupsListSortedByNameAscending(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	if _, err := db.InsertTransactionGroup(cashDB, db.CreateTransactionGroupInput{Name: "zeta"}); err != nil {
		t.Fatalf("InsertTransactionGroup(zeta) returned error: %v", err)
	}
	if _, err := db.InsertTransactionGroup(cashDB, db.CreateTransactionGroupInput{Name: "alpha"}); err != nil {
		t.Fatalf("InsertTransactionGroup(alpha) returned error: %v", err)
	}

	mainAccount, err := db.GetAccountByName(cashDB, cfg.Default.Account)
	if err != nil {
		t.Fatalf("GetAccountByName returned error: %v", err)
	}
	placeID, err := db.InsertPlace(cashDB, db.CreatePlaceInput{Name: "Groups Test"})
	if err != nil {
		t.Fatalf("InsertPlace returned error: %v", err)
	}
	alpha, err := db.GetTransactionGroupByName(cashDB, "alpha")
	if err != nil {
		t.Fatalf("GetTransactionGroupByName(alpha) returned error: %v", err)
	}
	_, err = db.InsertTransaction(cashDB, db.CreateTransactionInput{Identifier: "2026.05.1", Amount: -10, Description: "A", Datetime: time.Date(2026, time.May, 27, 0, 0, 0, 0, time.UTC), AccountID: mainAccount.ID, PlaceID: &placeID, GroupID: &alpha.ID})
	if err != nil {
		t.Fatalf("InsertTransaction returned error: %v", err)
	}
	_, err = db.InsertTransaction(cashDB, db.CreateTransactionInput{Identifier: "2026.05.2", Amount: -5, Description: "B", Datetime: time.Date(2026, time.May, 29, 0, 0, 0, 0, time.UTC), AccountID: mainAccount.ID, PlaceID: &placeID, GroupID: &alpha.ID})
	if err != nil {
		t.Fatalf("InsertTransaction returned error: %v", err)
	}

	output := captureStdout(t, func() {
		err := Groups(parser.ParsedCmdLine{Command: "groups", Subcommand: "list"}, cfg, cashDB)
		if err != nil {
			t.Fatalf("Groups returned error: %v", err)
		}
	})

	alphaIndex := strings.Index(output, "alpha")
	zetaIndex := strings.Index(output, "zeta")
	if alphaIndex == -1 || zetaIndex == -1 {
		t.Fatalf("output missing expected names: %s", output)
	}
	if alphaIndex > zetaIndex {
		t.Fatalf("groups are not sorted by name ascending: %s", output)
	}
	if !strings.Contains(output, "-15.00") {
		t.Fatalf("output missing transaction sum for alpha: %s", output)
	}
	if !strings.Contains(output, "2") {
		t.Fatalf("output missing transaction count for alpha: %s", output)
	}
	if !strings.Contains(output, "2026-05-27") || !strings.Contains(output, "2026-05-29") {
		t.Fatalf("output missing start/end dates for alpha: %s", output)
	}
}
