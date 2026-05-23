package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"

	"github.com/nschaetti/cashwarrior/cmd"
	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func main() {
	// Check configuration exists
	cfg, err := config.InitConfig()
	if err != nil {
		fmt.Println("Error creating config file:", err)
		os.Exit(1)
	}

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
		fmt.Println("Error:", parseErr.Message)
		os.Exit(1)
	}

	// Open the database
	fmt.Println("Opening database:", cfg.Database)
	cashDb, dErr := db.Open(cfg.Database)
	if dErr != nil {
		fmt.Println("Error opening database:", dErr)
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
		fmt.Println(dispatchErr)
	}
}
