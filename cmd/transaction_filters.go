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
)

func classifyFilter(tokenFilter parser.Token) int {
	switch tokenFilter.Kind {
	case parser.TokenPeriod:
		return FilterTypeDatetime
	case parser.TokenText:
		if tokenFilter.Raw[0] == 'T' {
			return FilterTypeTransactionID
		}
	case parser.TokenAttribute:
		if tokenFilter.Key == "account" {
			return FilterTypeAccountName
		} else if tokenFilter.Key == "currency" {
			return FilterTypeCurrency
		} else if tokenFilter.Key == "store" {
			return FilterTypeStore
		} else if tokenFilter.Key == "desc" {
			return FilterTypeDescription
		} else if tokenFilter.Key == "date" {
			return FilterTypeDatetime
		} else if tokenFilter.Key == "period" {
			return FilterTypeDatetime
		} else if tokenFilter.Key == "time" {
			return FilterTypeDatetime
		} else if tokenFilter.Key == "group" {
			return FilterTypeGroup
		}
	default:
		return FilterUnknown
	}
	return FilterUnknown
}

func createTransactionIDFilter(token parser.Token) (db.SQLFilter, error) {
	var transactionID string = token.Raw[1:]
	_, err := domain.ParseTransactionID(transactionID)
	if err != nil {
		return nil, err
	}
	return db.TransactionIDFilter{ID: transactionID}, nil
}

func createAccountNameFilter(token parser.Token) (db.SQLFilter, error) {
	return db.TransactionAccountNameFilter{Name: token.Value}, nil
}

func createCurrencyFilter(token parser.Token) (db.SQLFilter, error) {
	return db.TransactionCurrencyFilter{Currency: token.Value}, nil
}

func createStoreFilter(token parser.Token) (db.SQLFilter, error) {
	return db.TransactionStoreNameFilter{Store: token.Value}, nil
}

func createDescriptionFilter(token parser.Token) (db.SQLFilter, error) {
	return db.TransactionDescriptionFilter{Description: token.Value}, nil
}

func parseDateOnly(value string, config config.Config) (time.Time, error) {
	dateFormat := strings.Split(config.Display.DateFormat, " ")[0]
	return time.Parse(dateFormat, value)
}

func createDatetimeFilter(token parser.Token, config config.Config) (db.SQLFilter, error) {
	if token.Kind == parser.TokenPeriod {
		from, to, err := domain.GetTimeShortcut(token.Raw)
		if err != nil {
			return nil, err
		}
		return db.TransactionDatetimeFilter{From: from, To: to}, nil
	}

	if domain.IsTimeShortcut(token.Value) {
		from, to, err := domain.GetTimeShortcut(token.Value)
		if err != nil {
			return nil, err
		}
		return db.TransactionDatetimeFilter{From: from, To: to}, nil
	}

	if token.Key == "date" {
		if strings.Contains(token.Value, "..") {
			datetimeRange := strings.SplitN(token.Value, "..", 2)
			datetimeFrom, err := parseDateOnly(datetimeRange[0], config)
			if err != nil {
				return nil, err
			}
			datetimeTo, err := parseDateOnly(datetimeRange[1], config)
			if err != nil {
				return nil, err
			}
			return db.TransactionDatetimeFilter{From: datetimeFrom, To: datetimeTo.Add(24*time.Hour - time.Nanosecond)}, nil
		}

		datetime, err := parseDateOnly(token.Value, config)
		if err == nil {
			return db.TransactionDatetimeFilter{From: datetime, To: datetime.Add(24*time.Hour - time.Nanosecond)}, nil
		}
	}

	if strings.Contains(token.Value, "..") {
		datetimeRange := strings.SplitN(token.Value, "..", 2)
		datetimeFrom, err := time.Parse(config.Display.DateFormat, datetimeRange[0])
		if err != nil {
			return nil, err
		}
		datetimeTo, err := time.Parse(config.Display.DateFormat, datetimeRange[1])
		if err != nil {
			return nil, err
		}
		return db.TransactionDatetimeFilter{From: datetimeFrom, To: datetimeTo}, nil
	}

	if token.Key == "datetime" {
		datetime, err := time.Parse(config.Display.DateFormat, token.Value)
		if err == nil {
			return db.TransactionDatetimeFilter{From: datetime, To: datetime}, nil
		}
	}

	return nil, fmt.Errorf("unknown datetime format: %s. Must be given as shortcuts, exact date, or from..to", token.Value)
}

func createGroupFilter(token parser.Token) (db.SQLFilter, error) {
	return db.TransactionGroupNameFilter{Name: token.Value}, nil
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
