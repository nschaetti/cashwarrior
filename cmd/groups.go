package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Groups(parsed parser.ParsedCmdLine, _ config.Config, cashDb db.DBTX) error {
	switch parsed.Subcommand {
	case "list":
		return listGroups(cashDb)
	default:
		return fmt.Errorf("unknown groups subcommand %s", parsed.Subcommand)
	}
}

func listGroups(cashDb db.DBTX) error {
	groups, err := db.ListTransactionGroups(cashDb, db.TransactionGroupListFilter{})
	if err != nil {
		return err
	}

	transactions, err := db.ListTransactions(cashDb, nil, nil)
	if err != nil {
		return err
	}

	totalByGroupID := make(map[int64]float64)
	countByGroupID := make(map[int64]int)
	oldestByGroupID := make(map[int64]time.Time)
	newestByGroupID := make(map[int64]time.Time)
	for _, transaction := range transactions {
		if transaction.GroupID == nil {
			continue
		}
		groupID := *transaction.GroupID
		totalByGroupID[groupID] += transaction.Amount
		countByGroupID[groupID]++

		if currentOldest, ok := oldestByGroupID[groupID]; !ok || transaction.Datetime.Before(currentOldest) {
			oldestByGroupID[groupID] = transaction.Datetime
		}
		if currentNewest, ok := newestByGroupID[groupID]; !ok || transaction.Datetime.After(currentNewest) {
			newestByGroupID[groupID] = transaction.Datetime
		}
	}

	rows := make([][]string, 0, len(groups))
	for _, group := range groups {
		startDate := "-"
		if dt, ok := oldestByGroupID[group.ID]; ok {
			startDate = dt.Format("2006-01-02")
		}
		endDate := "-"
		if dt, ok := newestByGroupID[group.ID]; ok {
			endDate = dt.Format("2006-01-02")
		}

		rows = append(rows, []string{
			strconv.FormatInt(group.ID, 10),
			group.Name,
			strconv.Itoa(countByGroupID[group.ID]),
			startDate,
			endDate,
			fmt.Sprintf("%+.2f", totalByGroupID[group.ID]),
		})
	}

	theme := gui.CurrentTheme()
	t := gui.NewTable().
		WithTitle("Groups", theme.CategoriesTitleBackground).
		WithSubtitle("Configured transaction groups").
		WithHeaderBackground(theme.CategoriesHeaderBackground).
		WithHeaders("ID", "Name", "Transactions", "Start Date", "End Date", "Transactions Sum").
		AddRows(rows)

	fmt.Println(t.Render())
	fmt.Println()
	fmt.Println()
	return nil
}
