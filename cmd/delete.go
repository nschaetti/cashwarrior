package cmd

import (
	"fmt"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Delete(parsed parser.ParsedCmdLine, _ config.Config, query db.DBTX) error {
	_ = parsed
	_ = query
	fmt.Println("delete")
	return nil
}
