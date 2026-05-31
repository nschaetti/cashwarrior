package cmd

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	backuputil "github.com/nschaetti/cashwarrior/internal/backup"
	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
	_ "modernc.org/sqlite"
)

func openFileTestDB(t *testing.T) (config.Config, *sql.DB, *sql.Tx, string) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "cash.db")
	cfg := config.GetDefaultConfig()
	cfg.Database = dbPath
	cfg.Backup.Keep = 2

	database, err := db.Open(cfg)
	if err != nil {
		t.Fatalf("db.Open returned error: %v", err)
	}
	tx, err := database.Begin()
	if err != nil {
		_ = database.Close()
		t.Fatalf("Begin returned error: %v", err)
	}
	return cfg, database, tx, dbPath
}

func TestBackupCommandCreatesRotatedBackup(t *testing.T) {
	cfg, database, tx, dbPath := openFileTestDB(t)
	defer database.Close()
	defer tx.Rollback()

	output := captureStdout(t, func() {
		if err := Backup(parser.ParsedCmdLine{Command: "backup", Subcommand: "default"}, cfg, tx); err != nil {
			t.Fatalf("Backup returned error: %v", err)
		}
	})

	entries, err := os.ReadDir(filepath.Dir(dbPath))
	if err != nil {
		t.Fatalf("ReadDir returned error: %v", err)
	}
	backupCount := 0
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "cash.db.backup-") {
			backupCount++
		}
	}
	if backupCount != 1 {
		t.Fatalf("backupCount = %d, want 1", backupCount)
	}
	if !strings.Contains(output, "Database backed up to") {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestBackupCommandCopiesToExplicitOutput(t *testing.T) {
	cfg, database, tx, dbPath := openFileTestDB(t)
	defer database.Close()
	defer tx.Rollback()

	outputPath := filepath.Join(filepath.Dir(dbPath), "custom", "backup.db")
	if err := Backup(parser.ParsedCmdLine{Command: "backup", Subcommand: "default", Args: []parser.Token{{Raw: "output:" + outputPath, Kind: parser.TokenAttribute, Key: "output", Value: outputPath}}}, cfg, tx); err != nil {
		t.Fatalf("Backup returned error: %v", err)
	}

	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("Stat(outputPath) returned error: %v", err)
	}

	entries, err := os.ReadDir(filepath.Dir(dbPath))
	if err != nil {
		t.Fatalf("ReadDir returned error: %v", err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "cash.db.backup-") {
			t.Fatalf("unexpected rotated backup created: %s", entry.Name())
		}
	}
}

func TestBackupRunNowRotatesOldest(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "cash.db")
	if err := os.WriteFile(dbPath, []byte("db"), 0644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	for _, when := range []time.Time{
		time.Date(2026, time.May, 31, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.May, 31, 11, 0, 0, 0, time.UTC),
		time.Date(2026, time.May, 31, 12, 0, 0, 0, time.UTC),
	} {
		if _, err := backuputil.RunNow(dbPath, 2, when); err != nil {
			t.Fatalf("RunNow returned error: %v", err)
		}
	}

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("ReadDir returned error: %v", err)
	}
	backupNames := make([]string, 0)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "cash.db.backup-") {
			backupNames = append(backupNames, entry.Name())
		}
	}
	if len(backupNames) != 2 {
		t.Fatalf("len(backupNames) = %d, want 2", len(backupNames))
	}
	for _, name := range backupNames {
		if name == "cash.db.backup-20260531-100000" {
			t.Fatalf("oldest backup was not rotated: %v", backupNames)
		}
	}
}
