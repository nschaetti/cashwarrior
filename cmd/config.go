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
	fmt.Println("  cash config")
	fmt.Println("  cash config get <key>")
	fmt.Println("  cash config set <key> <value>")
	fmt.Println("  cash config <key>:<value>")
	fmt.Println()
	fmt.Println("Parameters:")
	fmt.Println("  database                Path to SQLite database file")
	fmt.Println("  default.currency        Default currency code string (non-empty)")
	fmt.Println("  default.account         Default account name (must exist)")
	fmt.Println("  gui.date_format         Go time format containing 2006, 01, 02 (optionally 15, 04)")
	fmt.Println("  gui.show_currency       Boolean: true or false")
	fmt.Println("  gui.show_header         Boolean: true or false")
	fmt.Println("  gui.show_info           Boolean: true or false")
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
	configPath := utils.ExpandPath(config.DefaultConfigFile)
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	if parsed.Subcommand == "print" {
		fmt.Printf("database: %s\ndefault.currency: %s\ndefault.account: %s\ngui.date_format: %s\ngui.show_currency: %t\ngui.theme: %s\ngui.show_header: %t\ngui.show_info: %t\nbackup.period: %s\nbackup.keep: %d\n",
			cfg.Database, cfg.Default.Currency, cfg.Default.Account, cfg.Display.DateFormat,
			cfg.Display.ShowCurrency, cfg.Display.Theme, cfg.Display.ShowHeader, cfg.Display.ShowInfo,
			cfg.Backup.Period, cfg.Backup.Keep)
		return nil
	}

	if parsed.Subcommand == "get" {
		if len(parsed.Args) != 1 {
			return fmt.Errorf("usage: cash config get <key>")
		}
		key, err := configArgText(parsed.Args[0])
		if err != nil {
			return err
		}
		value, err := configValue(cfg, key)
		if err != nil {
			return err
		}
		fmt.Println(value)
		return nil
	}

	var key, rawValue string
	if len(parsed.Args) == 1 {
		// Compatibility for callers that already construct an attribute token.
		if attr, ok := parsed.Args[0].(parser.ArgAttribute); ok {
			key, rawValue = attr.Key, attr.Value.Raw
		}
	}
	if len(parsed.Args) == 2 {
		key, err = configArgText(parsed.Args[0])
		if err == nil {
			rawValue, err = configArgText(parsed.Args[1])
		}
		if err != nil {
			return err
		}
	}
	if key == "" || rawValue == "" {
		return fmt.Errorf("usage: cash config set <key> <value>")
	}

	switch key {
	case "database":
		dbPath := utils.ExpandPath(rawValue)
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
		if strings.TrimSpace(rawValue) == "" {
			return fmt.Errorf("default.currency cannot be empty")
		}
		cfg.Default.Currency = rawValue

	case "default.account":
		if strings.TrimSpace(rawValue) == "" {
			return fmt.Errorf("default.account cannot be empty")
		}
		_, err = db.GetAccountByName(cashDb, rawValue)
		if err != nil {
			return fmt.Errorf("account does not exist: %s", rawValue)
		}
		cfg.Default.Account = rawValue

	case "gui.date_format":
		v := rawValue
		required := []string{"2006", "01", "02"}
		for _, token := range required {
			if !strings.Contains(v, token) {
				return fmt.Errorf("invalid gui.date_format: must contain 2006, 01 and 02")
			}
		}
		cfg.Display.DateFormat = v

	case "gui.show_currency":
		v, parseErr := strconv.ParseBool(rawValue)
		if parseErr != nil {
			return fmt.Errorf("invalid gui.show_currency: expected boolean")
		}
		cfg.Display.ShowCurrency = v
	case "gui.show_header":
		v, parseErr := strconv.ParseBool(rawValue)
		if parseErr != nil {
			return fmt.Errorf("invalid gui.show_header: expected boolean")
		}
		cfg.Display.ShowHeader = v
	case "gui.show_info":
		v, parseErr := strconv.ParseBool(rawValue)
		if parseErr != nil {
			return fmt.Errorf("invalid gui.show_info: expected boolean")
		}
		cfg.Display.ShowInfo = v

	case "gui.theme":
		if !gui.ThemeExists(rawValue) {
			themes := gui.ThemeNames()
			sort.Strings(themes)
			return fmt.Errorf("unknown theme %q (available: %s)", rawValue, strings.Join(themes, ", "))
		}
		cfg.Display.Theme = rawValue

	case "backup.period":
		cfg.Backup.Period = rawValue
		if err := backup.ValidateConfig(cfg.Backup); err != nil {
			return err
		}

	case "backup.keep":
		keep, parseErr := strconv.Atoi(rawValue)
		if parseErr != nil {
			return fmt.Errorf("invalid backup.keep: expected integer")
		}
		cfg.Backup.Keep = keep
		if err := backup.ValidateConfig(cfg.Backup); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	if err = config.SaveConfig(configPath, cfg); err != nil {
		return err
	}

	pterm.Success.Println("Config updated: " + key + "=" + rawValue)
	return nil
}

func configArgText(arg parser.Arg) (string, error) {
	text, ok := arg.(parser.ArgText)
	if !ok {
		return "", fmt.Errorf("config key and value must be text")
	}
	return text.Text, nil
}

func configValue(cfg config.Config, key string) (string, error) {
	switch key {
	case "database":
		return cfg.Database, nil
	case "default.currency":
		return cfg.Default.Currency, nil
	case "default.account":
		return cfg.Default.Account, nil
	case "gui.date_format":
		return cfg.Display.DateFormat, nil
	case "gui.show_currency":
		return strconv.FormatBool(cfg.Display.ShowCurrency), nil
	case "gui.show_header":
		return strconv.FormatBool(cfg.Display.ShowHeader), nil
	case "gui.show_info":
		return strconv.FormatBool(cfg.Display.ShowInfo), nil
	case "gui.theme":
		return cfg.Display.Theme, nil
	case "backup.period":
		return cfg.Backup.Period, nil
	case "backup.keep":
		return strconv.Itoa(cfg.Backup.Keep), nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}
