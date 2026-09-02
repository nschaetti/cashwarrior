package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/output"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

type groupsSortOptions struct {
	Field string
	Desc  bool
}

func defaultGroupsSortOptions() groupsSortOptions {
	return groupsSortOptions{Field: "name", Desc: false}
}

func Groups(parsed parser.ParsedCmdLine, _ config.Config, cashDb db.DBTX) error {
	switch parsed.Subcommand {
	case "list":
		sortOptions, err := parseGroupsSortOptions(parsed)
		if err != nil {
			return err
		}
		format, err := commandOutputFormat(parsed)
		if err != nil {
			return err
		}
		if format == output.FormatJSON {
			data, err := getGroupsData(cashDb, sortOptions)
			if err != nil {
				return err
			}
			return renderJSON("groups", data, len(data.Groups))
		}
		return listGroups(cashDb, sortOptions)
	default:
		return fmt.Errorf("unknown groups subcommand %s", parsed.Subcommand)
	}
}

func listGroups(cashDb db.DBTX, sortOptions groupsSortOptions) error {
	groups, err := db.ListTransactionGroups(cashDb, db.TransactionGroupListFilter{})
	if err != nil {
		return err
	}

	transactions, err := db.ListTransactions(cashDb, []db.SQLFilter{}, []db.Filter[db.Transaction]{}, false)
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

	sort.SliceStable(groups, func(i, j int) bool {
		left := groups[i]
		right := groups[j]

		compare := 0
		switch sortOptions.Field {
		case "name":
			if left.Name < right.Name {
				compare = -1
			} else if left.Name > right.Name {
				compare = 1
			}
		case "start_date":
			compare = compareGroupDates(oldestByGroupID[left.ID], oldestByGroupID[right.ID])
		case "end_date":
			compare = compareGroupDates(newestByGroupID[left.ID], newestByGroupID[right.ID])
		}

		if compare == 0 {
			if left.Name < right.Name {
				compare = -1
			} else if left.Name > right.Name {
				compare = 1
			}
		}

		if sortOptions.Desc {
			return compare > 0
		}
		return compare < 0
	})

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

func compareGroupDates(left time.Time, right time.Time) int {
	leftSet := !left.IsZero()
	rightSet := !right.IsZero()

	if leftSet && rightSet {
		if left.Before(right) {
			return -1
		}
		if left.After(right) {
			return 1
		}
		return 0
	}

	if leftSet && !rightSet {
		return -1
	}
	if !leftSet && rightSet {
		return 1
	}

	return 0
}

func parseGroupsSortOptions(parsed parser.ParsedCmdLine) (groupsSortOptions, error) {
	sortOptions := defaultGroupsSortOptions()
	orderSpecified := false

	for _, filter := range parsed.Filters {
		attr, ok := filter.(parser.ArgAttribute)
		if !ok {
			continue
		}

		switch attr.Key {
		case "order":
			if orderSpecified {
				return sortOptions, fmt.Errorf("order specified multiple times")
			}
			if attr.Value.Raw != "name" && attr.Value.Raw != "start_date" && attr.Value.Raw != "end_date" {
				return sortOptions, fmt.Errorf("unsupported groups order field %s", attr.Value.Raw)
			}
			sortOptions.Field = attr.Value.Raw
			orderSpecified = true
		case "desc":
			desc, err := strconv.ParseBool(attr.Value.Raw)
			if err != nil {
				return sortOptions, fmt.Errorf("invalid desc value %s", attr.Value.Raw)
			}
			sortOptions.Desc = desc
		}
	}

	return sortOptions, nil
}
