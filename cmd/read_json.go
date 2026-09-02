package cmd

import (
	"sort"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/output"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

type jsonGroupAggregate struct {
	count      int
	sum        float64
	start, end time.Time
}

func getAccountListData(cashDb db.DBTX) (struct {
	Accounts []output.AccountListItem `json:"accounts"`
}, error) {
	result := struct {
		Accounts []output.AccountListItem `json:"accounts"`
	}{Accounts: make([]output.AccountListItem, 0)}
	accounts, err := db.ListAccounts(cashDb, []db.SQLFilter{}, []db.Filter[db.Account]{}, []string{})
	if err != nil {
		return result, err
	}
	transactions, err := db.ListTransactions(cashDb, []db.SQLFilter{}, []db.Filter[db.Transaction]{}, false)
	if err != nil {
		return result, err
	}
	type totals struct {
		balance, income, expenses, transfers float64
		operations                           int
	}
	byID := make(map[int64]*totals, len(accounts))
	for _, account := range accounts {
		byID[account.ID] = &totals{balance: account.InitialBalance}
	}
	year, month, _ := time.Now().Date()
	for _, tx := range transactions {
		if tx.AccountID == nil || byID[*tx.AccountID] == nil {
			continue
		}
		t := byID[*tx.AccountID]
		t.balance += tx.Amount
		t.operations++
		ty, tm, _ := tx.Datetime.Date()
		if ty != year || tm != month {
			continue
		}
		switch tx.Type {
		case "income":
			t.income += tx.Amount
		case "expense":
			t.expenses += tx.Amount
		case "transfer_in", "transfer_out":
			t.transfers += tx.Amount
		}
	}
	for _, account := range accounts {
		t := byID[account.ID]
		result.Accounts = append(result.Accounts, output.AccountListItem{ID: account.ID, Name: account.Name, Currency: account.Currency, InitialBalance: account.InitialBalance, Balance: t.balance, Operations: t.operations, MonthIncome: t.income, MonthExpenses: t.expenses, MonthNet: t.income + t.expenses, MonthTransfers: t.transfers})
	}
	return result, nil
}

func getAccountBalanceData(cashDb db.DBTX, account db.Account, transactions []db.Transaction, balances map[int64]float64) (output.AccountBalanceData, error) {
	data := output.AccountBalanceData{Account: account.Name, Currency: account.Currency, InitialBalance: account.InitialBalance, Transactions: make([]output.AccountBalanceItem, 0, len(transactions))}
	for _, tx := range transactions {
		vendor := ""
		place, err := tx.GetPlace(cashDb)
		if err != nil {
			return output.AccountBalanceData{}, err
		}
		if place != nil {
			vendor = place.Name
		}
		category := ""
		if tx.CategoryID != nil {
			value, err := tx.GetCategory(cashDb)
			if err != nil {
				return output.AccountBalanceData{}, err
			}
			if value != nil {
				category = value.Name
			}
		}
		data.Transactions = append(data.Transactions, output.AccountBalanceItem{ID: tx.ID, Identifier: tx.Identifier, Type: tx.Type, Amount: tx.Amount, Currency: account.Currency, Vendor: vendor, Description: tx.Description, Date: tx.Datetime, Category: category, Balance: balances[tx.ID]})
	}
	return data, nil
}

func getGroupsData(cashDb db.DBTX, options groupsSortOptions) (output.GroupsData, error) {
	groups, err := db.ListTransactionGroups(cashDb, db.TransactionGroupListFilter{})
	if err != nil {
		return output.GroupsData{}, err
	}
	transactions, err := db.ListTransactions(cashDb, []db.SQLFilter{}, []db.Filter[db.Transaction]{}, false)
	if err != nil {
		return output.GroupsData{}, err
	}
	aggregates := make(map[int64]*jsonGroupAggregate, len(groups))
	for _, group := range groups {
		aggregates[group.ID] = &jsonGroupAggregate{}
	}
	for _, tx := range transactions {
		if tx.GroupID == nil || aggregates[*tx.GroupID] == nil {
			continue
		}
		a := aggregates[*tx.GroupID]
		a.count++
		a.sum += tx.Amount
		if a.start.IsZero() || tx.Datetime.Before(a.start) {
			a.start = tx.Datetime
		}
		if a.end.IsZero() || tx.Datetime.After(a.end) {
			a.end = tx.Datetime
		}
	}
	sort.SliceStable(groups, func(i, j int) bool { return compareGroupForJSON(groups[i], groups[j], aggregates, options) })
	data := output.GroupsData{Groups: make([]output.GroupListItem, 0, len(groups))}
	for _, group := range groups {
		a := aggregates[group.ID]
		item := output.GroupListItem{ID: group.ID, Name: group.Name, Transactions: a.count, Sum: a.sum}
		if !a.start.IsZero() {
			v := a.start
			item.StartDate = &v
		}
		if !a.end.IsZero() {
			v := a.end
			item.EndDate = &v
		}
		data.Groups = append(data.Groups, item)
	}
	return data, nil
}

func compareGroupForJSON(left, right db.TransactionGroup, aggregates map[int64]*jsonGroupAggregate, options groupsSortOptions) bool {
	compare := 0
	switch options.Field {
	case "name":
		if left.Name < right.Name {
			compare = -1
		} else if left.Name > right.Name {
			compare = 1
		}
	case "start_date":
		compare = compareGroupDates(aggregates[left.ID].start, aggregates[right.ID].start)
	case "end_date":
		compare = compareGroupDates(aggregates[left.ID].end, aggregates[right.ID].end)
	}
	if compare == 0 {
		if left.Name < right.Name {
			compare = -1
		} else if left.Name > right.Name {
			compare = 1
		}
	}
	if options.Desc {
		return compare > 0
	}
	return compare < 0
}

func getPlacesData(cashDb db.DBTX) (output.PlacesData, error) {
	places, err := db.ListPlaces(cashDb, db.PlaceListFilter{})
	if err != nil {
		return output.PlacesData{}, err
	}
	data := output.PlacesData{Places: make([]output.PlaceListItem, 0, len(places))}
	for _, place := range places {
		data.Places = append(data.Places, output.PlaceListItem{ID: place.ID, Name: place.Name})
	}
	return data, nil
}

func getTagsData(cashDb db.DBTX) (output.TagsData, error) {
	tags, err := db.ListTags(cashDb, db.TagListFilter{})
	if err != nil {
		return output.TagsData{}, err
	}
	data := output.TagsData{Tags: make([]output.TagListItem, 0, len(tags))}
	for _, tag := range tags {
		count, err := db.CountTransactionsByTagID(cashDb, tag.ID)
		if err != nil {
			return output.TagsData{}, err
		}
		data.Tags = append(data.Tags, output.TagListItem{ID: tag.ID, Name: tag.Name, Transactions: count})
	}
	return data, nil
}

func getCategoriesData(cashDb db.DBTX) (output.CategoriesData, error) {
	categories, err := db.ListCategories(cashDb, db.CategoryListFilter{})
	if err != nil {
		return output.CategoriesData{}, err
	}
	transactions, err := db.ListTransactions(cashDb, []db.SQLFilter{}, []db.Filter[db.Transaction]{}, false)
	if err != nil {
		return output.CategoriesData{}, err
	}
	type stats struct {
		count             int
		expenses, incomes float64
	}
	base := make(map[int64]*stats)
	children := make(map[int64][]db.Category)
	var top []db.Category
	rootID := int64(0)
	for _, category := range categories {
		if category.Name == "root" {
			rootID = category.ID
			continue
		}
		base[category.ID] = &stats{}
	}
	for _, tx := range transactions {
		if tx.CategoryID == nil || base[*tx.CategoryID] == nil {
			continue
		}
		s := base[*tx.CategoryID]
		switch tx.Type {
		case "income":
			s.count++
			s.incomes += tx.Amount
		case "expense":
			s.count++
			s.expenses += tx.Amount
		}
	}
	for _, category := range categories {
		if category.Name == "root" {
			continue
		}
		if category.ParentID == nil || *category.ParentID == rootID {
			top = append(top, category)
		} else {
			children[*category.ParentID] = append(children[*category.ParentID], category)
		}
	}
	sort.Slice(top, func(i, j int) bool { return top[i].Name < top[j].Name })
	for id, items := range children {
		sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
		children[id] = items
	}
	aggregate := make(map[int64]*stats)
	visiting := make(map[int64]bool)
	var rollup func(int64) *stats
	rollup = func(id int64) *stats {
		if aggregate[id] != nil {
			return aggregate[id]
		}
		if visiting[id] {
			return base[id]
		}
		visiting[id] = true
		source := base[id]
		total := &stats{count: source.count, expenses: source.expenses, incomes: source.incomes}
		for _, child := range children[id] {
			childTotal := rollup(child.ID)
			total.count += childTotal.count
			total.expenses += childTotal.expenses
			total.incomes += childTotal.incomes
		}
		visiting[id] = false
		aggregate[id] = total
		return total
	}
	for id := range base {
		rollup(id)
	}
	data := output.CategoriesData{Categories: make([]output.CategoryListItem, 0)}
	var appendRows func([]db.Category, int)
	appendRows = func(items []db.Category, depth int) {
		for _, category := range items {
			s := aggregate[category.ID]
			data.Categories = append(data.Categories, output.CategoryListItem{ID: category.ID, Name: category.Name, ParentID: category.ParentID, Depth: depth, Transactions: s.count, Expenses: s.expenses, Incomes: s.incomes, Net: s.incomes + s.expenses})
			appendRows(children[category.ID], depth+1)
		}
	}
	appendRows(top, 0)
	return data, nil
}

func accountListJSON(parsed parser.ParsedCmdLine, cfg config.Config, cashDb db.DBTX) error {
	format, err := commandOutputFormat(parsed)
	if err != nil {
		return err
	}
	if format != output.FormatJSON {
		return listAccounts(cfg, cashDb)
	}
	data, err := getAccountListData(cashDb)
	if err != nil {
		return err
	}
	return renderJSON("accounts", data, len(data.Accounts))
}
