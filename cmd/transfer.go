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
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/output"
	"github.com/nschaetti/cashwarrior/internal/parser"
	"github.com/pterm/pterm"
)

func Transfer(parsed parser.ParsedCmdLine, cfg config.Config, cashDb db.DBTX) error {
	switch parsed.Subcommand {
	case "add", "", "default":
		return transferAdd(parsed, cfg, cashDb)
	case "list":
		return listTransfers(parsed, cashDb)
	case "delete":
		return deleteTransfer(parsed, cashDb)
	default:
		return fmt.Errorf("unknown transfer subcommand: %s", parsed.Subcommand)
	}
}

func transferAdd(parsed parser.ParsedCmdLine, cfg config.Config, cashDb db.DBTX) error {
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

func listTransfers(parsed parser.ParsedCmdLine, cashDb db.DBTX) error {
	transfers, err := db.ListTransfers(cashDb, db.TransferListFilter{})
	if err != nil {
		return err
	}

	items := make([]output.TransferListItem, 0, len(transfers))
	rows := make([][]string, 0, len(transfers))
	for _, transfer := range transfers {
		fromTx, err := transfer.GetFromTransaction(cashDb)
		if err != nil {
			return err
		}
		toTx, err := transfer.GetToTransaction(cashDb)
		if err != nil {
			return err
		}
		fromAccount, err := transfer.GetFromAccount(cashDb)
		if err != nil {
			return err
		}
		toAccount, err := transfer.GetToAccount(cashDb)
		if err != nil {
			return err
		}
		items = append(items, output.TransferListItem{
			ID:          transfer.ID,
			FromAccount: fromAccount.Name,
			ToAccount:   toAccount.Name,
			FromID:      fromTx.Identifier,
			ToID:        toTx.Identifier,
			Amount:      transfer.Amount,
			Currency:    fromAccount.Currency,
			Date:        fromTx.Datetime,
		})
		rows = append(rows, []string{
			strconv.FormatInt(transfer.ID, 10),
			fromAccount.Name,
			toAccount.Name,
			strconv.FormatFloat(transfer.Amount, 'f', 2, 64),
			fromAccount.Currency,
			fromTx.Identifier,
			toTx.Identifier,
			fromTx.Datetime.Format("2006-01-02"),
		})
	}

	format, err := commandOutputFormat(parsed)
	if err != nil {
		return err
	}
	if format == output.FormatJSON {
		return renderJSON("transfers", output.TransfersData{Transfers: items}, len(items))
	}

	theme := gui.CurrentTheme()
	t := gui.NewTable().
		WithTitle("Transfers", theme.TransactionListTitleBackground).
		WithSubtitle("Configured transfers").
		WithHeaderBackground(theme.TransactionListHeaderBackground).
		WithHeaders("ID", "From", "To", "Amount", "Currency", "From ID", "To ID", "Date").
		AddRows(rows)

	fmt.Println(t.Render())
	fmt.Println()
	fmt.Println()
	return nil
}

func deleteTransfer(parsed parser.ParsedCmdLine, cashDb db.DBTX) error {
	if err := requireYesForJSON(parsed); err != nil {
		return err
	}

	identifier, err := getTransferDeleteIdentifier(parsed)
	if err != nil {
		return err
	}

	transaction, err := db.GetTransactionByIdentifier(cashDb, identifier)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("transaction %s does not exist", identifier)
	}
	if err != nil {
		return err
	}

	transfer, err := db.GetTransferByTransactionID(cashDb, transaction.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("transaction %s is not a transfer", identifier)
	}
	if err != nil {
		return err
	}

	if !parsed.HasFlag("yes") {
		fromAccount, err := transfer.GetFromAccount(cashDb)
		if err != nil {
			return err
		}
		toAccount, err := transfer.GetToAccount(cashDb)
		if err != nil {
			return err
		}
		pterm.FgWhite.Printf("Delete transfer %s -> %s: %.2f %s\n", fromAccount.Name, toAccount.Name, transfer.Amount, fromAccount.Currency)
		confirmed, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Confirm transfer deletion (N/y) ?").Show()
		if err != nil {
			return fmt.Errorf("error confirming transfer deletion: %w", err)
		}
		if !confirmed {
			if isJSONOutput(parsed) {
				return renderJSONResult(output.FailureResult("transfer", output.Error{Code: "CANCELLED", Message: "transfer deletion cancelled"}))
			}
			pterm.Warning.Println("Transfer deletion cancelled by user")
			return nil
		}
	}

	if err := db.UpdateTransferDeleted(cashDb, transfer.ID, true); err != nil {
		return err
	}
	if err := db.UpdateTransactionDeleted(cashDb, transfer.FromTransactionID, true); err != nil {
		return err
	}
	if transfer.ToTransactionID != transfer.FromTransactionID {
		if err := db.UpdateTransactionDeleted(cashDb, transfer.ToTransactionID, true); err != nil {
			return err
		}
	}

	if isJSONOutput(parsed) {
		return renderJSON("transfer", map[string]any{"action": "deleted", "identifier": identifier}, 1)
	}
	fmt.Printf("Transfer %s deleted\n", identifier)
	return nil
}

func getTransferDeleteIdentifier(parsed parser.ParsedCmdLine) (string, error) {
	if len(parsed.Args) != 1 {
		return "", fmt.Errorf("transfer delete requires an identifier")
	}
	attr, ok := parsed.Args[0].(parser.ArgAttribute)
	if !ok || attr.Key != "identifier" {
		return "", fmt.Errorf("transfer delete requires an identifier")
	}
	value := strings.TrimSpace(attr.Value.Raw)
	if value == "" {
		return "", fmt.Errorf("transfer delete requires an identifier")
	}
	return value, nil
}
