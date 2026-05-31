package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
)

type fileInfo struct {
	Path    string
	ModTime time.Time
}

func parsePeriod(period string) (int, string, error) {
	period = strings.TrimSpace(strings.ToLower(period))
	if period == "" {
		return 0, "", nil
	}

	for _, unit := range []string{"days", "day", "weeks", "week", "months", "month"} {
		if !strings.HasSuffix(period, unit) {
			continue
		}
		countPart := strings.TrimSuffix(period, unit)
		if countPart == "" {
			return 1, unit, nil
		}
		count, err := strconv.Atoi(countPart)
		if err != nil || count <= 0 {
			return 0, "", fmt.Errorf("invalid backup period: %s", period)
		}
		return count, unit, nil
	}

	return 0, "", fmt.Errorf("invalid backup period: %s", period)
}

func ValidateConfig(cfg config.BackupConfig) error {
	if cfg.Keep < 0 {
		return fmt.Errorf("backup.keep must be >= 0")
	}
	if cfg.Period == "" {
		return nil
	}
	_, _, err := parsePeriod(cfg.Period)
	return err
}

func Enabled(cfg config.BackupConfig) bool {
	return cfg.Keep > 0 && strings.TrimSpace(cfg.Period) != ""
}

func cutoffTime(now time.Time, period string) (time.Time, error) {
	count, unit, err := parsePeriod(period)
	if err != nil {
		return time.Time{}, err
	}
	if unit == "" {
		return time.Time{}, nil
	}

	switch unit {
	case "day", "days":
		return now.AddDate(0, 0, -count), nil
	case "week", "weeks":
		return now.AddDate(0, 0, -7*count), nil
	case "month", "months":
		return now.AddDate(0, -count, 0), nil
	default:
		return time.Time{}, fmt.Errorf("invalid backup period: %s", period)
	}
}

func backupPrefix(databasePath string) string {
	return filepath.Base(databasePath) + ".backup-"
}

func listBackups(databasePath string) ([]fileInfo, error) {
	dir := filepath.Dir(databasePath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	prefix := backupPrefix(databasePath)
	backups := make([]fileInfo, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		backups = append(backups, fileInfo{Path: filepath.Join(dir, entry.Name()), ModTime: info.ModTime()})
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].ModTime.Before(backups[j].ModTime)
	})
	return backups, nil
}

func due(databasePath string, cfg config.BackupConfig, now time.Time) (bool, []fileInfo, error) {
	backups, err := listBackups(databasePath)
	if err != nil {
		return false, nil, err
	}
	if len(backups) == 0 {
		return true, backups, nil
	}
	cutoff, err := cutoffTime(now, cfg.Period)
	if err != nil {
		return false, nil, err
	}
	return backups[len(backups)-1].ModTime.Before(cutoff), backups, nil
}

func copyFile(srcPath string, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return err
	}

	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return dst.Sync()
}

func rotate(backups []fileInfo, keep int) error {
	if keep <= 0 {
		for _, backup := range backups {
			if err := os.Remove(backup.Path); err != nil {
				return err
			}
		}
		return nil
	}
	if len(backups) <= keep {
		return nil
	}
	for _, backup := range backups[:len(backups)-keep] {
		if err := os.Remove(backup.Path); err != nil {
			return err
		}
	}
	return nil
}

func Run(databasePath string, cfg config.BackupConfig, now time.Time) error {
	if !Enabled(cfg) {
		return nil
	}
	if err := ValidateConfig(cfg); err != nil {
		return err
	}
	if _, err := os.Stat(databasePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	shouldBackup, backups, err := due(databasePath, cfg, now)
	if err != nil {
		return err
	}
	if !shouldBackup {
		return nil
	}

	timestamp := now.Format("20060102-150405")
	backupPath := filepath.Join(filepath.Dir(databasePath), backupPrefix(databasePath)+timestamp)
	if err := copyFile(databasePath, backupPath); err != nil {
		return err
	}

	backups = append(backups, fileInfo{Path: backupPath, ModTime: now})
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].ModTime.Before(backups[j].ModTime)
	})
	return rotate(backups, cfg.Keep)
}

func RunNow(databasePath string, keep int, now time.Time) (string, error) {
	if _, err := os.Stat(databasePath); err != nil {
		return "", err
	}

	backups, err := listBackups(databasePath)
	if err != nil {
		return "", err
	}

	backupPath := filepath.Join(filepath.Dir(databasePath), backupPrefix(databasePath)+now.Format("20060102-150405"))
	if err := copyFile(databasePath, backupPath); err != nil {
		return "", err
	}

	backups = append(backups, fileInfo{Path: backupPath, ModTime: now})
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].ModTime.Before(backups[j].ModTime)
	})
	if keep > 0 {
		if err := rotate(backups, keep); err != nil {
			return "", err
		}
	}

	return backupPath, nil
}

func CopyToPath(databasePath string, outputPath string) error {
	if _, err := os.Stat(databasePath); err != nil {
		return err
	}
	return copyFile(databasePath, outputPath)
}
