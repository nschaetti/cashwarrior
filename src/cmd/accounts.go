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

func Accounts(parsed parser.ParsedCmdLine, config config.Config, cashDb *sql.DB) error {
	accounts, err := db.ListAccounts(cashDb, db.AccountListFilter{})
	if err != nil {
		return err
	}

	rows := make([][]string, 0, len(accounts))

	for _, account := range accounts {
		rows = append(rows, []string{
			strconv.FormatInt(account.ID, 10),
			account.Name,
			account.Currency,
			account.CreatedAt.Format(config.Display.DateFormat),
			account.UpdatedAt.Format(config.Display.DateFormat),
		})
	}

	theme := gui.CurrentTheme()

	t := gui.NewTable().
		WithTitle("Accounts", theme.AccountsTitleBackground).
		WithSubtitle("Configured cash accounts").
		WithHeaderBackground(theme.AccountsHeaderBackground).
		WithHeaders("ID", "Name", "Currency", "Created at", "Updated at").
		AddRows(rows)

	fmt.Println(t.Render())

	return nil
}
