package cmd

import (
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestInitReportsInitializedDatabase(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()
	parsed, err := parser.ParseAndValidateCmdLine([]string{"init"}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	text := captureListOutput(t, func() error { return Dispatch(parsed, cfg, cashDB) })
	if !strings.Contains(text, "Database initialized at "+cfg.Database+" (1 accounts)") {
		t.Fatalf("output = %s", text)
	}
	parsed.Flags = []parser.Arg{testArg(t, "--json")}
	text = captureListOutput(t, func() error { return Dispatch(parsed, cfg, cashDB) })
	var result struct {
		Success bool   `json:"success"`
		Type    string `json:"type"`
		Count   int    `json:"count"`
		Data    struct {
			Database string `json:"database"`
			Accounts int    `json:"accounts"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		t.Fatal(err)
	}
	if !result.Success || result.Type != "init" || result.Count != 1 || result.Data.Database != cfg.Database || result.Data.Accounts != 1 {
		t.Fatalf("result = %#v", result)
	}
}

func TestInitRejectsUninitializedDatabase(t *testing.T) {
	cashDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer cashDB.Close()
	if err := Init(parser.ParsedCmdLine{}, config.GetDefaultConfig(), cashDB); err == nil || !strings.Contains(err.Error(), "database is not initialized") {
		t.Fatalf("err = %v", err)
	}
}

func TestSetBalanceIsUnknownCommand(t *testing.T) {
	cfg := config.GetDefaultConfig()
	if _, err := parser.ParseAndValidateCmdLine([]string{"set-balance"}, cfg); err == nil {
		t.Fatal("removed command still parses")
	}
	err := Dispatch(parser.ParsedCmdLine{Command: "set-balance"}, cfg, nil)
	if err == nil || err.Error() != "unknown command" {
		t.Fatalf("err = %v", err)
	}
	if _, ok := Handlers["set-balance"]; ok {
		t.Fatal("removed command still has a handler")
	}
}
