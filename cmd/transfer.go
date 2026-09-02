package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/domain"
	"github.com/nschaetti/cashwarrior/internal/output"
	"github.com/nschaetti/cashwarrior/internal/parser"
	"github.com/pterm/pterm"
)

func Transfer(parsed parser.ParsedCmdLine, cfg config.Config, cashDb db.DBTX) error {
	if err := requireYesForJSON(parsed); err != nil {
		return err
	}
	attributes := getAttributes(parsed)
	fromName, ok := attributes["from"]
	if !ok || fromName.Raw == "" {
		return fmt.Errorf("missing from: account")
	}
	toName, ok := attributes["to"]
	if !ok || toName.Raw == "" {
		return fmt.Errorf("missing to: account")
	}
	if fromName.Raw == toName.Raw {
		return fmt.Errorf("from and to accounts must be different")
	}

	amountValue, ok := attributes["amount"]
	if !ok || amountValue.ValueShape != parser.AttributeValueShapeSingle {
		return fmt.Errorf("transfer requires an amount")
	}
	amountItem, ok := amountValue.Value.(parser.FloatItem)
	if !ok || amountItem.Value <= 0 {
		return fmt.Errorf("transfer amount must be positive")
	}

	fromAccount, err := db.GetAccountByName(cashDb, fromName.Raw)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("account %s does not exist", fromName.Raw)
	}
	if err != nil {
		return err
	}
	toAccount, err := db.GetAccountByName(cashDb, toName.Raw)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("account %s does not exist", toName.Raw)
	}
	if err != nil {
		return err
	}
	if fromAccount.Currency != toAccount.Currency {
		return fmt.Errorf("transfer accounts must use the same currency")
	}

	transferPlace, err := db.GetStoreByName(cashDb, "transfer")
	if err != nil {
		return fmt.Errorf("special place 'transfer' does not exist: %w", err)
	}
	transactionInput := db.CreateTransactionInput{Type: "transfer_out", Amount: -amountItem.Value, AccountID: fromAccount.ID, PlaceID: &transferPlace.ID}
	transactionDate, err := getTransactionDatetime(&transactionInput, attributes, cfg)
	if err != nil {
		return err
	}
	nextIdentifier, err := getNextIdentifier(&transactionInput, cashDb, transactionDate)
	if err != nil {
		return err
	}
	toIdentifier := domain.TransactionID{Year: nextIdentifier.Year, Month: nextIdentifier.Month, Num: nextIdentifier.Num + 1}
	description := "Transfer"
	textParts := make([]string, 0)
	for _, arg := range parsed.Args {
		if text, ok := arg.(parser.ArgText); ok {
			textParts = append(textParts, text.Text)
		}
	}
	if len(textParts) > 0 {
		description = strings.Join(textParts, " ")
	}
	transactionInput.Description = description

	if !parsed.HasFlag("yes") {
		pterm.FgWhite.Printf("Transfer %s -> %s: %.2f %s\n", fromAccount.Name, toAccount.Name, amountItem.Value, fromAccount.Currency)
		confirmed, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Confirm transfer (N/y) ?").Show()
		if err != nil {
			return fmt.Errorf("error confirming transfer: %w", err)
		}
		if !confirmed {
			if isJSONOutput(parsed) {
				return renderJSONResult(output.FailureResult("transfer", output.Error{Code: "CANCELLED", Message: "transfer cancelled"}))
			}
			pterm.Warning.Println("Transfer cancelled by user")
			return nil
		}
	}

	fromID, err := db.InsertTransaction(cashDb, transactionInput)
	if err != nil {
		return err
	}
	toInput := db.CreateTransactionInput{Identifier: toIdentifier.String(), Type: "transfer_in", Amount: amountItem.Value, Description: description, Date: transactionDate, AccountID: toAccount.ID, PlaceID: &transferPlace.ID}
	toID, err := db.InsertTransaction(cashDb, toInput)
	if err != nil {
		return err
	}
	transferID, err := db.InsertTransfer(cashDb, db.CreateTransferInput{FromTransactionID: fromID, ToTransactionID: toID, FromAccountID: fromAccount.ID, ToAccountID: toAccount.ID, Amount: amountItem.Value})
	if err != nil {
		return err
	}
	if isJSONOutput(parsed) {
		return renderJSON("transfer", map[string]any{"action": "created", "id": transferID, "from": transactionInput.Identifier, "to": toInput.Identifier, "amount": amountItem.Value, "currency": fromAccount.Currency}, 1)
	}
	pterm.Success.Printf("Transfer added with id: %d\n", transferID)
	return nil
}
