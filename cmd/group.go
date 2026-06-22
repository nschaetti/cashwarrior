package cmd

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/domain"
	"github.com/nschaetti/cashwarrior/internal/parser"
	"github.com/pterm/pterm"
)

func parseGroupArgs(parsed parser.ParsedCmdLine) ([]string, string, error) {
	transactionRefs := make([]string, 0, len(parsed.Args))
	groupName := ""
	transactionID := ""

	for _, arg := range parsed.Args {
		if arg.ArgKind() == parser.ArgKindAttribute {
			switch arg.(parser.ArgAttribute).Key {
			case "group":
				if groupName != "" {
					return nil, "", fmt.Errorf("multiple groups given")
				}
				groupName = arg.(parser.ArgAttribute).Value.Value.(parser.StringItem).Value
				break
			case "T", "id", "identifier":
				transactionID = arg.(parser.ArgAttribute).Value.Value.(parser.StringItem).Value
				if isTransactionReference(transactionID) {
					transactionRefs = append(transactionRefs, transactionID)
					continue
				}
				break
			}
		}
	}

	if len(transactionRefs) == 0 {
		return nil, "", fmt.Errorf("no transaction given")
	}
	if groupName == "" {
		return nil, "", fmt.Errorf("no group given")
	}

	return transactionRefs, groupName, nil
}

func isTransactionReference(raw string) bool {
	_, err := domain.ParseTransactionID(raw)
	return err == nil
}

func ensureTransactionGroup(cashDb db.DBTX, groupName string) (int64, error) {
	group, err := db.GetGroupByName(cashDb, groupName)
	if errors.Is(err, sql.ErrNoRows) {
		return db.InsertTransactionGroup(cashDb, db.CreateTransactionGroupInput{Name: groupName})
	}
	if err != nil {
		return 0, err
	}
	return group.ID, nil
}

func getTransactionByReference(cashDb db.DBTX, transactionRef string) (db.Transaction, error) {
	transaction, err := db.GetTransactionByIdentifier(cashDb, transactionRef)
	if errors.Is(err, sql.ErrNoRows) {
		return db.Transaction{}, fmt.Errorf("transaction %s does not exist", transactionRef)
	}
	if err != nil {
		return db.Transaction{}, err
	}
	return transaction, nil
}

func Group(parsed parser.ParsedCmdLine, _ config.Config, cashDb db.DBTX) error {
	transactionRefs, groupName, err := parseGroupArgs(parsed)
	if err != nil {
		return err
	}

	pterm.FgWhite.Println("Transaction to be added:")
	pterm.FgWhite.Println("========================")
	pterm.FgWhite.Println("Group: ", groupName, "")
	for _, transactionRef := range transactionRefs {
		pterm.FgWhite.Println("Transaction: ", transactionRef, "")
	}

	// Confirm grouping
	ok, err := pterm.DefaultInteractiveConfirm.
		WithDefaultText("Confirm grouping (N/y) ?").
		Show()
	if err != nil {
		panic(fmt.Errorf("error confirming grouping: %w", err))
	}

	if ok {
		// Get or create group
		var groupID int64
		groupID, err = ensureTransactionGroup(cashDb, groupName)
		if err != nil {
			return err
		}

		linkedCount := 0
		var transaction db.Transaction
		for _, transactionRef := range transactionRefs {
			transaction, err = getTransactionByReference(cashDb, transactionRef)
			if err != nil {
				return err
			}
			if err = db.UpdateTransactionGroupID(cashDb, transaction.ID, &groupID); err != nil {
				return err
			}
			linkedCount++
		}

		pterm.Success.Printf("Linked %d transactions to group %s\n", linkedCount, groupName)
	} else {
		pterm.Warning.Println("Aborted grouping")
	}
	return nil
}
