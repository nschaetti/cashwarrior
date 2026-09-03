package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/domain"
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/output"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

type balanceInterval struct {
	Start *time.Time
	End   time.Time
}

func Balance(parsed parser.ParsedCmdLine, cfg config.Config, query db.DBTX) error {
	now := time.Now()
	accountNames, currencies, interval, err := parseBalanceFilters(parsed.Filters, now)
	if err != nil {
		return err
	}

	accounts, err := selectBalanceAccounts(query, accountNames, currencies)
	if err != nil {
		return err
	}
	transactions, err := db.ListTransactions(query, []db.SQLFilter{}, []db.Filter[db.Transaction]{}, false)
	if err != nil {
		return err
	}

	data := calculateBalance(accounts, transactions, interval)
	format, err := commandOutputFormat(parsed)
	if err != nil {
		return err
	}
	if format == output.FormatJSON {
		return renderJSON("balance", data, len(data.Accounts))
	}
	return printBalanceTables(data)
}

func parseBalanceFilters(filters []parser.Arg, now time.Time) ([]string, []string, balanceInterval, error) {
	var accountNames []string
	var currencies []string
	interval := balanceInterval{End: now}
	timeFilters := 0
	var err error

	for _, arg := range filters {
		switch value := arg.(type) {
		case parser.ArgText:
			if !domain.IsTimeShortcut(value.Text) {
				return nil, nil, interval, fmt.Errorf("unknown balance filter: %s", value.RawString())
			}
			timeFilters++
			interval, err = resolveBalanceShortcut(value.Text, now)
			if err != nil {
				return nil, nil, interval, err
			}
		case parser.ArgAttribute:
			switch value.Key {
			case "account":
				accountNames, err = balanceStringValues(value.Value)
				if err != nil {
					return nil, nil, interval, fmt.Errorf("invalid account filter: %w", err)
				}
			case "currency":
				currencies, err = balanceStringValues(value.Value)
				if err != nil {
					return nil, nil, interval, fmt.Errorf("invalid currency filter: %w", err)
				}
			case "date":
				timeFilters++
				interval, err = resolveBalanceDate(value.Value, now)
				if err != nil {
					return nil, nil, interval, err
				}
			default:
				return nil, nil, interval, fmt.Errorf("unknown balance filter: %s", value.RawString())
			}
		default:
			return nil, nil, interval, fmt.Errorf("unknown balance filter: %s", arg.RawString())
		}
	}

	if timeFilters > 1 {
		return nil, nil, interval, fmt.Errorf("balance accepts only one time filter")
	}
	return accountNames, currencies, interval, nil
}

func balanceStringValues(value parser.AttributeValue) ([]string, error) {
	values := make([]string, 0, 1)
	switch value.ValueShape {
	case parser.AttributeValueShapeSingle:
		item, ok := value.Value.(parser.StringItem)
		if !ok {
			return nil, fmt.Errorf("expected a string value")
		}
		values = append(values, item.Value)
	case parser.AttributeValueShapeList:
		for _, rawItem := range value.Items {
			item, ok := rawItem.(parser.StringItem)
			if !ok {
				return nil, fmt.Errorf("expected a string value")
			}
			values = append(values, item.Value)
		}
	default:
		return nil, fmt.Errorf("expected a single value or a list")
	}

	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, fmt.Errorf("value cannot be empty")
		}
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result, nil
}

func resolveBalanceDate(value parser.AttributeValue, now time.Time) (balanceInterval, error) {
	var start time.Time
	var end time.Time

	switch value.ValueShape {
	case parser.AttributeValueShapeSingle:
		item, ok := value.Value.(parser.TimeItem)
		if !ok {
			return balanceInterval{}, fmt.Errorf("invalid date filter")
		}
		start = startOfDay(item.Value)
		end = start.AddDate(0, 0, 1).Add(-time.Nanosecond)
	case parser.AttributeValueShapeRange:
		startItem, startOK := value.Range.Start.(parser.TimeItem)
		endItem, endOK := value.Range.End.(parser.TimeItem)
		if !startOK || !endOK {
			return balanceInterval{}, fmt.Errorf("invalid date range")
		}
		start = startOfDay(startItem.Value)
		end = startOfDay(endItem.Value).AddDate(0, 0, 1).Add(-time.Nanosecond)
		if end.Before(start) {
			return balanceInterval{}, fmt.Errorf("balance date range ends before it starts")
		}
	case parser.AttributeValueShapeShortcut:
		return resolveBalanceShortcut(value.Shortcut.Name, now)
	default:
		return balanceInterval{}, fmt.Errorf("balance date must be a day, range, or shortcut")
	}

	if !now.Before(start) && now.Before(end) {
		end = now
	}
	return balanceInterval{Start: &start, End: end}, nil
}

func resolveBalanceShortcut(name string, now time.Time) (balanceInterval, error) {
	start, end, err := domain.GetTimeShortcut(name)
	if err != nil {
		return balanceInterval{}, err
	}
	if !now.Before(start) && now.Before(end) {
		end = now
	}
	return balanceInterval{Start: &start, End: end}, nil
}

func startOfDay(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}

