package cmd

import (
	"database/sql"
	"fmt"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Add(parsed parser.ParsedCmdLine, _ config.Config, db *sql.DB) error {
	// We need at least two arguments
	if len(parsed.Args) < 2 {
		return fmt.Errorf("we need at least an amount and a description")
	}

	// No filter
	if len(parsed.Filters) != 0 {
		fmt.Println("Filter given, ignoring.")
	}
	fmt.Println("add")
	fmt.Println(parsed)
	return nil
}
