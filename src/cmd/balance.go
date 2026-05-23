package cmd

import (
	"database/sql"
	"fmt"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Balance(parsed parser.ParsedCmdLine, _ config.Config, db *sql.DB) error {
	fmt.Println("balance")
	return nil
}
