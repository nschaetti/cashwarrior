package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	_ "modernc.org/sqlite"

	"github.com/nschaetti/cashwarrior/cmd"
	"github.com/nschaetti/cashwarrior/internal/backup"
	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/output"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func printLogo(theme gui.Theme) error {
	titleColor, err := gui.HexToRGB(theme.CashWarriorTitle)
	if err != nil {
		return fmt.Errorf("error parsing theme color: %w", err)
	}
	srender, err2 := pterm.DefaultBigText.WithLetters(
		putils.LettersFromStringWithRGB("cashwarrior", titleColor),
	).Srender()
	if err2 != nil {
		return err2
	}
	fmt.Println()
	fmt.Println(srender)
	return nil
}

func closeAndRollback(cashDb *sql.DB, tx *sql.Tx) {
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		fmt.Fprintln(os.Stderr, "Error rolling back transaction:", err)
	}
	if err := cashDb.Close(); err != nil {
		fmt.Fprintln(os.Stderr, "Error closing database:", err)
	}
}

func detectOutputFormat(args []string) (output.Format, error) {
	format := output.FormatTable
	jsonRequested := false
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch {
		case arg == "--json":
			jsonRequested = true
		case strings.HasPrefix(arg, "--format="):
			value := strings.TrimPrefix(arg, "--format=")
			parsed, err := output.ParseFormat(value)
			if err != nil {
				return "", err
			}
			format = parsed
		case arg == "--format":
			if index+1 >= len(args) {
				return "", fmt.Errorf("missing value for --format")
			}
			index++
			parsed, err := output.ParseFormat(args[index])
			if err != nil {
				return "", err
			}
			format = parsed
		}
	}
	if jsonRequested {
		return output.FormatJSON, nil
	}
	return format, nil
}

func createHelpCmdLine() parser.ParsedCmdLine {
	return parser.ParsedCmdLine{
		Command:    "help",
		Subcommand: "show",
		Filters:    []parser.Arg{},
		Args:       []parser.Arg{},
		Flags:      []parser.Arg{parser.ArgFlag{Raw: "--help", Key: "help", Value: parser.BoolItem{Raw: "true", Value: true}}},
	}
}

func run() error {
	argv := os.Args[1:]
	format, err := detectOutputFormat(argv)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}
	machineOutput := format == output.FormatJSON

	// Check configuration exists
	cfg, configErr := config.InitConfig()
	if configErr != nil {
		return fmt.Errorf("error initializing configuration: %w", configErr)
	}

	// Parse the command line before producing any decorative output.
	parsedCmd, parseErr := parser.ParseAndValidateCmdLine(argv, cfg)
	if parseErr != nil {
		if parseErr.Code == parser.ParseErrorEmptyCommandLine || (parseErr.Code == parser.ParseErrorNoCommand && containsHelpFlag(argv)) {
			parsedCmd = createHelpCmdLine()
		} else {
			return fmt.Errorf("error parsing command line: %w", parseErr)
		}
	}

	// Theme
	gui.SetTheme(cfg.Display.Theme)
	theme := gui.CurrentTheme()
	gui.ApplyPTermTheme(theme)

	// Print logo
	if !machineOutput && cfg.Display.ShowHeader {
		if err := printLogo(theme); err != nil {
			return err
		}
	}

	// Backup
	if rErr := backup.Run(cfg.Database, cfg.Backup, time.Now()); rErr != nil {
		fmt.Fprintln(os.Stderr, "Automatic backup failed:", rErr)
	}

	// Start time & theme
	if !machineOutput && cfg.Display.ShowInfo {
		pterm.Info.Println("Using theme ", cfg.Display.Theme)
	}
	start := time.Now()

	// Open the database
	if !machineOutput && cfg.Display.ShowInfo {
		pterm.Info.Println("Using SQLite database ", cfg.Database)
	}
	cashDb, dErr := db.Open(cfg)
	if dErr != nil {
		return fmt.Errorf("error opening database: %w", dErr)
	}

	// Now we can begin SQL
	tx, berr := cashDb.Begin()
	if berr != nil {
		_ = cashDb.Close()
		return fmt.Errorf("error opening transaction: %w", berr)
	}
	defer closeAndRollback(cashDb, tx)

	// Dispatch the command
	dispatchErr := cmd.Dispatch(parsedCmd, cfg, tx)
	if dispatchErr != nil {
		return fmt.Errorf("command error: %w", dispatchErr)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	elapsed := time.Since(start)
	if !machineOutput {
		pterm.Success.Println("Command success, done in ", elapsed)
	}
	return nil
}

func containsHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
