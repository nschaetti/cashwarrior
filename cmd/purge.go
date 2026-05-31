package cmd

import (
	"fmt"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Purge(parsed parser.ParsedCmdLine, _ config.Config, query db.DBTX) error {
	identifier := parsed.Args[0].Raw
	if err := db.PurgeTransactionByIdentifier(query, identifier); err != nil {
		return err
	}
	fmt.Printf("Transaction %s purged\n", identifier)
	return nil
}
