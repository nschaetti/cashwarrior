package cmd

import (
	"database/sql"
	"encoding/json"
	"errors"
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

	parsed.Filters = []parser.Arg{testArg(t, "order:start_date")}
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

	parsed.Filters = []parser.Arg{
		testArg(t, "order:end_date"),
		testArg(t, "desc:true"),
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

	parsed.Filters = []parser.Arg{testArg(t, "order:amount")}
	err := Groups(parsed, cfg, cashDB)
	if err == nil || err.Error() != "unsupported groups order field amount" {
		t.Fatalf("err = %v, want unsupported groups order field amount", err)
	}
}

func groupsParsed(t *testing.T, subcommand string, args ...string) parser.ParsedCmdLine {
	t.Helper()
	parsed := parser.ParsedCmdLine{Command: "groups", Subcommand: subcommand}
	for _, raw := range args {
		parsed.Args = append(parsed.Args, testArg(t, raw))
	}
	return parsed
}

func TestGroupsAddTextFormLinksToExistingGroup(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed := groupsParsed(t, "add", "alpha", "identifier:2026.05.3")
	parsed.Flags = yesFlag()

	output := captureStdout(t, func() {
		if err := Groups(parsed, cfg, cashDB); err != nil {
			t.Fatalf("Groups returned error: %v", err)
		}
	})

	if !strings.Contains(output, "Linked 1 transactions to group alpha") {
		t.Fatalf("output missing success message: %s", output)
	}
	transaction, err := db.GetTransactionByIdentifier(cashDB, "2026.05.3")
	if err != nil {
		t.Fatalf("GetTransactionByIdentifier returned error: %v", err)
	}
	alpha, err := db.GetGroupByName(cashDB, "alpha")
	if err != nil {
		t.Fatalf("GetGroupByName(alpha) returned error: %v", err)
	}
	if transaction.GroupID == nil || *transaction.GroupID != alpha.ID {
		t.Fatalf("transaction.GroupID = %v, want %d", transaction.GroupID, alpha.ID)
	}
}

func TestGroupsAddAttributeFormLinks(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed := groupsParsed(t, "add", "group:zeta", "identifier:2026.05.1")
	parsed.Flags = yesFlag()

	if err := Groups(parsed, cfg, cashDB); err != nil {
		t.Fatalf("Groups returned error: %v", err)
	}
	transaction, err := db.GetTransactionByIdentifier(cashDB, "2026.05.1")
	if err != nil {
		t.Fatalf("GetTransactionByIdentifier returned error: %v", err)
	}
	zeta, err := db.GetGroupByName(cashDB, "zeta")
	if err != nil {
		t.Fatalf("GetGroupByName(zeta) returned error: %v", err)
	}
	if transaction.GroupID == nil || *transaction.GroupID != zeta.ID {
		t.Fatalf("transaction.GroupID = %v, want %d", transaction.GroupID, zeta.ID)
	}
}

func TestGroupsAddCreatesGroup(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed := groupsParsed(t, "add", "newgrp", "identifier:2026.05.3")
	parsed.Flags = yesFlag()

	output := captureStdout(t, func() {
		if err := Groups(parsed, cfg, cashDB); err != nil {
			t.Fatalf("Groups returned error: %v", err)
		}
	})

	if !strings.Contains(output, "Linked 1 transactions to group newgrp") {
		t.Fatalf("output missing success message: %s", output)
	}
	if _, err := db.GetGroupByName(cashDB, "newgrp"); err != nil {
		t.Fatalf("GetGroupByName(newgrp) returned error: %v", err)
	}
}

func TestGroupsAddJSONRequiresYes(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed := groupsParsed(t, "add", "alpha", "identifier:2026.05.3")
	parsed.Flags = []parser.Arg{testArg(t, "--json")}

	err := Groups(parsed, cfg, cashDB)
	if err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("err = %v, want --yes required error", err)
	}
}

func TestGroupsAddJSON(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed := groupsParsed(t, "add", "alpha", "identifier:2026.05.3")
	parsed.Flags = append(yesFlag(), testArg(t, "--json"))

	var envelope struct {
		Data map[string]any `json:"data"`
	}
	got := captureListOutput(t, func() error {
		return Groups(parsed, cfg, cashDB)
	})
	if err := json.Unmarshal([]byte(got), &envelope); err != nil {
		t.Fatalf("JSON output is invalid: %v", err)
	}
	if envelope.Data["group"] != "alpha" {
		t.Fatalf("JSON group = %v, want alpha", envelope.Data["group"])
	}
}

func TestGroupsAddTransactionNotFound(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed := groupsParsed(t, "add", "alpha", "identifier:2026.04.1")
	parsed.Flags = yesFlag()

	err := Groups(parsed, cfg, cashDB)
	if err == nil || !strings.Contains(err.Error(), "transaction 2026.04.1 does not exist") {
		t.Fatalf("err = %v, want transaction does not exist error", err)
	}
}

