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

func Transfer(parsed parser.ParsedCmdLine, cfg config.Config, cashDb *sql.DB) error {
	counts := parsed.GetTokenKindCount(false)
	if len(parsed.Filters) != 0 {
		return fmt.Errorf("no filters allowed")
	}
	if counts[parser.TokenAmount] != 1 {
		return fmt.Errorf("transfer requires exactly one amount")
	}
	if counts[parser.TokenAttribute] < 2 {
		return fmt.Errorf("transfer requires from: and to: attributes")
	}

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
	if counts[parser.TokenText] > 0 {
		textParts := make([]string, 0, counts[parser.TokenText])
		for _, arg := range parsed.Args {
			if arg.Kind == parser.TokenText {
				textParts = append(textParts, arg.Raw)
			}
		}
		if len(textParts) > 0 {
			description = strings.Join(textParts, " ")
		}
	}

	tx, err := cashDb.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	fromResult, err := tx.Exec(`
INSERT INTO transactions (identifier, type, amount, description, datetime, account_id, category_id, place_id, group_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
`, nextIdentifier.String(), "transfer_out", -float64(amount), description, transactionTime, fromAccount.ID, nil, transferPlace.ID, nil)
	if err != nil {
		return err
	}
	fromTransactionID, err := fromResult.LastInsertId()
	if err != nil {
		return err
	}

	toResult, err := tx.Exec(`
INSERT INTO transactions (identifier, type, amount, description, datetime, account_id, category_id, place_id, group_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
`, secondIdentifier.String(), "transfer_in", float64(amount), description, transactionTime, toAccount.ID, nil, transferPlace.ID, nil)
	if err != nil {
		return err
	}
	toTransactionID, err := toResult.LastInsertId()
	if err != nil {
		return err
	}

	transferResult, err := tx.Exec(`
INSERT INTO transfers (from_transaction_id, to_transaction_id, from_account_id, to_account_id, amount)
VALUES (?, ?, ?, ?, ?)
`, fromTransactionID, toTransactionID, fromAccount.ID, toAccount.ID, float64(amount))
	if err != nil {
		return err
	}

	transferID, err := transferResult.LastInsertId()
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	pterm.Success.Println("Transfer added with id: " + strconv.FormatInt(transferID, 10))
	return nil
}
