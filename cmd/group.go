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

func Group(parsed parser.ParsedCmdLine, _ config.Config, cashDb db.DBTX) error {
	transactionRefs, groupName, err := parseGroupArgs(parsed)
	if err != nil {
		return err
	}

	groupID, err := ensureTransactionGroup(cashDb, groupName)
	if err != nil {
		return err
	}

	linkedCount := 0
	for _, transactionRef := range transactionRefs {
		transaction, err := getTransactionByReference(cashDb, transactionRef)
		if err != nil {
			return err
		}
		if err := db.UpdateTransactionGroupID(cashDb, transaction.ID, &groupID); err != nil {
			return err
		}
		linkedCount++
	}

	pterm.Success.Printf("Linked %d transactions to group %s\n", linkedCount, groupName)
	return nil
}

func parseGroupArgs(parsed parser.ParsedCmdLine) ([]string, string, error) {
	transactionRefs := make([]string, 0, len(parsed.Args))
	groupName := ""

	for _, arg := range parsed.Args {
		if isTransactionReference(arg.Raw) {
			transactionRefs = append(transactionRefs, arg.Raw)
			continue
		}

		if groupName != "" {
			return nil, "", fmt.Errorf("multiple groups given")
		}
		groupName = arg.Raw
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
	if len(raw) < 2 || raw[0] != 'T' {
		return false
	}
	_, err := domain.ParseTransactionID(raw[1:])
	return err == nil
}

func ensureTransactionGroup(cashDb db.DBTX, groupName string) (int64, error) {
	group, err := db.GetTransactionGroupByName(cashDb, groupName)
	if errors.Is(err, sql.ErrNoRows) {
		return db.InsertTransactionGroup(cashDb, db.CreateTransactionGroupInput{Name: groupName})
	}
	if err != nil {
		return 0, err
	}
	return group.ID, nil
}

func getTransactionByReference(cashDb db.DBTX, transactionRef string) (db.Transaction, error) {
	transaction, err := db.GetTransactionByIdentifier(cashDb, transactionRef[1:])
	if errors.Is(err, sql.ErrNoRows) {
		return db.Transaction{}, fmt.Errorf("transaction %s does not exist", transactionRef)
	}
	if err != nil {
		return db.Transaction{}, err
	}
	return transaction, nil
}
