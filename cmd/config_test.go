package cmd

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
	"github.com/nschaetti/cashwarrior/internal/utils"
	_ "modernc.org/sqlite"
)

func TestConfigPrintAndGet(t *testing.T) {
	withHome(t, t.TempDir(), func() {
		cfg, cashDB := openTestDB(t)
		defer cashDB.Close()
		cfg.Display.ShowCurrency = true
		writeTestConfig(t, cfg)
		text := captureStdout(t, func() {
			if err := Config(parser.ParsedCmdLine{Subcommand: "print"}, cfg, cashDB); err != nil {
				t.Fatal(err)
			}
		})
		for _, want := range []string{"default.currency = USD", "default.account = main", "gui.show_currency = true", "database = "} {
			if !strings.Contains(text, want) {
				t.Fatalf("missing %q: %s", want, text)
			}
		}
		text = captureStdout(t, func() {
			if err := Config(parser.ParsedCmdLine{Subcommand: "get", Args: []parser.Arg{testArg(t, "default.currency")}}, cfg, cashDB); err != nil {
				t.Fatal(err)
			}
		})
		if !strings.Contains(text, "default.currency = USD") {
			t.Fatalf("output = %s", text)
		}
		if err := Config(parser.ParsedCmdLine{Subcommand: "get", Args: []parser.Arg{testArg(t, "unknown")}}, cfg, cashDB); err == nil || err.Error() != "unknown config key: unknown" {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestConfigSetAndLegacyPersist(t *testing.T) {
	withHome(t, t.TempDir(), func() {
		cfg, cashDB := openTestDB(t)
		defer cashDB.Close()
		path := writeTestConfig(t, cfg)
		parsed, parseErr := parser.ParseAndValidateCmdLine([]string{"config", "set", "default.currency", "CHF"}, cfg)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		if err := Config(parsed, cfg, cashDB); err != nil {
			t.Fatal(err)
		}
		saved, err := config.LoadConfig(path)
		if err != nil {
			t.Fatal(err)
		}
		if saved.Default.Currency != "CHF" {
			t.Fatalf("currency = %s", saved.Default.Currency)
		}
		parsed = parser.ParsedCmdLine{Subcommand: "set", Args: []parser.Arg{testArg(t, "gui.show_currency"), testArg(t, "false")}, Flags: []parser.Arg{testArg(t, "--json"), testArg(t, "--yes")}}
		text := captureStdout(t, func() {
			if err := Config(parsed, cfg, cashDB); err != nil {
				t.Fatal(err)
			}
		})
		if !json.Valid([]byte(text)) {
			t.Fatalf("invalid JSON: %s", text)
		}
		saved, err = config.LoadConfig(path)
		if err != nil {
			t.Fatal(err)
		}
		if saved.Display.ShowCurrency {
			t.Fatal("show_currency was not persisted")
		}
		if err := Config(parser.ParsedCmdLine{Subcommand: "print", Args: []parser.Arg{testStringAttribute("backup.keep:3", "backup.keep", "3")}}, cfg, cashDB); err != nil {
			t.Fatal(err)
		}
		saved, err = config.LoadConfig(path)
		if err != nil {
			t.Fatal(err)
		}
		if saved.Backup.Keep != 3 {
			t.Fatalf("keep = %d", saved.Backup.Keep)
		}
	})
}

func TestConfigJSONSetGuards(t *testing.T) {
	withHome(t, t.TempDir(), func() {
		cfg, cashDB := openTestDB(t)
		defer cashDB.Close()
		path := writeTestConfig(t, cfg)
		parsed := parser.ParsedCmdLine{Subcommand: "set", Args: []parser.Arg{testArg(t, "default.currency"), testArg(t, "CHF")}, Flags: []parser.Arg{testArg(t, "--json")}}
		if err := Config(parsed, cfg, cashDB); err == nil || !strings.Contains(err.Error(), "--yes is required") {
			t.Fatalf("err = %v", err)
		}
		parsed.Flags = append(parsed.Flags, testArg(t, "--yes"))
		for _, value := range []string{filepath.Join(t.TempDir(), "missing.db"), t.TempDir()} {
			parsed.Args = []parser.Arg{testArg(t, "database"), testArg(t, value)}
			if err := Config(parsed, cfg, cashDB); err == nil || !strings.Contains(err.Error(), "database file does not exist") {
				t.Fatalf("err = %v", err)
			}
		}
		saved, err := config.LoadConfig(path)
		if err != nil {
			t.Fatal(err)
		}
		if saved.Database != cfg.Database || saved.Default.Currency != cfg.Default.Currency {
			t.Fatal("rejected mutation changed config")
		}
	})
}

func withHome(t *testing.T, home string, fn func()) {
	t.Helper()

	oldHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", home); err != nil {
		t.Fatalf("Setenv returned error: %v", err)
	}
	defer func() {
		if err := os.Setenv("HOME", oldHome); err != nil {
			t.Fatalf("restoring HOME returned error: %v", err)
		}
	}()

	fn()
}

func writeTestConfig(t *testing.T, cfg config.Config) string {
	t.Helper()

	configPath := utils.ExpandPath(config.DefaultConfigFile)
	if err := config.SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("SaveConfig returned error: %v", err)
	}
	return configPath
}

func TestConfigDatabaseCreatesAndInitializesMissingDB(t *testing.T) {
	tempHome := t.TempDir()
	withHome(t, tempHome, func() {
		cfg, cashDB := openTestDB(t)
		defer cashDB.Close()

		configPath := writeTestConfig(t, cfg)
		newDBPath := filepath.Join(tempHome, "data", "cash.db")

		withInput(t, "y\n", func() {
			err := Config(parser.ParsedCmdLine{
				Command:    "config",
<<<<<<< HEAD
				Subcommand: "set",
=======
				Subcommand: "default",
>>>>>>> 0f5b2e4b00ad8bb38f235429b4bb9db6bd8b606d
				Args:       []parser.Arg{testStringAttribute("database:"+newDBPath, "database", newDBPath)},
			}, cfg, cashDB)
			if err != nil {
				t.Fatalf("Config returned error: %v", err)
			}
		})

		if _, err := os.Stat(newDBPath); err != nil {
			t.Fatalf("Stat(newDBPath) returned error: %v", err)
		}

		savedCfg, err := config.LoadConfig(configPath)
		if err != nil {
			t.Fatalf("LoadConfig returned error: %v", err)
		}
		if savedCfg.Database != newDBPath {
			t.Fatalf("savedCfg.Database = %q, want %q", savedCfg.Database, newDBPath)
		}

		opened, err := sql.Open("sqlite", newDBPath)
		if err != nil {
			t.Fatalf("sql.Open returned error: %v", err)
		}
		defer opened.Close()
		if err := db.Init(opened, savedCfg); err != nil {
			t.Fatalf("Init returned error: %v", err)
		}
		if _, err := db.GetAccountByName(opened, savedCfg.Default.Account); err != nil {
			t.Fatalf("GetAccountByName returned error: %v", err)
		}
	})
}

func TestConfigDatabaseKeepsConfigWhenCreationDeclined(t *testing.T) {
	tempHome := t.TempDir()
	withHome(t, tempHome, func() {
		cfg, cashDB := openTestDB(t)
		defer cashDB.Close()

		configPath := writeTestConfig(t, cfg)
		newDBPath := filepath.Join(tempHome, "data", "cash.db")

		withInput(t, "n\n", func() {
			err := Config(parser.ParsedCmdLine{
				Command:    "config",
<<<<<<< HEAD
				Subcommand: "set",
=======
				Subcommand: "default",
>>>>>>> 0f5b2e4b00ad8bb38f235429b4bb9db6bd8b606d
				Args:       []parser.Arg{testStringAttribute("database:"+newDBPath, "database", newDBPath)},
			}, cfg, cashDB)
			if err != nil {
				t.Fatalf("Config returned error: %v", err)
			}
		})

		if _, err := os.Stat(newDBPath); !os.IsNotExist(err) {
			t.Fatalf("Stat(newDBPath) err = %v, want not exist", err)
		}

		savedCfg, err := config.LoadConfig(configPath)
		if err != nil {
			t.Fatalf("LoadConfig returned error: %v", err)
		}
		if savedCfg.Database != cfg.Database {
			t.Fatalf("savedCfg.Database = %q, want %q", savedCfg.Database, cfg.Database)
		}
	})
}

func TestConfigBackupPeriodUpdatesConfig(t *testing.T) {
	tempHome := t.TempDir()
	withHome(t, tempHome, func() {
		cfg, cashDB := openTestDB(t)
		defer cashDB.Close()

		configPath := writeTestConfig(t, cfg)

		err := Config(parser.ParsedCmdLine{
			Command:    "config",
<<<<<<< HEAD
			Subcommand: "set",
=======
			Subcommand: "default",
>>>>>>> 0f5b2e4b00ad8bb38f235429b4bb9db6bd8b606d
			Args:       []parser.Arg{testStringAttribute("backup.period:2weeks", "backup.period", "2weeks")},
		}, cfg, cashDB)
		if err != nil {
			t.Fatalf("Config returned error: %v", err)
		}

		savedCfg, err := config.LoadConfig(configPath)
		if err != nil {
			t.Fatalf("LoadConfig returned error: %v", err)
		}
		if savedCfg.Backup.Period != "2weeks" {
			t.Fatalf("savedCfg.Backup.Period = %q, want 2weeks", savedCfg.Backup.Period)
		}
	})
}

func TestConfigSetTextArgsAndGetValue(t *testing.T) {
	tempHome := t.TempDir()
	withHome(t, tempHome, func() {
		cfg, cashDB := openTestDB(t)
		defer cashDB.Close()
		configPath := writeTestConfig(t, cfg)

		err := Config(parser.ParsedCmdLine{
			Command:    "config",
			Subcommand: "set",
			Args:       []parser.Arg{parser.ArgText{Raw: "gui.date_format", Text: "gui.date_format"}, parser.ArgText{Raw: "2006-01-02", Text: "2006-01-02"}},
		}, cfg, cashDB)
		if err != nil {
			t.Fatalf("Config(set) returned error: %v", err)
		}

		savedCfg, err := config.LoadConfig(configPath)
		if err != nil {
			t.Fatalf("LoadConfig returned error: %v", err)
		}
		value, err := configValue(savedCfg, "gui.date_format")
		if err != nil {
			t.Fatalf("configValue returned error: %v", err)
		}
		if value != "2006-01-02" {
			t.Fatalf("configValue = %q, want 2006-01-02", value)
		}
	})
}

func TestConfigBackupKeepRejectsNegative(t *testing.T) {
	tempHome := t.TempDir()
	withHome(t, tempHome, func() {
		cfg, cashDB := openTestDB(t)
		defer cashDB.Close()

		writeTestConfig(t, cfg)

		err := Config(parser.ParsedCmdLine{
			Command:    "config",
<<<<<<< HEAD
			Subcommand: "set",
=======
			Subcommand: "default",
>>>>>>> 0f5b2e4b00ad8bb38f235429b4bb9db6bd8b606d
			Args:       []parser.Arg{testStringAttribute("backup.keep:-1", "backup.keep", "-1")},
		}, cfg, cashDB)
		if err == nil || err.Error() != "backup.keep must be >= 0" {
			t.Fatalf("err = %v, want backup.keep must be >= 0", err)
		}
	})
}
