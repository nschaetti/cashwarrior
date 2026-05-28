package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/domain"
	"github.com/nschaetti/cashwarrior/internal/parser"
	"github.com/pterm/pterm"
)

func Transfer(parsed parser.ParsedCmdLine, cfg config.Config, cashDb db.DBTX) error {
	attributes := getAttributes(parsed)
	fromName, ok := attributes["from"]
	if !ok || fromName == "" {
		return fmt.Errorf("missing from: account")
	}
	toName, ok := attributes["to"]
	if !ok || toName == "" {
		return fmt.Errorf("missing to: account")
	}
	if fromName == toName {
		return fmt.Errorf("from and to accounts must be different")
	}

	amount := parsed.GetAmounts()[0].Amount
	if amount <= 0 {
		return fmt.Errorf("transfer amount must be positive")
	}

	fromAccount, err := db.GetAccountByName(cashDb, fromName)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("account %s does not exist", fromName)
	}
	if err != nil {
		return err
	}

	toAccount, err := db.GetAccountByName(cashDb, toName)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("account %s does not exist", toName)
	}
	if err != nil {
		return err
	}

	transferPlace, err := db.GetPlaceByName(cashDb, "transfer")
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("special place 'transfer' does not exist")
	}
	if err != nil {
		return err
	}

	nextIdentifier, err := getNextIdentifier(cashDb)
	if err != nil {
		return err
	}
	secondIdentifier := domain.TransactionID{Year: nextIdentifier.Year, Month: nextIdentifier.Month, Num: nextIdentifier.Num + 1}

	transactionTime, err := getTransactionDatetime(attributes, cfg)
	if err != nil {
		return err
	}

	description := "Transfer"
	textParts := make([]string, 0)
	for _, arg := range parsed.Args {
		if arg.Kind == parser.TokenText {
			textParts = append(textParts, arg.Raw)
		}
	}
	if len(textParts) > 0 {
		description = strings.Join(textParts, " ")
	}

	fromTransactionID, err := db.InsertTransaction(cashDb, db.CreateTransactionInput{
		Identifier:  nextIdentifier.String(),
		Type:        "transfer_out",
		Amount:      -float64(amount),
		Description: description,
		Datetime:    transactionTime,
		AccountID:   fromAccount.ID,
		PlaceID:     &transferPlace.ID,
	})
	if err != nil {
		return err
	}

	toTransactionID, err := db.InsertTransaction(cashDb, db.CreateTransactionInput{
		Identifier:  secondIdentifier.String(),
		Type:        "transfer_in",
		Amount:      float64(amount),
		Description: description,
		Datetime:    transactionTime,
		AccountID:   toAccount.ID,
		PlaceID:     &transferPlace.ID,
	})
	if err != nil {
		return err
	}

	transferID, err := db.InsertTransfer(cashDb, db.CreateTransferInput{
		FromTransactionID: fromTransactionID,
		ToTransactionID:   toTransactionID,
		FromAccountID:     fromAccount.ID,
		ToAccountID:       toAccount.ID,
		Amount:            float64(amount),
	})
	if err != nil {
		return err
	}

	pterm.Success.Println("Transfer added with id: " + strconv.FormatInt(transferID, 10))
	return nil
}
