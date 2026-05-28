package cmd

import (
	"fmt"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Restore(parsed parser.ParsedCmdLine, cfg config.Config, query db.DBTX) error {
	if parsed.Subcommand == "list" {
		return listDeletedTransactions(cfg, query)
	}

	transaction, err := db.GetTransactionByIdentifier(query, parsed.Args[0].Raw)
	if err != nil {
		return err
	}
	if err := db.UpdateTransactionDeleted(query, transaction.ID, false); err != nil {
		return err
	}
	fmt.Printf("Transaction %s restored\n", transaction.Identifier)
	return nil
}
