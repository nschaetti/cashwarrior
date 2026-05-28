package cmd

import (
	"fmt"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Delete(parsed parser.ParsedCmdLine, cfg config.Config, query db.DBTX) error {
	switch parsed.Subcommand {
	case "list":
		return listDeletedTransactions(cfg, query)
	default:
		transaction, err := db.GetTransactionByIdentifier(query, parsed.Args[0].Raw)
		if err != nil {
			return err
		}
		if err := db.UpdateTransactionDeleted(query, transaction.ID, true); err != nil {
			return err
		}
		fmt.Printf("Transaction %s deleted\n", transaction.Identifier)
		return nil
	}
}
