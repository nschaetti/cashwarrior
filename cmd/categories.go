package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/output"
	"github.com/nschaetti/cashwarrior/internal/parser"
	"github.com/nschaetti/cashwarrior/internal/utils"
)

func Categories(parsed parser.ParsedCmdLine, config config.Config, cashDb db.DBTX) error {
	switch parsed.Subcommand {
	case "list":
		format, err := commandOutputFormat(parsed)
		if err != nil {
			return err
		}
		if format == output.FormatJSON {
			data, err := getCategoriesData(cashDb)
			if err != nil {
				return err
			}
			return renderJSON("categories", data, len(data.Categories))
		}
		return listCategories(config, cashDb)
	case "add":
		return addCategory(parsed, cashDb)
	case "modify":
		return modifyCategory(parsed, cashDb)
	case "delete":
		return deleteCategory(parsed, cashDb)
	default:
		return fmt.Errorf("unknown categories subcommand %s", parsed.Subcommand)
	}
}

func listCategories(config config.Config, cashDb db.DBTX) error {
	categories, err := db.ListCategories(cashDb, db.CategoryListFilter{})
	if err != nil {
		return err
	}

	transactions, err := db.ListTransactions(cashDb, []db.SQLFilter{}, []db.Filter[db.Transaction]{}, false)
	if err != nil {
		return err
	}

	type categoryStats struct {
		Transactions int
		Expenses     float64
		Incomes      float64
	}

	statsByCategoryID := make(map[int64]*categoryStats)
	for _, category := range categories {
		if category.Name == "root" {
			continue
		}
		statsByCategoryID[category.ID] = &categoryStats{}
	}

	for _, transaction := range transactions {
		if transaction.CategoryID == nil {
			continue
		}
		stats, ok := statsByCategoryID[*transaction.CategoryID]
		if !ok {
			continue
		}
		switch transaction.Type {
		case "income":
			stats.Transactions++
			stats.Incomes += transaction.Amount
		case "expense":
			stats.Transactions++
			stats.Expenses += transaction.Amount
		}
	}

	rootID := int64(0)
	hasRoot := false
	for _, category := range categories {
		if category.Name == "root" {
			rootID = category.ID
			hasRoot = true
			break
		}
	}

	childrenByParentID := make(map[int64][]db.Category)
	topLevel := make([]db.Category, 0)
	for _, category := range categories {
		if category.Name == "root" {
			continue
		}

		if category.ParentID == nil {
			topLevel = append(topLevel, category)
			continue
		}

		if hasRoot && *category.ParentID == rootID {
			topLevel = append(topLevel, category)
			continue
		}

		childrenByParentID[*category.ParentID] = append(childrenByParentID[*category.ParentID], category)
	}

	sort.Slice(topLevel, func(i, j int) bool {
		return topLevel[i].Name < topLevel[j].Name
	})
	for parentID := range childrenByParentID {
		children := childrenByParentID[parentID]
		sort.Slice(children, func(i, j int) bool {
			return children[i].Name < children[j].Name
		})
		childrenByParentID[parentID] = children
	}

	aggregateStats := make(map[int64]*categoryStats)
	visiting := make(map[int64]bool)
	visited := make(map[int64]bool)

	var rollupCategoryStats func(categoryID int64) *categoryStats
	rollupCategoryStats = func(categoryID int64) *categoryStats {
		if visited[categoryID] {
			return aggregateStats[categoryID]
		}
		if visiting[categoryID] {
			base := statsByCategoryID[categoryID]
			return &categoryStats{
				Transactions: base.Transactions,
				Expenses:     base.Expenses,
				Incomes:      base.Incomes,
			}
		}

		visiting[categoryID] = true
		base := statsByCategoryID[categoryID]
		total := &categoryStats{
			Transactions: base.Transactions,
			Expenses:     base.Expenses,
			Incomes:      base.Incomes,
		}

		children := childrenByParentID[categoryID]
		for _, child := range children {
			childTotals := rollupCategoryStats(child.ID)
			total.Transactions += childTotals.Transactions
			total.Expenses += childTotals.Expenses
			total.Incomes += childTotals.Incomes
		}

		visiting[categoryID] = false
		visited[categoryID] = true
		aggregateStats[categoryID] = total
		return total
	}

	for categoryID := range statsByCategoryID {
		rollupCategoryStats(categoryID)
	}

	formatAmount := func(amount float64, forceSign bool) string {
		if forceSign {
			return fmt.Sprintf("%+.2f", amount)
		}
		if amount == 0 {
			return "0.00"
		}
		return fmt.Sprintf("%.2f", amount)
	}

	rows := make([][]string, 0, len(categories))

	var appendCategoryRows func(items []db.Category, depth int)
	appendCategoryRows = func(items []db.Category, depth int) {
		for _, category := range items {
			stats := aggregateStats[category.ID]
			net := stats.Incomes + stats.Expenses
			rows = append(rows, []string{
				strconv.FormatInt(category.ID, 10),
				strings.Repeat("   ", depth) + category.Name,
				strconv.Itoa(stats.Transactions),
				formatAmount(stats.Expenses, false),
				formatAmount(stats.Incomes, true),
				formatAmount(net, true),
			})

			children, ok := childrenByParentID[category.ID]
			if !ok {
				continue
			}
			appendCategoryRows(children, depth+1)
		}
	}

	appendCategoryRows(topLevel, 0)

	theme := gui.CurrentTheme()

	t := gui.NewTable().
		WithTitle("Categories", theme.CategoriesTitleBackground).
		WithSubtitle("Configured categories").
		WithHeaderBackground(theme.CategoriesHeaderBackground).
		WithHeaders("ID", "Name", "Transactions", "Expenses", "Incomes", "Net").
		AddRows(rows)

	fmt.Println(t.Render())
	fmt.Println()
	fmt.Println()

	return nil
}

