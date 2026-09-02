package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/output"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func formatShowDatetime(datetime time.Time, cfg config.Config) string {
	layout := strings.Split(cfg.Display.DateFormat, " ")[0]
	return datetime.Format(layout)
}

func showValue(value string) string {
	if value == "" {
		return "none"
	}
	return value
}

func getShowData(query db.DBTX, identifier string) (output.ShowTransaction, error) {
	transaction, err := db.GetTransactionByIdentifier(query, identifier)
	if err != nil {
		return output.ShowTransaction{}, err
	}
	place, err := transaction.GetPlace(query)
	if err != nil {
		return output.ShowTransaction{}, err
	}
	category, err := transaction.GetCategory(query)
	if err != nil {
		return output.ShowTransaction{}, err
	}
	group, err := transaction.GetGroup(query)
	if err != nil {
		return output.ShowTransaction{}, err
	}
	tags, err := db.ListTagsByTransactionID(query, transaction.ID)
	if err != nil {
		return output.ShowTransaction{}, err
	}

	data := output.ShowTransaction{ID: transaction.Identifier, Type: transaction.Type, Amount: transaction.Amount,
		Currency: transaction.Currency, Account: transaction.AccountName, Description: transaction.Description,
		Date: transaction.Datetime, Deleted: transaction.Deleted, CreatedAt: transaction.CreatedAt, UpdatedAt: transaction.UpdatedAt,
		Tags: make([]string, 0, len(tags))}
	if place != nil {
		data.Place = place.Name
	}
	if category != nil {
		data.Category = category.Name
	}
	if group != nil {
		data.Group = group.Name
	}
	for _, tag := range tags {
		data.Tags = append(data.Tags, tag.Name)
	}

	transfer, err := db.GetTransferByTransactionID(query, transaction.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return output.ShowTransaction{}, err
	}
	if err == nil {
		from, err := transfer.GetFromAccount(query)
		if err != nil {
			return output.ShowTransaction{}, err
		}
		to, err := transfer.GetToAccount(query)
		if err != nil {
			return output.ShowTransaction{}, err
		}
		pairID := transfer.FromTransactionID
		if pairID == transaction.ID {
			pairID = transfer.ToTransactionID
		}
		pair, err := db.GetTransactionByID(query, pairID)
		if err != nil {
			return output.ShowTransaction{}, err
		}
		data.Transfer = &output.ShowTransfer{Amount: transfer.Amount, FromAccount: from.Name, ToAccount: to.Name, PairID: pair.Identifier}
	}
	return data, nil
}

func Show(parsed parser.ParsedCmdLine, cfg config.Config, query db.DBTX) error {
	format, err := commandOutputFormat(parsed)
	if err != nil {
		return err
	}
	if format == output.FormatJSON {
		data, err := getShowData(query, parsed.Args[0].RawString())
		if err != nil {
			return err
		}
		return renderJSON("transaction", data, 1)
	}
	transaction, err := db.GetTransactionByIdentifier(query, parsed.Args[0].RawString())
	if err != nil {
		return err
	}

	place, err := transaction.GetPlace(query)
	if err != nil {
		return err
	}

	category, err := transaction.GetCategory(query)
	if err != nil {
		return err
	}

	group, err := transaction.GetGroup(query)
	if err != nil {
		return err
	}

	tags, err := db.ListTagsByTransactionID(query, transaction.ID)
	if err != nil {
		return err
	}

	rows := [][]string{
		{"ID", transaction.Identifier},
		{"Type", transaction.Type},
		{"Amount", strconv.FormatFloat(transaction.Amount, 'f', 2, 64)},
		{"Currency", showValue(transaction.Currency)},
		{"Account", showValue(transaction.AccountName)},
		{"Place", "none"},
		{"Description", showValue(transaction.Description)},
		{"Date", formatShowDatetime(transaction.Datetime, cfg)},
		{"Category", "none"},
		{"Group", "none"},
		{"Tags", "none"},
		{"Deleted", strconv.FormatBool(transaction.Deleted)},
		{"Created at", formatShowDatetime(transaction.CreatedAt, cfg)},
		{"Updated at", formatShowDatetime(transaction.UpdatedAt, cfg)},
	}

	if place != nil {
		rows[5][1] = showValue(place.Name)
	}
	if category != nil {
		rows[8][1] = showValue(category.Name)
	}
	if group != nil {
		rows[9][1] = showValue(group.Name)
	}
	if len(tags) > 0 {
		tagNames := make([]string, 0, len(tags))
		for _, tag := range tags {
			tagNames = append(tagNames, tag.Name)
		}
		rows[10][1] = strings.Join(tagNames, ", ")
	}

	transfer, err := db.GetTransferByTransactionID(query, transaction.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if err == nil {
		fromAccount, err := transfer.GetFromAccount(query)
		if err != nil {
			return err
		}
		toAccount, err := transfer.GetToAccount(query)
		if err != nil {
			return err
		}

		pairTransactionID := transfer.FromTransactionID
		if pairTransactionID == transaction.ID {
			pairTransactionID = transfer.ToTransactionID
		} else {
			pairTransactionID = transfer.FromTransactionID
		}

		pairTransaction, err := db.GetTransactionByID(query, pairTransactionID)
		if err != nil {
			return err
		}

		rows = append(rows,
			[]string{"Transfer amount", strconv.FormatFloat(transfer.Amount, 'f', 2, 64)},
			[]string{"Transfer from account", showValue(fromAccount.Name)},
			[]string{"Transfer to account", showValue(toAccount.Name)},
			[]string{"Transfer pair ID", pairTransaction.Identifier},
		)
	}

	theme := gui.CurrentTheme()
	t := gui.NewTable().
		WithTitle("Transaction", theme.TransactionListTitleBackground).
		WithSubtitle(transaction.Identifier).
		AddRows(rows)

	fmt.Printf("\n%s\n\n", t.Render())
	return nil
}