func selectBalanceAccounts(query db.DBTX, requestedNames []string, requestedCurrencies []string) ([]db.Account, error) {
	accounts, err := db.ListAccounts(query, []db.SQLFilter{}, []db.Filter[db.Account]{}, []string{})
	if err != nil {
		return nil, err
	}

	knownNames := make(map[string]bool, len(accounts))
	for _, account := range accounts {
		knownNames[account.Name] = true
	}
	for _, name := range requestedNames {
		if !knownNames[name] {
			return nil, fmt.Errorf("account %s does not exist", name)
		}
	}

	wantedNames := stringSet(requestedNames)
	wantedCurrencies := stringSet(requestedCurrencies)
	selected := make([]db.Account, 0, len(accounts))
	for _, account := range accounts {
		if len(wantedNames) > 0 && !wantedNames[account.Name] {
			continue
		}
		if len(wantedCurrencies) > 0 && !wantedCurrencies[account.Currency] {
			continue
		}
		selected = append(selected, account)
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].Name < selected[j].Name })
	return selected, nil
}

func stringSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}

func calculateBalance(accounts []db.Account, transactions []db.Transaction, interval balanceInterval) output.BalanceData {
	data := output.BalanceData{
		Period: output.BalancePeriod{To: interval.End.Format("2006-01-02")},
		Accounts:   make([]output.BalanceAccountItem, len(accounts)),
		Currencies: make([]output.BalanceCurrencyItem, 0),
	}
	if interval.Start != nil {
		from := interval.Start.Format("2006-01-02")
		data.Period.From = &from
	}

	accountIndex := make(map[int64]int, len(accounts))
	for index, account := range accounts {
		accountIndex[account.ID] = index
		data.Accounts[index] = output.BalanceAccountItem{
			Account:  account.Name,
			Currency: account.Currency,
			Opening:  account.InitialBalance,
		}
	}

	for _, transaction := range transactions {
		if transaction.AccountID == nil || transaction.Datetime.After(interval.End) {
			continue
		}
		index, ok := accountIndex[*transaction.AccountID]
		if !ok {
			continue
		}
		item := &data.Accounts[index]
		if interval.Start != nil && transaction.Datetime.Before(*interval.Start) {
			item.Opening += transaction.Amount
			continue
		}

		item.Net += transaction.Amount
		item.Operations++
		switch transaction.Type {
		case "income":
			item.Income += transaction.Amount
		case "expense":
			item.Expenses += transaction.Amount
		case "transfer_in", "transfer_out":
			item.Transfers += transaction.Amount
		}
	}

	byCurrency := make(map[string]*output.BalanceCurrencyItem)
	for index := range data.Accounts {
		item := &data.Accounts[index]
		item.Closing = item.Opening + item.Net
		total := byCurrency[item.Currency]
		if total == nil {
			total = &output.BalanceCurrencyItem{Currency: item.Currency}
			byCurrency[item.Currency] = total
		}
		total.Opening += item.Opening
		total.Income += item.Income
		total.Expenses += item.Expenses
		total.Transfers += item.Transfers
		total.Net += item.Net
		total.Closing += item.Closing
		total.Operations += item.Operations
	}

	currencies := make([]string, 0, len(byCurrency))
	for currency := range byCurrency {
		currencies = append(currencies, currency)
	}
	sort.Strings(currencies)
	for _, currency := range currencies {
		data.Currencies = append(data.Currencies, *byCurrency[currency])
	}
	return data
}

func printBalanceTables(data output.BalanceData) error {
	rows := make([][]string, 0, len(data.Accounts))
	for _, item := range data.Accounts {
		rows = append(rows, balanceRow(item.Account, item.Currency, item.Opening, item.Income, item.Expenses, item.Transfers, item.Net, item.Closing, item.Operations))
	}

	theme := gui.CurrentTheme()
	period := "through " + data.Period.To
	if data.Period.From != nil {
		period = *data.Period.From + " to " + data.Period.To
	}
	table := gui.NewTable().
		WithTitle("Balances", theme.AccountsTitleBackground).
		WithSubtitle(period).
		WithHeaderBackground(theme.AccountsHeaderBackground).
		WithHeaders("Account", "Currency", "Opening", "Income", "Expenses", "Transfers", "Net", "Closing", "Operations").
		AddRows(rows)
	fmt.Println(table.Render())

	if len(data.Currencies) > 0 {
		currencyRows := make([][]string, 0, len(data.Currencies))
		for _, item := range data.Currencies {
			currencyRows = append(currencyRows, balanceRow("Total", item.Currency, item.Opening, item.Income, item.Expenses, item.Transfers, item.Net, item.Closing, item.Operations))
		}
		totals := gui.NewTable().
			WithType(gui.TableTypeSummary).
			WithTitle("Totals by currency", theme.AccountsTitleBackground).
			WithSubtitle(period).
			WithHeaderBackground(theme.AccountsHeaderBackground).
			WithHeaders("Account", "Currency", "Opening", "Income", "Expenses", "Transfers", "Net", "Closing", "Operations").
			AddRows(currencyRows)
		fmt.Println(totals.Render())
	}
	fmt.Printf(" Returned %d accounts\n", len(data.Accounts))
	return nil
}

func balanceRow(name string, currency string, opening float64, income float64, expenses float64, transfers float64, net float64, closing float64, operations int) []string {
	return []string{
		name,
		currency,
		fmt.Sprintf("%+.2f", opening),
		fmt.Sprintf("%+.2f", income),
		fmt.Sprintf("%+.2f", expenses),
		fmt.Sprintf("%+.2f", transfers),
		fmt.Sprintf("%+.2f", net),
		fmt.Sprintf("%+.2f", closing),
		strconv.Itoa(operations),
	}
}
