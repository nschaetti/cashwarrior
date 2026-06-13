package cmd

import (
	"database/sql"
	"errors"
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
		transaction, err := db.GetTransactionByIdentifier(query, parsed.Args[0].RawString())
		if err != nil {
			return err
		}

		transfer, err := db.GetTransferByTransactionID(query, transaction.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if err == nil {
			if err := db.UpdateTransferDeleted(query, transfer.ID, true); err != nil {
				return err
			}
			if err := db.UpdateTransactionDeleted(query, transfer.FromTransactionID, true); err != nil {
				return err
			}
			if transfer.ToTransactionID != transfer.FromTransactionID {
				if err := db.UpdateTransactionDeleted(query, transfer.ToTransactionID, true); err != nil {
					return err
				}
			}
			fmt.Printf("Transaction %s deleted\n", transaction.Identifier)
			return nil
		}

		if err := db.UpdateTransactionDeleted(query, transaction.ID, true); err != nil {
			return err
		}
		fmt.Printf("Transaction %s deleted\n", transaction.Identifier)
		return nil
	}
}
