package cmd

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
	"github.com/nschaetti/cashwarrior/internal/utils"
	_ "modernc.org/sqlite"
)

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
				Subcommand: "default",
				Args: []parser.Token{{
					Kind:  parser.TokenAttribute,
					Key:   "database",
					Value: newDBPath,
					Raw:   "database:" + newDBPath,
				}},
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
				Subcommand: "default",
				Args: []parser.Token{{
					Kind:  parser.TokenAttribute,
					Key:   "database",
					Value: newDBPath,
					Raw:   "database:" + newDBPath,
				}},
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
			Subcommand: "default",
			Args: []parser.Token{{
				Kind:  parser.TokenAttribute,
				Key:   "backup.period",
				Value: "2weeks",
				Raw:   "backup.period:2weeks",
			}},
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

func TestConfigBackupKeepRejectsNegative(t *testing.T) {
	tempHome := t.TempDir()
	withHome(t, tempHome, func() {
		cfg, cashDB := openTestDB(t)
		defer cashDB.Close()

		writeTestConfig(t, cfg)

		err := Config(parser.ParsedCmdLine{
			Command:    "config",
			Subcommand: "default",
			Args: []parser.Token{{
				Kind:  parser.TokenAttribute,
				Key:   "backup.keep",
				Value: "-1",
				Raw:   "backup.keep:-1",
			}},
		}, cfg, cashDB)
		if err == nil || err.Error() != "backup.keep must be >= 0" {
			t.Fatalf("err = %v, want backup.keep must be >= 0", err)
		}
	})
}
