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

func configValues(cfg config.Config) map[string]any {
	return map[string]any{
		"database":          cfg.Database,
		"default.currency":  cfg.Default.Currency,
		"default.account":   cfg.Default.Account,
		"gui.date_format":   cfg.Display.DateFormat,
		"gui.show_currency": cfg.Display.ShowCurrency,
		"gui.theme":         cfg.Display.Theme,
		"backup.period":     cfg.Backup.Period,
		"backup.keep":       cfg.Backup.Keep,
	}
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

	switch parsed.Subcommand {
	case "get":
		return getConfigValue(parsed, cfg)
	case "set":
		return setConfigValue(parsed, &cfg, configPath, parsed.Args[0].RawString(), parsed.Args[1].RawString(), cashDb)
	default:
		if len(parsed.Args) == 0 {
			return printConfig(parsed, cfg)
		}
		attr, ok := parsed.Args[0].(parser.ArgAttribute)
		if !ok {
			return fmt.Errorf("config key must be an attribute")
		}
		return setConfigValue(parsed, &cfg, configPath, attr.Key, attr.Value.Raw, cashDb)
	}
}

func applyConfigValue(cfg *config.Config, key, value string, cashDb db.DBTX) error {
	switch key {
	case "database":
		dbPath := utils.ExpandPath(value)
		previousPath := cfg.Database
		if err := ensureDatabasePath(*cfg, dbPath); err != nil {
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
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("default.currency cannot be empty")
		}
		cfg.Default.Currency = value

	case "default.account":
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("default.account cannot be empty")
		}
		_, err := db.GetAccountByName(cashDb, value)
		if err != nil {
			return fmt.Errorf("account does not exist: %s", value)
		}
		cfg.Default.Account = value

	case "gui.date_format":
		v := value
		required := []string{"2006", "01", "02", "15", "04"}
		for _, token := range required {
			if !strings.Contains(v, token) {
				return fmt.Errorf("invalid gui.date_format: must contain 2006, 01, 02, 15 and 04")
			}
		}
		cfg.Display.DateFormat = v

	case "gui.show_currency":
		v, parseErr := strconv.ParseBool(value)
		if parseErr != nil {
			return fmt.Errorf("invalid gui.show_currency: expected boolean")
		}
		cfg.Display.ShowCurrency = v

	case "gui.theme":
		if !gui.ThemeExists(value) {
			themes := gui.ThemeNames()
			sort.Strings(themes)
			return fmt.Errorf("unknown theme %q (available: %s)", value, strings.Join(themes, ", "))
		}
		cfg.Display.Theme = value

	case "backup.period":
		cfg.Backup.Period = value
		if err := backup.ValidateConfig(cfg.Backup); err != nil {
			return err
		}

	case "backup.keep":
		keep, parseErr := strconv.Atoi(value)
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
	return nil
}

func setConfigValue(parsed parser.ParsedCmdLine, cfg *config.Config, configPath, key, value string, cashDb db.DBTX) error {
	if err := requireYesForJSON(parsed); err != nil {
		return err
	}
	if key == "database" && isJSONOutput(parsed) {
		info, err := os.Stat(utils.ExpandPath(value))
		if err != nil || info.IsDir() {
			return fmt.Errorf("database file does not exist: %s", value)
		}
	}
	previousDatabase := cfg.Database
	if err := applyConfigValue(cfg, key, value, cashDb); err != nil {
		return err
	}
	if key == "database" && cfg.Database == previousDatabase {
		if _, err := os.Stat(utils.ExpandPath(value)); os.IsNotExist(err) {
			return nil
		}
	}
	if err := config.SaveConfig(configPath, *cfg); err != nil {
		return err
	}
	if isJSONOutput(parsed) {
		return renderJSON("config", map[string]any{"action": "set", "key": key, "value": value}, 1)
	}
	pterm.Success.Println("Config updated: " + key + "=" + value)
	return nil
}

func printConfig(parsed parser.ParsedCmdLine, cfg config.Config) error {
	values := configValues(cfg)
	if isJSONOutput(parsed) {
		return renderJSON("config", values, len(values))
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("%s = %v\n", key, values[key])
	}
	return nil
}

func getConfigValue(parsed parser.ParsedCmdLine, cfg config.Config) error {
	key := parsed.Args[0].RawString()
	value, ok := configValues(cfg)[key]
	if !ok {
		return fmt.Errorf("unknown config key: %s", key)
	}
	if isJSONOutput(parsed) {
		return renderJSON("config", map[string]any{key: value}, 1)
	}
	fmt.Printf("%s = %v\n", key, value)
	return nil
}
