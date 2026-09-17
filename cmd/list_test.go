package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestParseListSortOptions(t *testing.T) {
	parsed := parser.ParsedCmdLine{
		Command: "list",
		Filters: []parser.Arg{
			testArg(t, "order"+":"+"datetime"),
			testArg(t, "desc"+":"+"false"),
			testArg(t, "date"+":"+"month"),
		},
	}

	filtered, options, err := parseListSortOptions(parsed)
	if err != nil {
		t.Fatalf("parseListSortOptions returned error: %v", err)
	}
	if options.Field != "datetime" {
		t.Fatalf("Field = %q, want %q", options.Field, "datetime")
	}
	if options.Desc {
		t.Fatal("Desc = true, want false")
	}
	if len(filtered.Filters) != 1 {
		t.Fatalf("len(filtered.Filters) = %d, want 1", len(filtered.Filters))
	}
	if filtered.Filters[0].(parser.ArgAttribute).Key != "date" {
		t.Fatalf("remaining filter key = %q, want %q", filtered.Filters[0].(parser.ArgAttribute).Key, "date")
	}
}

func TestParseListSortOptionsDefaultsToDesc(t *testing.T) {
	parsed := parser.ParsedCmdLine{
		Command: "list",
		Filters: []parser.Arg{
			testArg(t, "order"+":"+"description"),
		},
	}

	_, options, err := parseListSortOptions(parsed)
	if err != nil {
		t.Fatalf("parseListSortOptions returned error: %v", err)
	}
	if !options.Desc {
		t.Fatal("Desc = false, want true")
	}
}

func TestParseListSortOptionsRejectsUnsupportedField(t *testing.T) {
	parsed := parser.ParsedCmdLine{
		Command: "list",
		Filters: []parser.Arg{
			testArg(t, "order"+":"+"unknown"),
		},
	}

	_, _, err := parseListSortOptions(parsed)
	if err == nil {
		t.Fatal("parseListSortOptions expected error, got nil")
	}
}

func TestParseListSortOptionsAcceptsDateAlias(t *testing.T) {
	parsed := parser.ParsedCmdLine{
		Command: "list",
		Filters: []parser.Arg{testArg(t, "order"+":"+"date")},
	}

	_, options, err := parseListSortOptions(parsed)
	if err != nil {
		t.Fatalf("parseListSortOptions returned error: %v", err)
	}
	if options.Field != "date" {
		t.Fatalf("Field = %q, want %q", options.Field, "date")
	}
}

func TestParseListSortOptionsConsumesArgsAttributes(t *testing.T) {
	parsed := parser.ParsedCmdLine{
		Command: "list",
		Args: []parser.Arg{
			testArg(t, "order"+":"+"date"),
			testArg(t, "desc"+":"+"false"),
			testArg(t, "account"+":"+"main"),
		},
	}

	filtered, options, err := parseListSortOptions(parsed)
	if err != nil {
		t.Fatalf("parseListSortOptions returned error: %v", err)
	}
	if options.Field != "date" {
		t.Fatalf("Field = %q, want %q", options.Field, "date")
	}
	if options.Desc {
		t.Fatal("Desc = true, want false")
	}
	if len(filtered.Args) != 1 || filtered.Args[0].(parser.ArgAttribute).Key != "account" {
		t.Fatalf("filtered.Args = %#v, want only account attr", filtered.Args)
	}
}

func TestClassifyFilterGroup(t *testing.T) {
	arg := testArg(t, "group:ticket_0001")
	if got := classifyFilter(arg); got != FilterTypeGroup {
		t.Fatalf("classifyFilter(group) = %d, want %d", got, FilterTypeGroup)
	}
}

func TestListAccountsSubcommand(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	if _, err := db.InsertAccount(cashDB, db.CreateAccountInput{Name: "savings", Currency: "CHF"}); err != nil {
		t.Fatalf("InsertAccount returned error: %v", err)
	}

	output := captureStdout(t, func() {
		err := List(parser.ParsedCmdLine{Command: "list", Subcommand: "accounts"}, cfg, cashDB)
		if err != nil {
			t.Fatalf("List returned error: %v", err)
		}
	})

	if !strings.Contains(output, "savings") {
		t.Fatalf("output missing savings account: %s", output)
	}
	if !strings.Contains(output, cfg.Default.Account) {
		t.Fatalf("output missing default account %q: %s", cfg.Default.Account, output)
	}
}

