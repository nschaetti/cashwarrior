package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	_ "modernc.org/sqlite"

	"github.com/nschaetti/cashwarrior/cmd"
	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func printLogo(theme gui.Theme) {
	titleColor, err := gui.HexToRGB(theme.CashWarriorTitle)
	if err != nil {
		pterm.Error.Println("Error parsing theme color: ", err)
		os.Exit(1)
	}
	srender, err2 := pterm.DefaultBigText.WithLetters(
		putils.LettersFromStringWithRGB("cashwarrior", titleColor),
	).Srender()
	if err2 != nil {
		return
	}
	fmt.Println()
	fmt.Println(srender)
}

func closeAndRollback(cashDb *sql.DB, tx *sql.Tx) {
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		fmt.Println("Error rolling back transaction: ", err)
	}
	if err := cashDb.Close(); err != nil {
		fmt.Println("Error closing database: ", err)
	}
}

func run() error {
	// Check configuration exists
	cfg, err := config.InitConfig()
	if err != nil {
		return fmt.Errorf("error creating config file: %w", err)
	}

	// Theme
	gui.SetTheme(cfg.Display.Theme)
	theme := gui.CurrentTheme()
	gui.ApplyPTermTheme(theme)

	// Print logo
	printLogo(theme)

	// Start time & theme
	pterm.Info.Println("Using theme ", cfg.Display.Theme)
	start := time.Now()

	// Parse the command line
	parsedCmd, parseErr := parser.ParseAndValidateCmdLine(os.Args[1:])
	if parseErr != nil && parseErr.Code != parser.ParseErrorNoCommand {
		// Just call "list" if there is no command.
		parsedCmd = parser.ParsedCmdLine{
			Command: "list",
			Filters: []parser.Token{},
			Args:    []parser.Token{},
		}
	} else if parseErr != nil {
		return fmt.Errorf("error: %s", parseErr.Message)
	}

	// Open the database
	pterm.Info.Println("Using SQLite database ", cfg.Database)
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
	pterm.Success.Println("Command success, done in ", elapsed)
	return nil
}

func main() {
	if err := run(); err != nil {
		pterm.Error.Println(err)
		os.Exit(1)
	}
}
