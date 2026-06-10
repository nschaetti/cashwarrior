package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/domain"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

const (
	FilterUnknown = iota
	FilterTypeTransactionID
	FilterTypeAccountName
	FilterTypeCurrency
	FilterTypeStore
	FilterTypeDescription
	FilterTypeDatetime
	FilterTypeGroup
	FilterTypeIdentifier
)

func classifyFilter(tokenFilter parser.Token) int {
	switch tokenFilter.Kind {
	case parser.TokenText:
		if tokenFilter.Raw[0] == 'T' {
			return FilterTypeTransactionID
		}
	case parser.TokenAttribute:
		if tokenFilter.Attribute.Key == "account" {
			return FilterTypeAccountName
		} else if tokenFilter.Attribute.Key == "currency" {
			return FilterTypeCurrency
		} else if tokenFilter.Attribute.Key == "store" {
			return FilterTypeStore
		} else if tokenFilter.Attribute.Key == "desc" {
			return FilterTypeDescription
		} else if tokenFilter.Attribute.Key == "date" {
			return FilterTypeDatetime
		} else if tokenFilter.Attribute.Key == "period" {
			return FilterTypeDatetime
		} else if tokenFilter.Attribute.Key == "time" {
			return FilterTypeDatetime
		} else if tokenFilter.Attribute.Key == "group" {
			return FilterTypeGroup
		} else if tokenFilter.Attribute.Key == "identifier" {
			return FilterTypeIdentifier
		}
	default:
		return FilterUnknown
	}
	return FilterUnknown
}

func createTransactionIDFilter(token parser.Token) (db.SQLFilter, error) {
	transactionID := token.Raw
	if token.Kind == parser.TokenText && len(token.Raw) > 0 && token.Raw[0] == 'T' {
		transactionID = token.Raw[1:]
	}
	_, err := domain.ParseTransactionID(transactionID)
	if err != nil {
		return nil, err
	}
	return db.TransactionIDFilter{ID: transactionID}, nil
}

func createAccountNameFilter(token parser.Token) (db.SQLFilter, error) {
	return db.TransactionAccountNameFilter{Name: token.Attribute.Value.Raw}, nil
}

func createCurrencyFilter(token parser.Token) (db.SQLFilter, error) {
	return db.TransactionCurrencyFilter{Currency: token.Attribute.Value.Raw}, nil
}

func createStoreFilter(token parser.Token) (db.SQLFilter, error) {
	return db.TransactionStoreNameFilter{Store: token.Attribute.Value.Raw}, nil
}

func createDescriptionFilter(token parser.Token) (db.SQLFilter, error) {
	return db.TransactionDescriptionFilter{Description: token.Attribute.Value.Raw}, nil
}

func parseDateOnly(value string, config config.Config) (time.Time, error) {
	dateFormat := strings.Split(config.Display.DateFormat, " ")[0]
	return time.Parse(dateFormat, value)
}

func createDatetimeFilter(token parser.Token, config config.Config) (db.SQLFilter, error) {
	toDateFilter := func(from time.Time, to time.Time) db.SQLFilter {
		return db.TransactionDateFilter{From: from.Format("2006-01-02"), To: to.Format("2006-01-02")}
	}
	parseFlexibleDate := func(value string) (time.Time, error) {
		value = strings.TrimSpace(value)
		layouts := []string{
			"2006-01-02",
			"2006-01-02 15:04:05",
			"2006-01-02 15:04",
			"02.01.2006",
			"02.01.2006 15:04:05",
			"02.01.2006 15:04",
			"02/01/2006",
			"02/01/2006 15:04:05",
			"02/01/2006 15:04",
			config.Display.DateFormat,
		}
		seen := make(map[string]bool, len(layouts))
		for _, layout := range layouts {
			if layout == "" || seen[layout] {
				continue
			}
			seen[layout] = true
			parsed, err := time.Parse(layout, value)
			if err == nil {
				return parsed, nil
			}
		}
		return time.Time{}, fmt.Errorf("unknown datetime format: %s", value)
	}

	if domain.IsTimeShortcut(token.Attribute.Value.Raw) {
		from, to, err := domain.GetTimeShortcut(token.Attribute.Value.Raw)
		if err != nil {
			return nil, err
		}
		return toDateFilter(from, to), nil
	}

	if token.Attribute.Key == "date" {
		if strings.Contains(token.Attribute.Value.Raw, "..") {
			datetimeRange := strings.SplitN(token.Attribute.Value.Raw, "..", 2)
			datetimeFrom, err := parseFlexibleDate(datetimeRange[0])
			if err != nil {
				return nil, err
			}
			datetimeTo, err := parseFlexibleDate(datetimeRange[1])
			if err != nil {
				return nil, err
			}
			return toDateFilter(datetimeFrom, datetimeTo), nil
		}

		datetime, err := parseFlexibleDate(token.Attribute.Value.Raw)
		if err == nil {
			return toDateFilter(datetime, datetime), nil
		}
	}

	if strings.Contains(token.Attribute.Value.Raw, "..") {
		datetimeRange := strings.SplitN(token.Attribute.Value.Raw, "..", 2)
		datetimeFrom, err := parseFlexibleDate(datetimeRange[0])
		if err != nil {
			return nil, err
		}
		datetimeTo, err := parseFlexibleDate(datetimeRange[1])
		if err != nil {
			return nil, err
		}
		return toDateFilter(datetimeFrom, datetimeTo), nil
	}

	if token.Attribute.Key == "time" {
		now := time.Now()
		return toDateFilter(now, now), nil
	}

	if token.Attribute.Key == "datetime" {
		datetime, err := parseFlexibleDate(token.Attribute.Value.Raw)
		if err == nil {
			return toDateFilter(datetime, datetime), nil
		}
	}

	return nil, fmt.Errorf("unknown datetime format: %s. Must be given as shortcuts, exact date, or from..to", token.Attribute.Value)
}

