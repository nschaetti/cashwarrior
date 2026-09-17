package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/output"
	"github.com/nschaetti/cashwarrior/internal/parser"
	"github.com/nschaetti/cashwarrior/internal/utils"
	"github.com/pterm/pterm"
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
	case "list", "ls":
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
	case "add":
		return addGroups(parsed, cashDb)
	case "modify", "rename", "rn":
		return modifyGroups(parsed, cashDb)
	case "delete", "rm":
		return deleteGroups(parsed, cashDb)
	case "remove":
		return removeFromGroup(parsed, cashDb)
	default:
		return fmt.Errorf("unknown groups subcommand %s", parsed.Subcommand)
	}
}

func getGroupNameArg(arg parser.Arg) (string, error) {
	text, ok := arg.(parser.ArgText)
	if ok {
		if text.Text == "" {
			return "", fmt.Errorf("group name cannot be empty")
		}
		return text.Text, nil
	}
	attr, ok := arg.(parser.ArgAttribute)
	if ok && attr.Key == "group" && !attr.Value.IsEmpty() {
		return attr.Value.Raw, nil
	}
	return "", fmt.Errorf("group name is required")
}

func parseGroupAddArgs(parsed parser.ParsedCmdLine) ([]string, string, error) {
	transactionRefs := make([]string, 0, len(parsed.Args))
	groupName := ""

	for _, arg := range parsed.Args {
		switch token := arg.(type) {
		case parser.ArgText:
			if groupName != "" {
				return nil, "", fmt.Errorf("multiple groups given")
			}
			groupName = token.Text
		case parser.ArgAttribute:
			switch token.Key {
			case "group":
				if groupName != "" {
					return nil, "", fmt.Errorf("multiple groups given")
				}
				groupName = token.Value.Raw
			case "identifier":
				transactionRefs = append(transactionRefs, token.Value.Raw)
			}
		}
	}

	if len(transactionRefs) == 0 {
		return nil, "", fmt.Errorf("no transaction given")
	}
	if groupName == "" {
		return nil, "", fmt.Errorf("no group given")
	}

	return transactionRefs, groupName, nil
}

func addGroups(parsed parser.ParsedCmdLine, cashDb db.DBTX) error {
	if err := requireYesForJSON(parsed); err != nil {
		return err
	}
	transactionRefs, groupName, err := parseGroupAddArgs(parsed)
	if err != nil {
		return err
	}
	return confirmAndLinkTransactions(parsed, cashDb, groupName, transactionRefs)
}

func modifyGroups(parsed parser.ParsedCmdLine, cashDb db.DBTX) error {
	name, err := getGroupNameArg(parsed.Args[0])
	if err != nil {
		return err
	}
	group, err := db.GetGroupByName(cashDb, name)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("group %s does not exist", name)
	}
	if err != nil {
		return err
	}
	newName := ""
	for _, arg := range parsed.Args[1:] {
		attr, ok := arg.(parser.ArgAttribute)
		if ok && attr.Key == "group" {
			newName = attr.Value.Raw
		}
	}
	if newName == "" {
		return fmt.Errorf("new group name cannot be empty")
	}
	if newName != group.Name {
		exists, err := db.TransactionGroupExists(cashDb, newName)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("group %s already exists", newName)
		}
		if err := db.UpdateTransactionGroupName(cashDb, group.ID, newName); err != nil {
			return err
		}
	}
	if isJSONOutput(parsed) {
		return renderJSON("group", map[string]any{"action": "updated", "name": name}, 1)
	}
	fmt.Printf("Group %s updated\n", name)
	return nil
}

func deleteGroups(parsed parser.ParsedCmdLine, cashDb db.DBTX) error {
	if err := requireYesForJSON(parsed); err != nil {
		return err
	}
	name, err := getGroupNameArg(parsed.Args[0])
	if err != nil {
		return err
	}
	group, err := db.GetGroupByName(cashDb, name)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("group %s does not exist", name)
	}
	if err != nil {
		return err
	}
	count, err := db.CountTransactionsByGroupID(cashDb, group.ID)
	if err != nil {
		return err
	}
	if count > 0 && !isJSONOutput(parsed) {
		fmt.Printf("Warning: group %s has %d linked transactions, they will be detached\n", name, count)
	}
	if !parsed.HasFlag("yes") && !utils.AskYesNo(fmt.Sprintf("Delete group %s?", name)) {
		return nil
	}
	if err := db.ClearTransactionsGroupID(cashDb, group.ID); err != nil {
		return err
	}
	if err := db.DeleteTransactionGroupByID(cashDb, group.ID); err != nil {
		return err
	}
	if isJSONOutput(parsed) {
		return renderJSON("group", map[string]any{"action": "deleted", "name": name}, 1)
	}
	fmt.Printf("Group %s deleted\n", name)
	return nil
}

func removeFromGroup(parsed parser.ParsedCmdLine, cashDb db.DBTX) error {
	if err := requireYesForJSON(parsed); err != nil {
		return err
	}
	transactionRef, groupName, err := parseGroupRemoveArgs(parsed)
	if err != nil {
		return err
	}

	if !isJSONOutput(parsed) {
		pterm.FgWhite.Println("Transaction to be removed:")
		pterm.FgWhite.Println("==========================")
		pterm.FgWhite.Println("Group: ", groupName, "")
		pterm.FgWhite.Println("Transaction: ", transactionRef, "")
	}

	ok := parsed.HasFlag("yes")
	if !ok {
		ok, err = pterm.DefaultInteractiveConfirm.
			WithDefaultText("Confirm removal (N/y) ?").
			Show()
		if err != nil {
			panic(fmt.Errorf("error confirming removal: %w", err))
		}
	}

	if !ok {
		if isJSONOutput(parsed) {
			return renderJSONResult(output.FailureResult("group", output.Error{Code: "CANCELLED", Message: "removal cancelled"}))
		}
		pterm.Warning.Println("Aborted removal")
		return nil
	}

	group, err := db.GetGroupByName(cashDb, groupName)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("group %s does not exist", groupName)
	}
	if err != nil {
		return err
	}
	transaction, err := getTransactionByReference(cashDb, transactionRef)
	if err != nil {
		return err
	}
	if transaction.GroupID == nil || *transaction.GroupID != group.ID {
		return fmt.Errorf("transaction %s is not in group %s", transactionRef, groupName)
	}
	if err := db.UpdateTransactionGroupID(cashDb, transaction.ID, nil); err != nil {
		return err
	}

	if isJSONOutput(parsed) {
		return renderJSON("group", map[string]any{"action": "removed", "group": groupName, "identifier": transactionRef}, 1)
	}
	fmt.Printf("Removed transaction %s from group %s\n", transactionRef, groupName)
	return nil
}

func parseGroupRemoveArgs(parsed parser.ParsedCmdLine) (string, string, error) {
	transactionRef := ""
	groupName := ""
	for _, arg := range parsed.Args {
		attr, ok := arg.(parser.ArgAttribute)
		if !ok {
			continue
		}
		switch attr.Key {
		case "group":
			groupName = attr.Value.Raw
		case "identifier":
			transactionRef = attr.Value.Raw
		}
	}
	if transactionRef == "" {
		return "", "", fmt.Errorf("no transaction given")
	}
	if groupName == "" {
		return "", "", fmt.Errorf("no group given")
	}
	return transactionRef, groupName, nil
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
