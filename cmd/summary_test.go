package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestSummaryDaysShowsCountPerDay(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	mainAccount, err := db.GetAccountByName(cashDB, cfg.Default.Account)
	if err != nil {
		t.Fatalf("GetAccountByName returned error: %v", err)
	}
	placeID, err := db.InsertStore(cashDB, db.CreatePlaceInput{Name: "Summary Test"})
	if err != nil {
		t.Fatalf("InsertPlace returned error: %v", err)
	}

	_, err = db.InsertTransaction(cashDB, db.CreateTransactionInput{Identifier: "2026.05.1", Amount: -10, Description: "A", Date: time.Date(2026, time.May, 27, 10, 0, 0, 0, time.UTC), AccountID: mainAccount.ID, PlaceID: &placeID})
	if err != nil {
		t.Fatalf("InsertTransaction returned error: %v", err)
	}
	_, err = db.InsertTransaction(cashDB, db.CreateTransactionInput{Identifier: "2026.05.2", Amount: -5, Description: "B", Date: time.Date(2026, time.May, 27, 20, 0, 0, 0, time.UTC), AccountID: mainAccount.ID, PlaceID: &placeID})
	if err != nil {
		t.Fatalf("InsertTransaction returned error: %v", err)
	}
	_, err = db.InsertTransaction(cashDB, db.CreateTransactionInput{Identifier: "2026.05.3", Amount: -2, Description: "C", Date: time.Date(2026, time.May, 28, 9, 0, 0, 0, time.UTC), AccountID: mainAccount.ID, PlaceID: &placeID})
	if err != nil {
		t.Fatalf("InsertTransaction returned error: %v", err)
	}

	out := captureSummaryStdout(t, func() {
		err = Summary(parser.ParsedCmdLine{Command: "summary", Subcommand: "days"}, cfg, cashDB)
	})
	if err != nil {
		t.Fatalf("Summary returned error: %v", err)
	}
	if !strings.Contains(out, "2026-05-27") || !strings.Contains(out, "2026-05-28") {
		t.Fatalf("output missing dates: %s", out)
	}
	if !strings.Contains(out, "2") || !strings.Contains(out, "1") {
		t.Fatalf("output missing counts: %s", out)
	}
	if !strings.Contains(out, "Expenses (USD)") || !strings.Contains(out, "Income (USD)") || !strings.Contains(out, "Net (USD)") {
		t.Fatalf("output missing currency summary headers: %s", out)
	}
}

func captureSummaryStdout(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe returned error: %v", err)
	}
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = original

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, r)
	if err != nil {
		t.Fatalf("io.Copy returned error: %v", err)
	}
	_ = r.Close()
	return buf.String()
}
