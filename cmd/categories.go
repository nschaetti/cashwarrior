package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Categories(parsed parser.ParsedCmdLine, config config.Config, cashDb db.DBTX) error {
	categories, err := db.ListCategories(cashDb, db.CategoryListFilter{})
	if err != nil {
		return err
	}

	transactions, err := db.ListTransactions(cashDb, []db.SQLFilter{}, []db.Filter[db.Transaction]{})
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
