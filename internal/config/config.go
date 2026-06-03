package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nschaetti/cashwarrior/internal/utils"
	"github.com/pelletier/go-toml/v2"
)

const DefaultConfigFile = "~/.config/cashwarrior/config.toml"
const DefaultDatabaseRelPath = "~/.cashwarrior/cash.db"

type ConfigErrorCode string

const (
	ErrorConfigFileExistsButDir ConfigErrorCode = "CONFIG_IS_DIR"
	ErrorConfigInvalidFile      ConfigErrorCode = "INVALID_CONFIG_FILE"
	ErrorConfigCannotCreateFile ConfigErrorCode = "CANNOT_CREATE_CONFIG_FILE"
)

type ConfigError struct {
	Code    ConfigErrorCode
	Message string
}

func (e ConfigError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

type Config struct {
	Database string        `toml:"database"`
	Default  AccountConfig `toml:"default"`
	Display  DisplayConfig `toml:"gui"`
	Backup   BackupConfig  `toml:"backup"`
}

type AccountConfig struct {
	Currency string `toml:"currency"`
	Account  string `toml:"account"`
}

type DisplayConfig struct {
	DateFormat   string `toml:"date_format"`
	ShowCurrency bool   `toml:"show_currency"`
	Theme        string `toml:"theme"`
	ShowHeader   bool   `toml:"show_header"`
	ShowInfo     bool   `toml:"show_info"`
}

type BackupConfig struct {
	Period string `toml:"period"`
	Keep   int    `toml:"keep"`
}

func (c Config) String() string {
	return fmt.Sprintf(
		"<Config: database_path=%s, default=%s, display=%s, backup=%s>",
		c.Database,
		c.Default,
		c.Display,
		c.Backup,
	)
}

func (c AccountConfig) String() string {
	return fmt.Sprintf("<AccountConfig: currency=%s, account=%s>", c.Currency, c.Account)
}

func (c DisplayConfig) String() string {
	return fmt.Sprintf(
		"<DisplayConfig: date_format=%s, show_currency=%t, theme=%s, show_header=%v, show_info=%v>",
		c.DateFormat,
		c.ShowCurrency,
		c.Theme,
		c.ShowHeader,
		c.ShowInfo,
	)
}

func (c BackupConfig) String() string {
	return fmt.Sprintf("<BackupConfig: period=%s, keep=%d>", c.Period, c.Keep)
}

func LoadConfig(path string) (Config, error) {
	var cfg Config

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	err = toml.Unmarshal(data, &cfg)
	return cfg, err
}

func GetDefaultConfig() Config {
	return Config{
		Database: utils.ExpandPath(DefaultDatabaseRelPath),
		Default: AccountConfig{
			Currency: "USD",
			Account:  "main",
		},
		Display: DisplayConfig{
			DateFormat:   "2006-01-02",
			ShowCurrency: true,
			Theme:        "dracula",
		},
		Backup: BackupConfig{},
	}
}

func ConfigFileExists(path string) (os.FileInfo, bool) {
	info, err := os.Stat(path)
	return info, !os.IsNotExist(err)
}

func askSampleCreation() bool {
	message := fmt.Sprintf(`A configuration file could not be found in
	
%s
`, DefaultConfigFile)
	fmt.Println(message)
	return utils.AskYesNo("Would you like a sample to be created? ")
}

func askDBPath() string {
	var dbPath string
	var err error
	for {
		dbPath, err = utils.AskPath("Path to the database", DefaultDatabaseRelPath)
		if err == nil {
			break
		}
		fmt.Println("Invalid path. Please try again: ", err)
	}
	return dbPath
}

func askCurrency() string {
	var currency string
	var err error
	for {
		currency, err = utils.Ask("Default currency", "USD", true)
		if err == nil {
			break
		}
		fmt.Println("Invalid currency. Please try again: ", err)
	}
	return currency
}

func askAccountName() string {
	var accountName string
	var err error
	for {
		accountName, err = utils.Ask("Main account name", "main", true)
		if err == nil {
			break
		}
		fmt.Println("Invalid account name. Please try again: ", err)
	}
	return accountName
}

func askDateFormat() string {
	var dateFormat string
	var err error
	for {
		dateFormat, err = utils.Ask("Date format (Y: Year, M: Month, D: Day)", "YYYY-MM-DD", true)
		if err == nil {
			break
		}
		fmt.Println("Invalid date format. Please try again: ", err)
	}
	return dateFormat
}

func configCreation() (Config, *ConfigError) {
	// Create a config file?
	createConfig := askSampleCreation()
	if !createConfig {
		return Config{}, &ConfigError{
			Code:    ErrorConfigInvalidFile,
			Message: "config file does not exist",
		}
	}

	// Ask for database path
	dbPath := utils.ExpandPath(askDBPath())
	currency := askCurrency()
	accountName := askAccountName()
	dateFormat := askDateFormat()

	return Config{
		Database: dbPath,
		Default: AccountConfig{
			Currency: currency,
			Account:  accountName,
		},
		Display: DisplayConfig{
			DateFormat:   dateFormat,
			ShowCurrency: true,
			Theme:        "dracula",
		},
		Backup: BackupConfig{},
	}, nil
}

func SaveConfig(path string, config Config) error {
	data, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	fmt.Println("Saving config file to", path)
	return os.WriteFile(path, data, 0644)
}

func InitConfig() (Config, *ConfigError) {
	// Check if config file exists
	info, exists := ConfigFileExists(utils.ExpandPath(DefaultConfigFile))

	// It exists, load it
	if exists {
		if info.IsDir() {
			return Config{}, &ConfigError{
				Code:    ErrorConfigFileExistsButDir,
				Message: "config file is a directory"}
		}
		config, err := LoadConfig(utils.ExpandPath(DefaultConfigFile))
		if err != nil {
			return Config{}, &ConfigError{
				Code:    ErrorConfigInvalidFile,
				Message: fmt.Sprintf("invalid config file: %v", err),
			}
		}
		return config, nil
	}

	// It doesn't exist, create it'
	config, err := configCreation()
	if err != nil {
		return Config{}, err
	}

	// Save it
	cErr := SaveConfig(utils.ExpandPath(DefaultConfigFile), config)
	if cErr != nil {
		return Config{}, &ConfigError{
			Code:    ErrorConfigInvalidFile,
			Message: fmt.Sprintf("failed to save config file: %v", err),
		}
	}

	return config, nil
}
