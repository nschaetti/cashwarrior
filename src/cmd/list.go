package cmd

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func List(parsed parser.ParsedCmdLine, config config.Config, cashDb *sql.DB) error {
	transactions, err := db.ListTransactions(cashDb, []db.Filter{})
	if err != nil {
		return err
	}

	rows := make([][]string, 0, len(transactions))

	for _, transaction := range transactions {
		account, err := db.GetAccountByID(cashDb, *transaction.AccountID)
		if err != nil {
			return err
		}
		var categoryName string
		if transaction.CategoryID != nil {
			category, err := db.GetCategoryByID(cashDb, *transaction.CategoryID)
			if err != nil {
				return err
			}
			categoryName = category.Name
		} else {
			categoryName = "none"
		}
		var vendorName string
		vendor, err := db.GetPlaceByID(cashDb, *transaction.PlaceID)
		if err != nil {
			return err
		}
		vendorName = vendor.Name
		rows = append(rows, []string{
			transaction.Identifier,
			transaction.Type,
			strconv.FormatFloat(transaction.Amount, 'f', 2, 64),
			vendorName,
			transaction.Description,
			transaction.Datetime.Format(config.Display.DateFormat),
			account.Name,
			categoryName,
			transaction.CreatedAt.Format(config.Display.DateFormat),
			transaction.UpdatedAt.Format(config.Display.DateFormat),
		})
	}

	theme := gui.CurrentTheme()

	t := gui.NewTable().
		WithTitle("Transactions", theme.TransactionListTitleBackground).
		WithSubtitle("Configured transactions").
		WithHeaderBackground(theme.TransactionListHeaderBackground).
		WithHeaders("ID", "Type", "Amount", "Vendor", "Description", "Datetime", "Account", "Category", "Created at", "Updated at").
		AddRows(rows)

	fmt.Println(t.Render())

	return nil
}
