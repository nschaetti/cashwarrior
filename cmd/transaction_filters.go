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

func classifyFilter(argFilter parser.Arg) int {
	switch a := argFilter.(type) {
	case parser.ArgText:
		if a.Text[0] == 'T' {
			return FilterTypeTransactionID
		}
	case parser.ArgAttribute:
		if a.Key == "account" {
			return FilterTypeAccountName
		} else if a.Key == "currency" {
			return FilterTypeCurrency
		} else if a.Key == "store" {
			return FilterTypeStore
		} else if a.Key == "desc" {
			return FilterTypeDescription
		} else if a.Key == "date" {
			return FilterTypeDatetime
		} else if a.Key == "period" {
			return FilterTypeDatetime
		} else if a.Key == "time" {
			return FilterTypeDatetime
		} else if a.Key == "group" {
			return FilterTypeGroup
		} else if a.Key == "identifier" {
			return FilterTypeIdentifier
		}
	default:
		return FilterUnknown
	}
	return FilterUnknown
}

func createTransactionIDFilter(arg parser.Arg) (db.SQLFilter, error) {
	transactionID := arg.RawString()
	text, ok := arg.(parser.ArgText)
	if ok && len(text.Text) > 0 && text.Text[0] == 'T' {
		transactionID = text.Text[1:]
	}
	_, err := domain.ParseTransactionID(transactionID)
	if err != nil {
		return nil, err
	}
	return db.TransactionIDFilter{ID: transactionID}, nil
}

func createAccountNameFilter(arg parser.Arg) (db.SQLFilter, error) {
	attr, isAttr := arg.(parser.ArgAttribute)
	text, isText := arg.(parser.ArgText)
	if !isAttr && !isText {
		return nil, fmt.Errorf("account filter requires an account name")
	}
	if isAttr {
		return db.TransactionAccountNameFilter{Name: attr.Value.Raw}, nil
	}
	return db.TransactionAccountNameFilter{Name: text.Text}, nil
}

func createCurrencyFilter(arg parser.Arg) (db.SQLFilter, error) {
	attr, isAttr := arg.(parser.ArgAttribute)
	if !isAttr {
		return nil, fmt.Errorf("currency filter requires a currency")
	}
	return db.TransactionCurrencyFilter{Currency: attr.Value.Raw}, nil
}

func createStoreFilter(arg parser.Arg) (db.SQLFilter, error) {
	attr, isAttr := arg.(parser.ArgAttribute)
	if !isAttr {
		return nil, fmt.Errorf("store filter requires a store name")
	}
	return db.TransactionStoreNameFilter{Store: attr.Value.Raw}, nil
}

func createDescriptionFilter(arg parser.Arg) (db.SQLFilter, error) {
	attr, isAttr := arg.(parser.ArgAttribute)
	if !isAttr {
		return nil, fmt.Errorf("description filter requires a description")
	}
	return db.TransactionDescriptionFilter{Description: attr.Value.Raw}, nil
}

func parseDateOnly(value string, config config.Config) (time.Time, error) {
	dateFormat := strings.Split(config.Display.DateFormat, " ")[0]
	return time.Parse(dateFormat, value)
}

func createDatetimeFilter(arg parser.Arg, config config.Config) (db.SQLFilter, error) {
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

	attr, ok := arg.(parser.ArgAttribute)
	if !ok {
		return nil, fmt.Errorf("datetime filter requires a datetime")
	}

	if domain.IsTimeShortcut(attr.Value.Raw) {
		from, to, err := domain.GetTimeShortcut(attr.Value.Raw)
		if err != nil {
			return nil, err
		}
		return toDateFilter(from, to), nil
	}

	if attr.Key == "date" {
		if strings.Contains(attr.Value.Raw, "..") {
			datetimeRange := strings.SplitN(attr.Value.Raw, "..", 2)
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

		datetime, err := parseFlexibleDate(attr.Value.Raw)
		if err == nil {
			return toDateFilter(datetime, datetime), nil
		}
	}

	if strings.Contains(attr.Value.Raw, "..") {
		datetimeRange := strings.SplitN(attr.Value.Raw, "..", 2)
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

	if attr.Key == "time" {
		now := time.Now()
		return toDateFilter(now, now), nil
	}

	if attr.Key == "datetime" {
		datetime, err := parseFlexibleDate(attr.Value.Raw)
		if err == nil {
			return toDateFilter(datetime, datetime), nil
		}
	}

	return nil, fmt.Errorf("unknown datetime format: %s. Must be given as shortcuts, exact date, or from..to", attr.Value)
}

func createGroupFilter(arg parser.Arg) (db.SQLFilter, error) {
	attr, isAttr := arg.(parser.ArgAttribute)
	if !isAttr {
		return nil, fmt.Errorf("group filter requires a group name")
	}
	return db.TransactionGroupNameFilter{Name: attr.Value.Raw}, nil
}

func createIdentifierFilter(arg parser.Arg) (db.SQLFilter, error) {
	attr, isAttr := arg.(parser.ArgAttribute)
	if !isAttr {
		return nil, fmt.Errorf("identifier filter requires an identifier")
	}
	_, err := domain.ParseTransactionID(attr.Value.Raw)
	if err != nil {
		return nil, err
	}
	return db.TransactionIDFilter{ID: attr.Value.Raw}, nil
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
			return nil, nil, fmt.Errorf("unknown filter: %s", filter.RawString())
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
