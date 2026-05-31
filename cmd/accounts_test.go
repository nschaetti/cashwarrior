package cmd

import (
	"os"
	"testing"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestAddAccount(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	err := addAccount(parser.ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "add",
		Args: []parser.Token{
			{Kind: parser.TokenText, Raw: "savings"},
			{Kind: parser.TokenAttribute, Key: "currency", Value: "EUR", Raw: "currency:EUR"},
		},
	}, cfg, cashDB)
	if err != nil {
		t.Fatalf("addAccount returned error: %v", err)
	}

	account, err := db.GetAccountByName(cashDB, "savings")
	if err != nil {
		t.Fatalf("GetAccountByName returned error: %v", err)
	}
	if account.Currency != "EUR" {
		t.Fatalf("currency = %q, want EUR", account.Currency)
	}
}

func TestAddAccountWithInitialBalance(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	err := addAccount(parser.ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "add",
		Args: []parser.Token{
			{Kind: parser.TokenText, Raw: "wallet"},
			{Kind: parser.TokenAttribute, Key: "initial_balance", Value: "125.75", Raw: "initial_balance:125.75"},
		},
	}, cfg, cashDB)
	if err != nil {
		t.Fatalf("addAccount returned error: %v", err)
	}

	account, err := db.GetAccountByName(cashDB, "wallet")
	if err != nil {
		t.Fatalf("GetAccountByName returned error: %v", err)
	}
	if account.InitialBalance != 125.75 {
		t.Fatalf("initial balance = %v, want 125.75", account.InitialBalance)
	}
}

func TestModifyAccount(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	_, err := db.InsertAccount(cashDB, db.CreateAccountInput{Name: "savings", Currency: "CHF"})
	if err != nil {
		t.Fatalf("InsertAccount returned error: %v", err)
	}

	err = modifyAccount(parser.ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "modify",
		Args: []parser.Token{
			{Kind: parser.TokenText, Raw: "savings"},
			{Kind: parser.TokenAttribute, Key: "account", Value: "brokerage", Raw: "account:brokerage"},
			{Kind: parser.TokenAttribute, Key: "currency", Value: "USD", Raw: "currency:USD"},
		},
	}, cfg, cashDB)
	if err != nil {
		t.Fatalf("modifyAccount returned error: %v", err)
	}

	account, err := db.GetAccountByName(cashDB, "brokerage")
	if err != nil {
		t.Fatalf("GetAccountByName returned error: %v", err)
	}
	if account.Currency != "USD" {
		t.Fatalf("currency = %q, want USD", account.Currency)
	}
}

func TestModifyAccountInitialBalance(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	_, err := db.InsertAccount(cashDB, db.CreateAccountInput{Name: "savings", Currency: "CHF"})
	if err != nil {
		t.Fatalf("InsertAccount returned error: %v", err)
	}

	err = modifyAccount(parser.ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "modify",
		Args: []parser.Token{
			{Kind: parser.TokenText, Raw: "savings"},
			{Kind: parser.TokenAttribute, Key: "initial_balance", Value: "50.25", Raw: "initial_balance:50.25"},
		},
	}, cfg, cashDB)
	if err != nil {
		t.Fatalf("modifyAccount returned error: %v", err)
	}

	account, err := db.GetAccountByName(cashDB, "savings")
	if err != nil {
		t.Fatalf("GetAccountByName returned error: %v", err)
	}
	if account.InitialBalance != 50.25 {
		t.Fatalf("initial balance = %v, want 50.25", account.InitialBalance)
	}
}

func TestSetAccountInitialBalanceWithAccountThenAmount(t *testing.T) {
	_, cashDB := openTestDB(t)
	defer cashDB.Close()

	if err := setAccountInitialBalance(parser.ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "initial-balance",
		Args: []parser.Token{
			{Kind: parser.TokenText, Raw: "main"},
			{Kind: parser.TokenText, Raw: "200.5"},
		},
	}, cashDB); err != nil {
		t.Fatalf("setAccountInitialBalance returned error: %v", err)
	}

	account, err := db.GetAccountByName(cashDB, "main")
	if err != nil {
		t.Fatalf("GetAccountByName returned error: %v", err)
	}
	if account.InitialBalance != 200.5 {
		t.Fatalf("initial balance = %v, want 200.5", account.InitialBalance)
	}
}

func TestSetAccountInitialBalanceWithAmountThenAccount(t *testing.T) {
	_, cashDB := openTestDB(t)
	defer cashDB.Close()

	err := setAccountInitialBalance(parser.ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "initial-balance",
		Args: []parser.Token{
			{Kind: parser.TokenText, Raw: "300.75"},
			{Kind: parser.TokenText, Raw: "main"},
		},
	}, cashDB)
	if err != nil {
		t.Fatalf("setAccountInitialBalance returned error: %v", err)
	}

	account, err := db.GetAccountByName(cashDB, "main")
	if err != nil {
		t.Fatalf("GetAccountByName returned error: %v", err)
	}
	if account.InitialBalance != 300.75 {
		t.Fatalf("initial balance = %v, want 300.75", account.InitialBalance)
	}
}

