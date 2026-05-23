package cmd

import (
	"database/sql"
	"fmt"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Transfer(parsed parser.ParsedCmdLine, _ config.Config, db *sql.DB) error {
	fmt.Println("transfer")
	return nil
}
