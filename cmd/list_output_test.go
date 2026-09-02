package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/nschaetti/cashwarrior/internal/output"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestRenderListDataJSONContainsTableData(t *testing.T) {
	data := output.ListData{
		Transactions: []output.ListTransaction{{
			ID:          7,
			Identifier:  "2026.05.7",
			Type:        "expense",
			Amount:      -12.5,
			Currency:    "USD",
			Description: "Coffee",
			Account:     "main",
			Vendor:      "Cafe",
			Category:    "food",
			Date:        parseListOutputDate(t, "2026-05-27T00:00:00Z"),
		}},
		Summary: output.ListSummary{
			ByCurrency: []output.ListCurrencySummary{{Currency: "USD", Expenses: -12.5, Net: -12.5}},
			ByAccount:  []output.ListAccountSummary{{Account: "main", Expenses: -12.5, Net: -12.5}},
		},
	}

	jsonOutput := captureListOutput(t, func() error {
		return renderListData(parser.ParsedCmdLine{Flags: []parser.Arg{
			parser.ArgFlag{Key: "json", Value: parser.BoolItem{Raw: "true", Value: true}},
		}}, data)
	})
	var envelope struct {
		Success bool            `json:"success"`
		Type    string          `json:"type"`
		Data    output.ListData `json:"data"`
		Count   int             `json:"count"`
	}
	if err := json.Unmarshal([]byte(jsonOutput), &envelope); err != nil {
		t.Fatalf("JSON output is invalid: %v", err)
	}
	if !envelope.Success || envelope.Type != "transactions" || envelope.Count != 1 {
		t.Fatalf("JSON envelope = %#v", envelope)
	}

	tableOutput := captureListOutput(t, func() error {
		return renderListData(parser.ParsedCmdLine{}, data)
	})
	for _, value := range []string{"2026.05.7", "main", "-12.50", "Coffee", "food", "USD"} {
		if !strings.Contains(tableOutput, value) {
			t.Fatalf("table output missing %q:\n%s", value, tableOutput)
		}
	}
}

func captureListOutput(t *testing.T, fn func() error) string {
	t.Helper()
	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe returned error: %v", err)
	}
	os.Stdout = writer
	err = fn()
	_ = writer.Close()
	os.Stdout = original
	if err != nil {
		t.Fatalf("render returned error: %v", err)
	}

	var output bytes.Buffer
	if _, err := io.Copy(&output, reader); err != nil {
		t.Fatalf("io.Copy returned error: %v", err)
	}
	_ = reader.Close()
	return output.String()
}

func parseListOutputDate(t *testing.T, value string) (date time.Time) {
	t.Helper()
	date, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("time.Parse returned error: %v", err)
	}
	return date
}