func TestListAccountsSubcommandAlias(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	if _, err := db.InsertAccount(cashDB, db.CreateAccountInput{Name: "savings", Currency: "CHF"}); err != nil {
		t.Fatalf("InsertAccount returned error: %v", err)
	}

	output := captureStdout(t, func() {
		err := List(parser.ParsedCmdLine{Command: "list", Subcommand: "a"}, cfg, cashDB)
		if err != nil {
			t.Fatalf("List returned error: %v", err)
		}
	})

	if !strings.Contains(output, "savings") {
		t.Fatalf("output missing savings account: %s", output)
	}
}

func TestListGroupsSubcommand(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	for _, name := range []string{"zeta", "alpha"} {
		if _, err := db.InsertTransactionGroup(cashDB, db.CreateTransactionGroupInput{Name: name}); err != nil {
			t.Fatalf("InsertTransactionGroup(%s) returned error: %v", name, err)
		}
	}

	output := captureStdout(t, func() {
		err := List(parser.ParsedCmdLine{
			Command:    "list",
			Subcommand: "groups",
			Filters:    []parser.Arg{testArg(t, "order:name")},
		}, cfg, cashDB)
		if err != nil {
			t.Fatalf("List returned error: %v", err)
		}
	})

	alphaIndex := strings.Index(output, "alpha")
	zetaIndex := strings.Index(output, "zeta")
	if alphaIndex == -1 || zetaIndex == -1 {
		t.Fatalf("output missing expected group names: %s", output)
	}
	if alphaIndex > zetaIndex {
		t.Fatalf("groups are not sorted by name ascending: %s", output)
	}
}

func TestListTagsSubcommand(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	if _, err := db.InsertTag(cashDB, db.CreateTagInput{Name: "food"}); err != nil {
		t.Fatalf("InsertTag returned error: %v", err)
	}

	output := captureStdout(t, func() {
		err := List(parser.ParsedCmdLine{Command: "list", Subcommand: "tags"}, cfg, cashDB)
		if err != nil {
			t.Fatalf("List returned error: %v", err)
		}
	})

	if !strings.Contains(output, "food") {
		t.Fatalf("output missing food tag: %s", output)
	}
}

func TestListTagsSubcommandAlias(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	if _, err := db.InsertTag(cashDB, db.CreateTagInput{Name: "food"}); err != nil {
		t.Fatalf("InsertTag returned error: %v", err)
	}

	output := captureStdout(t, func() {
		err := List(parser.ParsedCmdLine{Command: "list", Subcommand: "ta"}, cfg, cashDB)
		if err != nil {
			t.Fatalf("List returned error: %v", err)
		}
	})

	if !strings.Contains(output, "food") {
		t.Fatalf("output missing food tag: %s", output)
	}
}

func TestListTransactionsSubcommandRegression(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	mainAccount, err := db.GetAccountByName(cashDB, cfg.Default.Account)
	if err != nil {
		t.Fatalf("GetAccountByName returned error: %v", err)
	}
	placeID, err := db.InsertStore(cashDB, db.CreatePlaceInput{Name: "List Test"})
	if err != nil {
		t.Fatalf("InsertStore returned error: %v", err)
	}
	if _, err := db.InsertTransaction(cashDB, db.CreateTransactionInput{
		Identifier:  "2026.05.1",
		Amount:      -3.50,
		Description: "Coffee",
		Date:        time.Date(2026, time.May, 27, 0, 0, 0, 0, time.UTC),
		AccountID:   mainAccount.ID,
		PlaceID:     &placeID,
	}); err != nil {
		t.Fatalf("InsertTransaction returned error: %v", err)
	}

	for _, subcommand := range []string{"", "transactions", "t"} {
		output := captureStdout(t, func() {
			err := List(parser.ParsedCmdLine{Command: "list", Subcommand: subcommand}, cfg, cashDB)
			if err != nil {
				t.Fatalf("List(%q) returned error: %v", subcommand, err)
			}
		})

		if !strings.Contains(output, "Coffee") || !strings.Contains(output, "2026.05.1") {
			t.Fatalf("List(%q) output missing transaction: %s", subcommand, output)
		}
	}
}

func TestListUnknownSubcommand(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	err := List(parser.ParsedCmdLine{Command: "list", Subcommand: "unknown"}, cfg, cashDB)
	if err == nil || err.Error() != "unknown list subcommand: unknown" {
		t.Fatalf("err = %v, want unknown list subcommand: unknown", err)
	}
}
