package cmd

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/output"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestSumFilteredTotals(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()
	main, err := db.GetAccountByName(cashDB, cfg.Default.Account)
	if err != nil {
		t.Fatal(err)
	}
	chfID, err := db.InsertAccount(cashDB, db.CreateAccountInput{Name: "swiss", Currency: "CHF"})
	if err != nil {
		t.Fatal(err)
	}
	place, err := db.GetStoreByName(cashDB, "transfer")
	if err != nil {
		t.Fatal(err)
	}
	tagID, err := db.InsertTag(cashDB, db.CreateTagInput{Name: "courses"})
	if err != nil {
		t.Fatal(err)
	}
	for i, tc := range []struct {
		kind            string
		amount          float64
		account         int64
		month           time.Month
		deleted, tagged bool
	}{
		{"income", 100, main.ID, time.May, false, false},
		{"expense", -25, main.ID, time.June, false, true},
		{"income", 40, chfID, time.June, false, false},
		{"expense", -10, chfID, time.June, false, true},
		{"transfer_out", -500, main.ID, time.June, false, false},
		{"transfer_in", 500, chfID, time.June, false, false},
		{"expense", -999, main.ID, time.June, true, false},
	} {
		id, err := db.InsertTransaction(cashDB, db.CreateTransactionInput{Identifier: fmt.Sprintf("2026.%02d.%d", tc.month, i+1), Type: tc.kind, Amount: tc.amount, Date: time.Date(2026, tc.month, 15, 12, 0, 0, 0, time.UTC), AccountID: tc.account, PlaceID: &place.ID})
		if err != nil {
			t.Fatal(err)
		}
		if tc.deleted {
			if err := db.UpdateTransactionDeleted(cashDB, id, true); err != nil {
				t.Fatal(err)
			}
		}
		if tc.tagged {
			if err := db.InsertTransactionTag(cashDB, id, tagID); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, tc := range []struct {
		name    string
		filters []string
		count   int
		totals  []output.ListCurrencySummary
	}{
		{"all", nil, 4, []output.ListCurrencySummary{{Currency: "CHF", Income: 40, Expenses: -10, Net: 30}, {Currency: "USD", Income: 100, Expenses: -25, Net: 75}}},
		{"date", []string{"date:2026-06-01..2026-06-30"}, 3, []output.ListCurrencySummary{{Currency: "CHF", Income: 40, Expenses: -10, Net: 30}, {Currency: "USD", Expenses: -25, Net: -25}}},
		{"account", []string{"account:swiss"}, 2, []output.ListCurrencySummary{{Currency: "CHF", Income: 40, Expenses: -10, Net: 30}}},
		{"tag", []string{"@courses"}, 2, []output.ListCurrencySummary{{Currency: "CHF", Expenses: -10, Net: -10}, {Currency: "USD", Expenses: -25, Net: -25}}},
		{"empty", []string{"date:2025-01-01..2025-01-31"}, 0, []output.ListCurrencySummary{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := append(append([]string{}, tc.filters...), "sum", "--json")
			parsed, err := parser.ParseAndValidateCmdLine(args, cfg)
			if err != nil {
				t.Fatal(err)
			}
			text := captureListOutput(t, func() error {
				return Sum(parsed, cfg, cashDB)
			})
			var result struct {
				Success bool   `json:"success"`
				Type    string `json:"type"`
				Count   int    `json:"count"`
				Data    struct {
					Transactions int                          `json:"transactions"`
					Currencies   []output.ListCurrencySummary `json:"currencies"`
				} `json:"data"`
			}
			if err := json.Unmarshal([]byte(text), &result); err != nil {
				t.Fatal(err)
			}
			if !result.Success || result.Type != "sum" || result.Count != tc.count || result.Data.Transactions != tc.count || !reflect.DeepEqual(result.Data.Currencies, tc.totals) {
				t.Fatalf("result = %#v, want count %d, totals %#v", result, tc.count, tc.totals)
			}
		})
	}
	text := captureListOutput(t, func() error {
		return Sum(parser.ParsedCmdLine{}, cfg, cashDB)
	})
	for _, want := range []string{"Currency", "Income", "Expenses", "Net", "CHF", "USD", "Transactions: 4"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %s: %s", want, text)
		}
	}
}
