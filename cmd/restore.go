package cmd

import (
	"database/sql"
	"errors"
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

	transfer, err := db.GetTransferByTransactionIDIncludingDeleted(query, transaction.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if err == nil {
		if err := db.UpdateTransferDeleted(query, transfer.ID, false); err != nil {
			return err
		}
		if err := db.UpdateTransactionDeleted(query, transfer.FromTransactionID, false); err != nil {
			return err
		}
		if transfer.ToTransactionID != transfer.FromTransactionID {
			if err := db.UpdateTransactionDeleted(query, transfer.ToTransactionID, false); err != nil {
				return err
			}
		}
		fmt.Printf("Transaction %s restored\n", transaction.Identifier)
		return nil
	}

	if err := db.UpdateTransactionDeleted(query, transaction.ID, false); err != nil {
		return err
	}
	fmt.Printf("Transaction %s restored\n", transaction.Identifier)
	return nil
}
