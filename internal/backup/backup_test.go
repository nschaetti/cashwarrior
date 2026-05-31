package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
)

func writeDBFile(t *testing.T, path string, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
}

func TestValidateConfig(t *testing.T) {
	if err := ValidateConfig(config.BackupConfig{Period: "2weeks", Keep: 2}); err != nil {
		t.Fatalf("ValidateConfig returned error: %v", err)
	}
	if err := ValidateConfig(config.BackupConfig{Period: "bad", Keep: 2}); err == nil {
		t.Fatal("ValidateConfig expected error, got nil")
	}
	if err := ValidateConfig(config.BackupConfig{Period: "day", Keep: -1}); err == nil {
		t.Fatal("ValidateConfig negative keep expected error, got nil")
	}
}

func TestRunCreatesBackupWhenDue(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "cash.db")
	writeDBFile(t, dbPath, "db")

	now := time.Date(2026, time.May, 31, 10, 15, 0, 0, time.UTC)
	if err := Run(dbPath, config.BackupConfig{Period: "day", Keep: 2}, now); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("ReadDir returned error: %v", err)
	}
	count := 0
	for _, entry := range entries {
		if entry.Name() != "cash.db" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("backup count = %d, want 1", count)
	}
}

func TestRunSkipsBackupWhenNotDue(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "cash.db")
	writeDBFile(t, dbPath, "db")

	first := time.Date(2026, time.May, 31, 10, 15, 0, 0, time.UTC)
	if err := Run(dbPath, config.BackupConfig{Period: "day", Keep: 2}, first); err != nil {
		t.Fatalf("first Run returned error: %v", err)
	}
	second := first.Add(12 * time.Hour)
	if err := Run(dbPath, config.BackupConfig{Period: "day", Keep: 2}, second); err != nil {
		t.Fatalf("second Run returned error: %v", err)
	}

	backups, err := listBackups(dbPath)
	if err != nil {
		t.Fatalf("listBackups returned error: %v", err)
	}
	if len(backups) != 1 {
		t.Fatalf("len(backups) = %d, want 1", len(backups))
	}
}

func TestRunRotatesOldBackups(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "cash.db")
	writeDBFile(t, dbPath, "db")

	base := time.Date(2026, time.May, 31, 10, 15, 0, 0, time.UTC)
	for _, when := range []time.Time{base, base.AddDate(0, 0, 2), base.AddDate(0, 0, 4)} {
		if err := Run(dbPath, config.BackupConfig{Period: "day", Keep: 2}, when); err != nil {
			t.Fatalf("Run(%v) returned error: %v", when, err)
		}
	}

	backups, err := listBackups(dbPath)
	if err != nil {
		t.Fatalf("listBackups returned error: %v", err)
	}
	if len(backups) != 2 {
		t.Fatalf("len(backups) = %d, want 2", len(backups))
	}
	for _, backup := range backups {
		if filepath.Base(backup.Path) == "cash.db.backup-20260531-101500" {
			t.Fatal("oldest backup should have been removed")
		}
	}
}