func createGroupFilter(token parser.Token) (db.SQLFilter, error) {
	return db.TransactionGroupNameFilter{Name: token.Attribute.Value.Raw}, nil
}

func createIdentifierFilter(token parser.Token) (db.SQLFilter, error) {
	_, err := domain.ParseTransactionID(token.Attribute.Value.Raw)
	if err != nil {
		return nil, err
	}
	return db.TransactionIDFilter{ID: token.Attribute.Value.Raw}, nil
}

func createTransactionFilters(
	parsed parser.ParsedCmdLine,
	config config.Config,
) ([]db.SQLFilter, []db.Filter[db.Transaction], error) {
	var dbFilters []db.SQLFilter = make([]db.SQLFilter, 0, len(parsed.Filters))
	var runFilters []db.Filter[db.Transaction] = make([]db.Filter[db.Transaction], 0, len(parsed.Filters))
	for _, filter := range parsed.Filters {
		filterType := classifyFilter(filter)
		if filterType == FilterUnknown {
			return nil, nil, fmt.Errorf("unknown filter: %s", filter.Raw)
		} else if filterType == FilterTypeTransactionID {
			newFilter, err := createTransactionIDFilter(filter)
			if err != nil {
				return nil, nil, err
			}
			dbFilters = append(dbFilters, newFilter)
		} else if filterType == FilterTypeAccountName {
			newFilter, err := createAccountNameFilter(filter)
			if err != nil {
				return nil, nil, err
			}
			dbFilters = append(dbFilters, newFilter)
		} else if filterType == FilterTypeCurrency {
			newFilter, err := createCurrencyFilter(filter)
			if err != nil {
				return nil, nil, err
			}
			dbFilters = append(dbFilters, newFilter)
		} else if filterType == FilterTypeStore {
			newFilter, err := createStoreFilter(filter)
			if err != nil {
				return nil, nil, err
			}
			dbFilters = append(dbFilters, newFilter)
		} else if filterType == FilterTypeDescription {
			newFilter, err := createDescriptionFilter(filter)
			if err != nil {
				return nil, nil, err
			}
			dbFilters = append(dbFilters, newFilter)
		} else if filterType == FilterTypeDatetime {
			newFilter, err := createDatetimeFilter(filter, config)
			if err != nil {
				return nil, nil, err
			}
			dbFilters = append(dbFilters, newFilter)
		} else if filterType == FilterTypeGroup {
			newFilter, err := createGroupFilter(filter)
			if err != nil {
				return nil, nil, err
			}
			dbFilters = append(dbFilters, newFilter)
		} else if filterType == FilterTypeIdentifier {
			newFilter, err := createIdentifierFilter(filter)
			if err != nil {
				return nil, nil, err
			}
			dbFilters = append(dbFilters, newFilter)
		}
	}
	return dbFilters, runFilters, nil
}

func createFilters(
	parsed parser.ParsedCmdLine,
	_ db.DBTX,
	config config.Config,
) ([]db.SQLFilter, []db.Filter[db.Transaction], error) {
	return createTransactionFilters(parsed, config)
}
