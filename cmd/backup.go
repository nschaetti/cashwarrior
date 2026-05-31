package cmd

import (
	"fmt"
	"time"

	backuputil "github.com/nschaetti/cashwarrior/internal/backup"
	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Backup(parsed parser.ParsedCmdLine, cfg config.Config, _ db.DBTX) error {
	attributes := getAttributes(parsed)
	if outputPath := attributes["output"]; outputPath != "" {
		if err := backuputil.CopyToPath(cfg.Database, outputPath); err != nil {
			return err
		}
		fmt.Printf("Database backed up to %s\n", outputPath)
		return nil
	}

	keep := cfg.Backup.Keep
	if keep <= 0 {
		keep = 1
	}
	backupPath, err := backuputil.RunNow(cfg.Database, keep, time.Now())
	if err != nil {
		return err
	}
	fmt.Printf("Database backed up to %s\n", backupPath)
	return nil
}
