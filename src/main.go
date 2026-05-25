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
		pterm.Error.Println("Error parsing theme color: ", err, "\n")
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

func main() {
	// Check configuration exists
	cfg, err := config.InitConfig()
	if err != nil {
		pterm.Error.Println("Error creating config file: ", err, "\n")
		os.Exit(1)
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
		pterm.Error.Println("Error: ", parseErr.Message, "\n")
		os.Exit(1)
	}

	// Open the database
	pterm.Info.Println("Using SQLite database ", cfg.Database)
	cashDb, dErr := db.Open(cfg)
	if dErr != nil {
		pterm.Error.Println("Error opening database: ", dErr, "\n")
		os.Exit(1)
	}
	defer func(mdb *sql.DB) {
		err := db.Close(mdb)
		if err != nil {

		}
	}(cashDb)

	// Dispatch the command
	dispatchErr := cmd.Dispatch(parsedCmd, cfg, cashDb)
	if dispatchErr != nil {
		pterm.Error.Println("Command error: ", dispatchErr, "\n")
		os.Exit(1)
	}

	elapsed := time.Since(start)
	pterm.Success.Println("Command success, done in ", elapsed, "\n")
	os.Exit(0)
}
