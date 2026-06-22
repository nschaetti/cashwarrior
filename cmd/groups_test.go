package cmd

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func setupGroupsTestData(t *testing.T) (config.Config, *sql.DB, parser.ParsedCmdLine) {
	t.Helper()

	cfg, cashDB := openTestDB(t)

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
	placeID, err := db.InsertStore(cashDB, db.CreatePlaceInput{Name: "Groups Test"})
	if err != nil {
		t.Fatalf("InsertPlace returned error: %v", err)
	}
	alpha, err := db.GetGroupByName(cashDB, "alpha")
	if err != nil {
		t.Fatalf("GetTransactionGroupByName(alpha) returned error: %v", err)
	}
	zeta, err := db.GetGroupByName(cashDB, "zeta")
	if err != nil {
		t.Fatalf("GetTransactionGroupByName(zeta) returned error: %v", err)
	}

	_, err = db.InsertTransaction(cashDB, db.CreateTransactionInput{Identifier: "2026.05.1", Amount: -10, Description: "A", Date: time.Date(2026, time.May, 27, 0, 0, 0, 0, time.UTC), AccountID: mainAccount.ID, PlaceID: &placeID, GroupID: &alpha.ID})
	if err != nil {
		t.Fatalf("InsertTransaction returned error: %v", err)
	}
	_, err = db.InsertTransaction(cashDB, db.CreateTransactionInput{Identifier: "2026.05.2", Amount: -5, Description: "B", Date: time.Date(2026, time.May, 29, 0, 0, 0, 0, time.UTC), AccountID: mainAccount.ID, PlaceID: &placeID, GroupID: &alpha.ID})
	if err != nil {
		t.Fatalf("InsertTransaction returned error: %v", err)
	}
	_, err = db.InsertTransaction(cashDB, db.CreateTransactionInput{Identifier: "2026.05.3", Amount: -7, Description: "C", Date: time.Date(2026, time.May, 26, 0, 0, 0, 0, time.UTC), AccountID: mainAccount.ID, PlaceID: &placeID, GroupID: &zeta.ID})
	if err != nil {
		t.Fatalf("InsertTransaction returned error: %v", err)
	}

	return cfg, cashDB, parser.ParsedCmdLine{Command: "groups", Subcommand: "list"}
}

func TestGroupsListSortedByNameAscending(t *testing.T) {
	cfg, cashDB, parsed := setupGroupsTestData(t)
	defer cashDB.Close()

	output := captureStdout(t, func() {
		err := Groups(parsed, cfg, cashDB)
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

func TestGroupsListSortedByStartDate(t *testing.T) {
	cfg, cashDB, parsed := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed.Filters = []parser.Token{{Raw: "order:start_date", Kind: parser.TokenAttribute, Key: "order", Value: "start_date"}}
	output := captureStdout(t, func() {
		err := Groups(parsed, cfg, cashDB)
		if err != nil {
			t.Fatalf("Groups returned error: %v", err)
		}
	})

	zetaIndex := strings.Index(output, "zeta")
	alphaIndex := strings.Index(output, "alpha")
	if zetaIndex == -1 || alphaIndex == -1 {
		t.Fatalf("output missing expected names: %s", output)
	}
	if zetaIndex > alphaIndex {
		t.Fatalf("groups are not sorted by start_date ascending: %s", output)
	}
}

func TestGroupsListSortedByEndDateDescending(t *testing.T) {
	cfg, cashDB, parsed := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed.Filters = []parser.Token{
		{Raw: "order:end_date", Kind: parser.TokenAttribute, Key: "order", Value: "end_date"},
		{Raw: "desc:true", Kind: parser.TokenAttribute, Key: "desc", Value: "true"},
	}
	output := captureStdout(t, func() {
		err := Groups(parsed, cfg, cashDB)
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
		t.Fatalf("groups are not sorted by end_date descending: %s", output)
	}
}

func TestGroupsListRejectsUnsupportedOrder(t *testing.T) {
	cfg, cashDB, parsed := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed.Filters = []parser.Token{{Raw: "order:amount", Kind: parser.TokenAttribute, Key: "order", Value: "amount"}}
	err := Groups(parsed, cfg, cashDB)
	if err == nil || err.Error() != "unsupported groups order field amount" {
		t.Fatalf("err = %v, want unsupported groups order field amount", err)
	}
}
