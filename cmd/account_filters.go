package cmd

import (
	"fmt"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func createAccountNameFilter(arg parser.Arg) (db.SQLFilter, error) {
	attr, isAttr := arg.(parser.ArgAttribute)
	if !isAttr {
		return nil, fmt.Errorf("account filter requires an account name")
	}
	stringAttr, isStringAttr := attr.Value.Value.(parser.StringItem)
	if !isStringAttr {
		return nil, fmt.Errorf("account filter requires a string value")
	}
	return db.AccountAccountNameFilter{Name: stringAttr.Value}, nil
}

func createAccountCurrencyFilter(arg parser.Arg) (db.SQLFilter, error) {
	attr, isAttr := arg.(parser.ArgAttribute)
	if !isAttr {
		return nil, fmt.Errorf("currency filter requires a currency")
	}
	return db.AccountCurrencyFilter{Currency: attr.Value.Raw}, nil
}

func createAccountInitialBalanceFilter(arg parser.Arg) (db.SQLFilter, error) {
	attr, isAttr := arg.(parser.ArgAttribute)
	if !isAttr {
		return nil, fmt.Errorf("initial balance filter requires an initial balance")
	}
	attrAsFloat, convOk := attr.Value.Value.(parser.FloatItem)
	if !convOk {
		return db.AccountInitialBalanceFilter{}, fmt.Errorf("invalid value for initial-balance: %v", attr)
	}
	return db.AccountInitialBalanceFilter{InitialBalance: attrAsFloat.Value}, nil
}

func createAccountFilters(
	parsed parser.ParsedCmdLine,
	config config.Config,
) ([]db.SQLFilter, []db.Filter[db.Account], error) {
	var dbFilters []db.SQLFilter = make([]db.SQLFilter, 0, len(parsed.Filters))
	var runFilters []db.Filter[db.Account] = make([]db.Filter[db.Account], 0, len(parsed.Filters))
	for _, filter := range parsed.Filters {
		filterType := classifyFilter(filter)
		if filterType == FilterUnknown {
			return nil, nil, fmt.Errorf("unknown filter: %s", filter.RawString())
		} else if filterType == FilterTypeAccountName {
			newFilter, err := createAccountNameFilter(filter)
			if err != nil {
				return nil, nil, err
			}
			dbFilters = append(dbFilters, newFilter)
		} else if filterType == FilterTypeCurrency {
			newFilter, err := createAccountCurrencyFilter(filter)
			if err != nil {
				return nil, nil, err
			}
			dbFilters = append(dbFilters, newFilter)
		} else if filterType == FilterTypeInitialBalance {
			newFilter, err := createAccountInitialBalanceFilter(filter)
			if err != nil {
				return nil, nil, err
			}
			dbFilters = append(dbFilters, newFilter)
		}
	}
	return dbFilters, runFilters, nil
}
