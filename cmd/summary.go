package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Summary(parsed parser.ParsedCmdLine, cfg config.Config, cashDb db.DBTX) error {
	switch parsed.Subcommand {
	case "days":
		return summaryDays(parsed, cfg, cashDb)
	default:
		return fmt.Errorf("unknown summary subcommand %s", parsed.Subcommand)
	}
}

func summaryDays(parsed parser.ParsedCmdLine, cfg config.Config, cashDb db.DBTX) error {
	parsedForFilters := parsed
	parsedForFilters.Filters = append(append([]parser.Arg{}, parsed.Filters...), parsed.Args...)
	parsedForFilters.Args = []parser.Arg{}

	dbFilters, runFilters, err := createFilters(parsedForFilters, cashDb, cfg)
	if err != nil {
		return err
	}
	transactions, err := db.ListTransactions(cashDb, dbFilters, runFilters)
	if err != nil {
		return err
	}

	counts := make(map[string]int)
	type dayCurrencyTotals struct {
		Income   float64
		Expenses float64
	}
	totalsByDayCurrency := make(map[string]map[string]*dayCurrencyTotals)
	currenciesSet := make(map[string]bool)
	for _, transaction := range transactions {
		day := time.Date(transaction.Datetime.Year(), transaction.Datetime.Month(), transaction.Datetime.Day(), 0, 0, 0, 0, transaction.Datetime.Location())
		key := day.Format("2006-01-02")
		counts[key]++

		currency := transaction.Currency
		if currency == "" {
			currency = "N/A"
		}
		currenciesSet[currency] = true

		if _, ok := totalsByDayCurrency[key]; !ok {
			totalsByDayCurrency[key] = make(map[string]*dayCurrencyTotals)
		}
		if _, ok := totalsByDayCurrency[key][currency]; !ok {
			totalsByDayCurrency[key][currency] = &dayCurrencyTotals{}
		}

		if transaction.Amount >= 0 {
			totalsByDayCurrency[key][currency].Income += transaction.Amount
		} else {
			totalsByDayCurrency[key][currency].Expenses += transaction.Amount
		}
	}

	days := make([]string, 0, len(counts))
	for day := range counts {
		days = append(days, day)
	}
	sort.Strings(days)

	currencies := make([]string, 0, len(currenciesSet))
	for currency := range currenciesSet {
		currencies = append(currencies, currency)
	}
	sort.Strings(currencies)

	headers := []string{"Date", "Transactions"}
	for _, currency := range currencies {
		headers = append(headers,
			fmt.Sprintf("Expenses (%s)", currency),
			fmt.Sprintf("Income (%s)", currency),
			fmt.Sprintf("Net (%s)", currency),
		)
	}

	rows := make([][]string, 0, len(days))
	for _, day := range days {
		row := []string{day, strconv.Itoa(counts[day])}
		for _, currency := range currencies {
			totals := totalsByDayCurrency[day][currency]
			if totals == nil {
				row = append(row, "0.00", "0.00", "0.00")
				continue
			}
			net := totals.Income + totals.Expenses
			row = append(row,
				fmt.Sprintf("%.2f", totals.Expenses),
				fmt.Sprintf("%.2f", totals.Income),
				fmt.Sprintf("%+.2f", net),
			)
		}
		rows = append(rows, row)
	}

	theme := gui.CurrentTheme()
	t := gui.NewTable().
		WithTitle("Summary by Day", theme.TransactionListTitleBackground).
		WithSubtitle("Transactions per day").
		WithHeaderBackground(theme.TransactionListHeaderBackground).
		WithHeaders(headers...).
		AddRows(rows)

	fmt.Println(t.Render())
	fmt.Printf(" Returned %d days\n\n\n", len(rows))
	return nil
}