func TestGroupsModifyRenamesGroup(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed := groupsParsed(t, "modify", "alpha", "group:beta")

	output := captureStdout(t, func() {
		if err := Groups(parsed, cfg, cashDB); err != nil {
			t.Fatalf("Groups returned error: %v", err)
		}
	})

	if !strings.Contains(output, "Group alpha updated") {
		t.Fatalf("output missing success message: %s", output)
	}
	if _, err := db.GetGroupByName(cashDB, "beta"); err != nil {
		t.Fatalf("GetGroupByName(beta) returned error: %v", err)
	}
	if _, err := db.GetGroupByName(cashDB, "alpha"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("GetGroupByName(alpha) err = %v, want sql.ErrNoRows", err)
	}
}

func TestGroupsModifyNotFound(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed := groupsParsed(t, "modify", "nosuch", "group:beta")
	err := Groups(parsed, cfg, cashDB)
	if err == nil || !strings.Contains(err.Error(), "group nosuch does not exist") {
		t.Fatalf("err = %v, want group does not exist error", err)
	}
}

func TestGroupsModifyConflict(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed := groupsParsed(t, "modify", "alpha", "group:zeta")
	err := Groups(parsed, cfg, cashDB)
	if err == nil || !strings.Contains(err.Error(), "group zeta already exists") {
		t.Fatalf("err = %v, want group already exists error", err)
	}
}

func TestGroupsModifyJSON(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed := groupsParsed(t, "modify", "alpha", "group:beta")
	parsed.Flags = []parser.Arg{testArg(t, "--json")}

	got := captureListOutput(t, func() error {
		return Groups(parsed, cfg, cashDB)
	})
	if !strings.Contains(got, `"action":"updated"`) || !strings.Contains(got, `"name":"alpha"`) {
		t.Fatalf("JSON output missing updated info: %s", got)
	}
}

func TestGroupsDeleteDetachesLinkedTransactions(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed := groupsParsed(t, "delete", "alpha")
	parsed.Flags = yesFlag()

	output := captureStdout(t, func() {
		if err := Groups(parsed, cfg, cashDB); err != nil {
			t.Fatalf("Groups returned error: %v", err)
		}
	})

	if !strings.Contains(output, "Warning: group alpha has 2 linked transactions, they will be detached") {
		t.Fatalf("output missing detach warning: %s", output)
	}
	if !strings.Contains(output, "Group alpha deleted") {
		t.Fatalf("output missing success message: %s", output)
	}
	if _, err := db.GetGroupByName(cashDB, "alpha"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("GetGroupByName(alpha) err = %v, want sql.ErrNoRows", err)
	}
	for _, identifier := range []string{"2026.05.1", "2026.05.2"} {
		transaction, err := db.GetTransactionByIdentifier(cashDB, identifier)
		if err != nil {
			t.Fatalf("GetTransactionByIdentifier(%s) returned error: %v", identifier, err)
		}
		if transaction.GroupID != nil {
			t.Fatalf("transaction %s.GroupID = %v, want nil", identifier, transaction.GroupID)
		}
	}
}

func TestGroupsDeleteEmptyGroup(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	if _, err := db.InsertTransactionGroup(cashDB, db.CreateTransactionGroupInput{Name: "empty"}); err != nil {
		t.Fatalf("InsertTransactionGroup(empty) returned error: %v", err)
	}

	parsed := groupsParsed(t, "delete", "empty")
	parsed.Flags = yesFlag()

	output := captureStdout(t, func() {
		if err := Groups(parsed, cfg, cashDB); err != nil {
			t.Fatalf("Groups returned error: %v", err)
		}
	})

	if strings.Contains(output, "linked transactions") {
		t.Fatalf("output should not contain detach warning: %s", output)
	}
	if !strings.Contains(output, "Group empty deleted") {
		t.Fatalf("output missing success message: %s", output)
	}
}

func TestGroupsDeleteNotFound(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed := groupsParsed(t, "delete", "nosuch")
	parsed.Flags = yesFlag()

	err := Groups(parsed, cfg, cashDB)
	if err == nil || !strings.Contains(err.Error(), "group nosuch does not exist") {
		t.Fatalf("err = %v, want group does not exist error", err)
	}
}

func TestGroupsDeleteJSONRequiresYes(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed := groupsParsed(t, "delete", "alpha")
	parsed.Flags = []parser.Arg{testArg(t, "--json")}

	err := Groups(parsed, cfg, cashDB)
	if err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("err = %v, want --yes required error", err)
	}
}

