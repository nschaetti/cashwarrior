package cmd

import (
	"database/sql"
	"fmt"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Show(parsed parser.ParsedCmdLine, _ config.Config, db *sql.DB) error {
	fmt.Println("show")
	return nil
}
