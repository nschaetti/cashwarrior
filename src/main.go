package main

import (
	"database/sql"
	"os"

	"github.com/pterm/pterm"
	_ "modernc.org/sqlite"

	"github.com/nschaetti/cashwarrior/cmd"
	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func main() {
	// Check configuration exists
	cfg, err := config.InitConfig()
	if err != nil {
		pterm.Error.Println("Error creating config file: ", err, "\n")
		os.Exit(1)
	}
	gui.SetTheme(cfg.Display.Theme)

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
	os.Exit(0)
}
