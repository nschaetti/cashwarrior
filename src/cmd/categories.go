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

func Categories(parsed parser.ParsedCmdLine, config config.Config, cashDb *sql.DB) error {
	categories, err := db.ListCategories(cashDb, db.CategoryListFilter{})
	if err != nil {
		return err
	}

	rows := make([][]string, 0, len(categories))

	for _, category := range categories {
		rows = append(rows, []string{
			strconv.FormatInt(category.ID, 10),
			category.Name,
			category.CreatedAt.Format(config.Display.DateFormat),
			category.UpdatedAt.Format(config.Display.DateFormat),
		})
	}

	theme := gui.CurrentTheme()

	t := gui.NewTable().
		WithTitle("Categories", theme.CategoriesTitleBackground).
		WithSubtitle("Configured categories").
		WithHeaderBackground(theme.CategoriesHeaderBackground).
		WithHeaders("ID", "Name", "Created at", "Updated at").
		AddRows(rows)

	fmt.Println(t.Render())

	return nil
}