func TestRenameAccount(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	_, err := db.InsertAccount(cashDB, db.CreateAccountInput{Name: "savings", Currency: "CHF"})
	if err != nil {
		t.Fatalf("InsertAccount returned error: %v", err)
	}

	err = renameAccount(parser.ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "rename",
		Args: []parser.Token{
			{Kind: parser.TokenText, Raw: "savings"},
			{Kind: parser.TokenText, Raw: "brokerage"},
		},
	}, cfg, cashDB)
	if err != nil {
		t.Fatalf("renameAccount returned error: %v", err)
	}

	_, err = db.GetAccountByName(cashDB, "savings")
	if err == nil {
		t.Fatal("GetAccountByName(savings) expected error, got nil")
	}

	_, err = db.GetAccountByName(cashDB, "brokerage")
	if err != nil {
		t.Fatalf("GetAccountByName(brokerage) returned error: %v", err)
	}
}

func TestRenameDefaultAccountUpdatesConfigAfterConfirmation(t *testing.T) {
	tempHome := t.TempDir()
	withHome(t, tempHome, func() {
		cfg, cashDB := openTestDB(t)
		defer cashDB.Close()

		if _, err := db.GetAccountByName(cashDB, cfg.Default.Account); err != nil {
			t.Fatalf("GetAccountByName returned error: %v", err)
		}
		configPath := writeTestConfig(t, cfg)

		withInput(t, "y\n", func() {
			err := renameAccount(parser.ParsedCmdLine{
				Command:    "accounts",
				Subcommand: "rename",
				Args: []parser.Token{
					{Kind: parser.TokenText, Raw: cfg.Default.Account},
					{Kind: parser.TokenText, Raw: "primary"},
				},
			}, cfg, cashDB)
			if err != nil {
				t.Fatalf("renameAccount returned error: %v", err)
			}
		})

		if _, err := db.GetAccountByName(cashDB, "primary"); err != nil {
			t.Fatalf("GetAccountByName(primary) returned error: %v", err)
		}

		savedCfg, err := config.LoadConfig(configPath)
		if err != nil {
			t.Fatalf("LoadConfig returned error: %v", err)
		}
		if savedCfg.Default.Account != "primary" {
			t.Fatalf("savedCfg.Default.Account = %q, want primary", savedCfg.Default.Account)
		}
	})
}

func TestRenameDefaultAccountCancelledKeepsName(t *testing.T) {
	tempHome := t.TempDir()
	withHome(t, tempHome, func() {
		cfg, cashDB := openTestDB(t)
		defer cashDB.Close()

		configPath := writeTestConfig(t, cfg)

		withInput(t, "n\n", func() {
			err := renameAccount(parser.ParsedCmdLine{
				Command:    "accounts",
				Subcommand: "rename",
				Args: []parser.Token{
					{Kind: parser.TokenText, Raw: cfg.Default.Account},
					{Kind: parser.TokenText, Raw: "primary"},
				},
			}, cfg, cashDB)
			if err != nil {
				t.Fatalf("renameAccount returned error: %v", err)
			}
		})

		if _, err := db.GetAccountByName(cashDB, cfg.Default.Account); err != nil {
			t.Fatalf("GetAccountByName(default) returned error: %v", err)
		}
		if _, err := db.GetAccountByName(cashDB, "primary"); err == nil {
			t.Fatal("GetAccountByName(primary) expected error, got nil")
		}

		savedCfg, err := config.LoadConfig(configPath)
		if err != nil {
			t.Fatalf("LoadConfig returned error: %v", err)
		}
		if savedCfg.Default.Account != cfg.Default.Account {
			t.Fatalf("savedCfg.Default.Account = %q, want %q", savedCfg.Default.Account, cfg.Default.Account)
		}

		if _, err := os.Stat(configPath); err != nil {
			t.Fatalf("Stat(configPath) returned error: %v", err)
		}
	})
}

func TestDeleteAccountRejectsLinkedTransactions(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	accountID, err := db.InsertAccount(cashDB, db.CreateAccountInput{Name: "savings", Currency: cfg.Default.Currency})
	if err != nil {
		t.Fatalf("InsertAccount returned error: %v", err)
	}
	placeID, err := db.InsertPlace(cashDB, db.CreatePlaceInput{Name: "Delete Account Test"})
	if err != nil {
		t.Fatalf("InsertPlace returned error: %v", err)
	}
	_, err = db.InsertTransaction(cashDB, db.CreateTransactionInput{Identifier: "2026.05.1", Amount: -5, Description: "x", Datetime: testTime(), AccountID: accountID, PlaceID: &placeID})
	if err != nil {
		t.Fatalf("InsertTransaction returned error: %v", err)
	}

	err = deleteAccount(parser.ParsedCmdLine{Command: "accounts", Subcommand: "delete", Args: []parser.Token{{Kind: parser.TokenText, Raw: "savings"}}}, cfg, cashDB)
	if err == nil {
		t.Fatal("deleteAccount expected error, got nil")
	}
}

func TestDeleteAccountDeletesEmptyAccountAfterConfirmation(t *testing.T) {
	cfg, cashDB := openTestDB(t)
	defer cashDB.Close()

	_, err := db.InsertAccount(cashDB, db.CreateAccountInput{Name: "savings", Currency: cfg.Default.Currency})
	if err != nil {
		t.Fatalf("InsertAccount returned error: %v", err)
	}

	withInput(t, "y\n", func() {
		err = deleteAccount(parser.ParsedCmdLine{Command: "accounts", Subcommand: "delete", Args: []parser.Token{{Kind: parser.TokenText, Raw: "savings"}}}, cfg, cashDB)
	})
	if err != nil {
		t.Fatalf("deleteAccount returned error: %v", err)
	}

	_, err = db.GetAccountByName(cashDB, "savings")
	if err == nil {
		t.Fatal("GetAccountByName expected error, got nil")
	}
}

func testTime() (t time.Time) { return }
