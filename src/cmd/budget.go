package cmd

import (
	"database/sql"
	"fmt"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Budget(parsed parser.ParsedCmdLine, _ config.Config, db *sql.DB) error {
	fmt.Println("budget")
	return nil
}
