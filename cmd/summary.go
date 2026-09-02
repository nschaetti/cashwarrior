package cmd

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/gui"
	"github.com/nschaetti/cashwarrior/internal/output"
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
	transactions, err := db.ListTransactions(cashDb, dbFilters, runFilters, false)
	if err != nil {
		return err
	}
	data := buildSummaryDaysData(transactions)
	format, err := commandOutputFormat(parsed)
	if err != nil {
		return err
	}
	if format == output.FormatJSON {
		return renderJSON("summary_days", data, len(data.Days))
	}

	return printSummaryDaysTable(data)
}

func buildSummaryDaysData(transactions []db.Transaction) output.SummaryDaysData {
	byDay := make(map[string]*output.SummaryDay)
	for _, transaction := range transactions {
		day := transaction.Datetime.Format("2006-01-02")
		item, ok := byDay[day]
		if !ok {
			item = &output.SummaryDay{Date: day, Currencies: make([]output.SummaryDayCurrency, 0)}
			byDay[day] = item
		}
		item.Transactions++
		currency := transaction.Currency
		if currency == "" {
			currency = "N/A"
		}
		var totals *output.SummaryDayCurrency
		for index := range item.Currencies {
			if item.Currencies[index].Currency == currency {
				totals = &item.Currencies[index]
				break
			}
		}
		if totals == nil {
			item.Currencies = append(item.Currencies, output.SummaryDayCurrency{Currency: currency})
			totals = &item.Currencies[len(item.Currencies)-1]
		}
		if transaction.Amount >= 0 {
			totals.Income += transaction.Amount
		} else {
			totals.Expenses += transaction.Amount
		}
		totals.Net = totals.Income + totals.Expenses
	}
	days := make([]output.SummaryDay, 0, len(byDay))
	for _, day := range byDay {
		sort.Slice(day.Currencies, func(i, j int) bool { return day.Currencies[i].Currency < day.Currencies[j].Currency })
		days = append(days, *day)
	}
	sort.Slice(days, func(i, j int) bool { return days[i].Date < days[j].Date })
	return output.SummaryDaysData{Days: days}
}

func printSummaryDaysTable(data output.SummaryDaysData) error {
	currenciesSet := make(map[string]bool)
	for _, day := range data.Days {
		for _, currency := range day.Currencies {
			currenciesSet[currency.Currency] = true
		}
	}
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

	rows := make([][]string, 0, len(data.Days))
	for _, day := range data.Days {
		row := []string{day.Date, strconv.Itoa(day.Transactions)}
		for _, currency := range currencies {
			var totals *output.SummaryDayCurrency
			for index := range day.Currencies {
				if day.Currencies[index].Currency == currency {
					totals = &day.Currencies[index]
					break
				}
			}
			if totals == nil {
				row = append(row, "0.00", "0.00", "0.00")
				continue
			}
			row = append(row,
				fmt.Sprintf("%.2f", totals.Expenses),
				fmt.Sprintf("%.2f", totals.Income),
				fmt.Sprintf("%+.2f", totals.Net),
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