func TestGroupsDeleteJSON(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed := groupsParsed(t, "delete", "alpha")
	parsed.Flags = append(yesFlag(), testArg(t, "--json"))

	got := captureListOutput(t, func() error {
		return Groups(parsed, cfg, cashDB)
	})
	if !strings.Contains(got, `"action":"deleted"`) || !strings.Contains(got, `"name":"alpha"`) {
		t.Fatalf("JSON output missing deleted info: %s", got)
	}
}

func TestGroupsRemoveDetachesTransaction(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed := groupsParsed(t, "remove", "identifier:2026.05.1", "group:alpha")
	parsed.Flags = yesFlag()

	output := captureStdout(t, func() {
		if err := Groups(parsed, cfg, cashDB); err != nil {
			t.Fatalf("Groups returned error: %v", err)
		}
	})

	if !strings.Contains(output, "Removed transaction 2026.05.1 from group alpha") {
		t.Fatalf("output missing success message: %s", output)
	}
	transaction, err := db.GetTransactionByIdentifier(cashDB, "2026.05.1")
	if err != nil {
		t.Fatalf("GetTransactionByIdentifier returned error: %v", err)
	}
	if transaction.GroupID != nil {
		t.Fatalf("transaction.GroupID = %v, want nil", transaction.GroupID)
	}
}

func TestGroupsRemoveNotInGroup(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed := groupsParsed(t, "remove", "identifier:2026.05.1", "group:zeta")
	parsed.Flags = yesFlag()

	err := Groups(parsed, cfg, cashDB)
	if err == nil || !strings.Contains(err.Error(), "transaction 2026.05.1 is not in group zeta") {
		t.Fatalf("err = %v, want not in group error", err)
	}
}

func TestGroupsRemoveGroupNotFound(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed := groupsParsed(t, "remove", "identifier:2026.05.1", "group:nosuch")
	parsed.Flags = yesFlag()

	err := Groups(parsed, cfg, cashDB)
	if err == nil || !strings.Contains(err.Error(), "group nosuch does not exist") {
		t.Fatalf("err = %v, want group does not exist error", err)
	}
}

func TestGroupsRemoveTransactionNotFound(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed := groupsParsed(t, "remove", "identifier:2026.04.1", "group:alpha")
	parsed.Flags = yesFlag()

	err := Groups(parsed, cfg, cashDB)
	if err == nil || !strings.Contains(err.Error(), "transaction 2026.04.1 does not exist") {
		t.Fatalf("err = %v, want transaction does not exist error", err)
	}
}

func TestGroupsRemoveJSONRequiresYes(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed := groupsParsed(t, "remove", "identifier:2026.05.1", "group:alpha")
	parsed.Flags = []parser.Arg{testArg(t, "--json")}

	err := Groups(parsed, cfg, cashDB)
	if err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("err = %v, want --yes required error", err)
	}
}

func TestGroupsRemoveJSON(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed := groupsParsed(t, "remove", "identifier:2026.05.1", "group:alpha")
	parsed.Flags = append(yesFlag(), testArg(t, "--json"))

	got := captureListOutput(t, func() error {
		return Groups(parsed, cfg, cashDB)
	})
	if !strings.Contains(got, `"action":"removed"`) || !strings.Contains(got, `"identifier":"2026.05.1"`) {
		t.Fatalf("JSON output missing removed info: %s", got)
	}
}

func TestGroupsUnknownSubcommand(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	parsed := groupsParsed(t, "frobnicate")
	err := Groups(parsed, cfg, cashDB)
	if err == nil || err.Error() != "unknown groups subcommand frobnicate" {
		t.Fatalf("err = %v, want unknown groups subcommand frobnicate", err)
	}
}

func TestGroupsAliasDispatch(t *testing.T) {
	cfg, cashDB, _ := setupGroupsTestData(t)
	defer cashDB.Close()

	if _, err := db.InsertTransactionGroup(cashDB, db.CreateTransactionGroupInput{Name: "empty"}); err != nil {
		t.Fatalf("InsertTransactionGroup(empty) returned error: %v", err)
	}

	ls := groupsParsed(t, "ls")
	if err := Groups(ls, cfg, cashDB); err != nil {
		t.Fatalf("Groups(ls) returned error: %v", err)
	}

	rn := groupsParsed(t, "rn", "alpha", "group:beta")
	if err := Groups(rn, cfg, cashDB); err != nil {
		t.Fatalf("Groups(rn) returned error: %v", err)
	}
	if _, err := db.GetGroupByName(cashDB, "beta"); err != nil {
		t.Fatalf("GetGroupByName(beta) returned error: %v", err)
	}

	rm := groupsParsed(t, "rm", "empty")
	rm.Flags = yesFlag()
	if err := Groups(rm, cfg, cashDB); err != nil {
		t.Fatalf("Groups(rm) returned error: %v", err)
	}
	if _, err := db.GetGroupByName(cashDB, "empty"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("GetGroupByName(empty) err = %v, want sql.ErrNoRows", err)
	}
}
