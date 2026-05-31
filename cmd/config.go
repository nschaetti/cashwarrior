package cmd

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/nschaetti/cashwarrior/internal/backup"
	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/parser"
	"github.com/nschaetti/cashwarrior/internal/utils"
	"github.com/pterm/pterm"
)

func printConfigHelp() {
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  cash config key:value")
	fmt.Println()
	fmt.Println("Parameters:")
	fmt.Println("  database                Path to SQLite database file")
	fmt.Println("  default.currency        Default currency code string (non-empty)")
	fmt.Println("  default.account         Default account name (must exist)")
	fmt.Println("  gui.date_format         Go time format containing 2006, 01, 02, 15, 04")
	fmt.Println("  gui.show_currency       Boolean: true or false")
	fmt.Println("  gui.theme               Theme name (available themes only)")
	fmt.Println("  backup.period           day, 2days, week, 2weeks, month, 2months, ...")
	fmt.Println("  backup.keep             Number of backup files to keep")
	fmt.Println()
}

func ensureDatabasePath(cfg config.Config, dbPath string) error {
	info, statErr := os.Stat(dbPath)
	if statErr == nil {
		if info.IsDir() {
			return fmt.Errorf("database path is a directory: %s", dbPath)
		}
		cfg.Database = dbPath
		return nil
	}

	if !os.IsNotExist(statErr) {
		return statErr
	}

	if !utils.AskYesNo("Database file does not exist. Create and initialize it?") {
		return nil
	}

	cfg.Database = dbPath
	newDB, err := db.Open(cfg)
	if err != nil {
		return err
	}
	return newDB.Close()
}

func Config(parsed parser.ParsedCmdLine, _ config.Config, cashDb db.DBTX) error {
	if len(parsed.Args) == 0 {
		printConfigHelp()
		return nil
	}

	arg := parsed.Args[0]

	configPath := utils.ExpandPath(config.DefaultConfigFile)
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	switch arg.Key {
	case "database":
		dbPath := utils.ExpandPath(arg.Value)
		previousPath := cfg.Database
		if err := ensureDatabasePath(cfg, dbPath); err != nil {
			return err
		}
		if _, statErr := os.Stat(dbPath); statErr == nil {
			cfg.Database = dbPath
		} else if os.IsNotExist(statErr) {
			if previousPath == cfg.Database {
				return nil
			}
		}

	case "default.currency":
		if strings.TrimSpace(arg.Value) == "" {
			return fmt.Errorf("default.currency cannot be empty")
		}
		cfg.Default.Currency = arg.Value

	case "default.account":
		if strings.TrimSpace(arg.Value) == "" {
			return fmt.Errorf("default.account cannot be empty")
		}
		_, err = db.GetAccountByName(cashDb, arg.Value)
		if err != nil {
			return fmt.Errorf("account does not exist: %s", arg.Value)
		}
		cfg.Default.Account = arg.Value

	case "gui.date_format":
		v := arg.Value
		required := []string{"2006", "01", "02", "15", "04"}
		for _, token := range required {
			if !strings.Contains(v, token) {
				return fmt.Errorf("invalid gui.date_format: must contain 2006, 01, 02, 15 and 04")
			}
		}
		cfg.Display.DateFormat = v

	case "gui.show_currency":
		v, parseErr := strconv.ParseBool(arg.Value)
		if parseErr != nil {
			return fmt.Errorf("invalid gui.show_currency: expected boolean")
		}
		cfg.Display.ShowCurrency = v

	case "gui.theme":
		if !gui.ThemeExists(arg.Value) {
			themes := gui.ThemeNames()
			sort.Strings(themes)
			return fmt.Errorf("unknown theme %q (available: %s)", arg.Value, strings.Join(themes, ", "))
		}
		cfg.Display.Theme = arg.Value

	case "backup.period":
		cfg.Backup.Period = arg.Value
		if err := backup.ValidateConfig(cfg.Backup); err != nil {
			return err
		}

	case "backup.keep":
		keep, parseErr := strconv.Atoi(arg.Value)
		if parseErr != nil {
			return fmt.Errorf("invalid backup.keep: expected integer")
		}
		cfg.Backup.Keep = keep
		if err := backup.ValidateConfig(cfg.Backup); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown config key: %s", arg.Key)
	}

	if err = config.SaveConfig(configPath, cfg); err != nil {
		return err
	}

	pterm.Success.Println("Config updated: " + arg.Key + "=" + arg.Value)
	return nil
}