func getCategoryNameArg(arg parser.Arg) (string, error) {
	text, ok := arg.(parser.ArgText)
	if ok {
		if text.Text == "" {
			return "", fmt.Errorf("category name cannot be empty")
		}
		return text.Text, nil
	}
	attr, ok := arg.(parser.ArgAttribute)
	if ok && attr.Key == "category" && !attr.Value.IsEmpty() {
		return attr.Value.Raw, nil
	}
	return "", fmt.Errorf("category name is required")
}

func addCategory(parsed parser.ParsedCmdLine, cashDb db.DBTX) error {
	name, err := getCategoryNameArg(parsed.Args[0])
	if err != nil {
		return err
	}
	exists, err := db.CategoryExists(cashDb, name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("category %s already exists", name)
	}
	attributes := getAttributesFromTokens(parsed.Args)
	var parentID *int64
	if parentName, ok := attributes["parent"]; ok {
		parent, err := db.GetCategoryByName(cashDb, parentName)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("parent category %s does not exist", parentName)
		}
		if err != nil {
			return err
		}
		parentID = &parent.ID
	}
	_, err = db.InsertCategory(cashDb, db.CreateCategoryInput{Name: name, ParentID: parentID})
	if err != nil {
		return err
	}
	if isJSONOutput(parsed) {
		return renderJSON("category", map[string]any{"action": "created", "name": name}, 1)
	}
	fmt.Printf("Category %s created\n", name)
	return nil
}

func modifyCategory(parsed parser.ParsedCmdLine, cashDb db.DBTX) error {
	name, err := getCategoryNameArg(parsed.Args[0])
	if err != nil {
		return err
	}
	category, err := db.GetCategoryByName(cashDb, name)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("category %s does not exist", name)
	}
	if err != nil {
		return err
	}
	if category.Name == db.RootCategoryName() {
		return fmt.Errorf("cannot modify root category")
	}
	attributes := getAttributesFromTokens(parsed.Args[1:])
	if newName, ok := attributes["category"]; ok {
		newName = strings.TrimSpace(newName)
		if newName == "" {
			return fmt.Errorf("category name cannot be empty")
		}
		if newName != category.Name {
			exists, err := db.CategoryExists(cashDb, newName)
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("category %s already exists", newName)
			}
			if err := db.UpdateCategoryName(cashDb, category.ID, newName); err != nil {
				return err
			}
		}
	}
	if parentName, ok := attributes["parent"]; ok {
		parentName = strings.TrimSpace(parentName)
		if parentName == "" {
			return fmt.Errorf("parent category cannot be empty")
		}
		parent, err := db.GetCategoryByName(cashDb, parentName)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("parent category %s does not exist", parentName)
		}
		if err != nil {
			return err
		}
		if parent.ID == category.ID {
			return fmt.Errorf("category cannot be its own parent")
		}
		if err := db.UpdateCategoryParentID(cashDb, category.ID, &parent.ID); err != nil {
			return err
		}
	}
	if isJSONOutput(parsed) {
		return renderJSON("category", map[string]any{"action": "updated", "name": name}, 1)
	}
	fmt.Printf("Category %s updated\n", name)
	return nil
}

func deleteCategory(parsed parser.ParsedCmdLine, cashDb db.DBTX) error {
	if err := requireYesForJSON(parsed); err != nil {
		return err
	}
	name, err := getCategoryNameArg(parsed.Args[0])
	if err != nil {
		return err
	}
	category, err := db.GetCategoryByName(cashDb, name)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("category %s does not exist", name)
	}
	if err != nil {
		return err
	}
	if category.Name == db.RootCategoryName() {
		return fmt.Errorf("cannot delete root category")
	}
	childCount, err := db.CountChildCategories(cashDb, category.ID)
	if err != nil {
		return err
	}
	if childCount > 0 {
		return fmt.Errorf("category %s has child categories", name)
	}
	txCount, err := db.CountTransactionsByCategoryID(cashDb, category.ID)
	if err != nil {
		return err
	}
	if txCount > 0 {
		return fmt.Errorf("category %s has linked transactions", name)
	}
	if !parsed.HasFlag("yes") && !utils.AskYesNo(fmt.Sprintf("Delete category %s?", name)) {
		return nil
	}
	if err := db.DeleteCategoryByID(cashDb, category.ID); err != nil {
		return err
	}
	if isJSONOutput(parsed) {
		return renderJSON("category", map[string]any{"action": "deleted", "name": name}, 1)
	}
	fmt.Printf("Category %s deleted\n", name)
	return nil
}
