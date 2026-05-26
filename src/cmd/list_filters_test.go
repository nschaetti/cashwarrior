package cmd

import (
	"testing"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestClassifyFilter_Currency(t *testing.T) {
	filterType := classifyFilter(parser.Token{Kind: parser.TokenAttribute, Key: "currency", Value: "CHF"})
	if filterType != FilterTypeCurrency {
		t.Fatalf("classifyFilter(currency:CHF) = %d, want %d", filterType, FilterTypeCurrency)
	}
}

func TestClassifyFilter_StoreAndDesc(t *testing.T) {
	storeFilterType := classifyFilter(parser.Token{Kind: parser.TokenAttribute, Key: "store", Value: "%migros%"})
	if storeFilterType != FilterTypeStore {
		t.Fatalf("classifyFilter(store:%%migros%%) = %d, want %d", storeFilterType, FilterTypeStore)
	}

	descFilterType := classifyFilter(parser.Token{Kind: parser.TokenAttribute, Key: "desc", Value: "%coca%"})
	if descFilterType != FilterTypeDescription {
		t.Fatalf("classifyFilter(desc:%%coca%%) = %d, want %d", descFilterType, FilterTypeDescription)
	}
}

func TestCreateFilters_CurrencyAndOtherFilters(t *testing.T) {
	parsed := parser.ParsedCmdLine{
		Command: "list",
		Filters: []parser.Token{
			{Raw: "currency:chf", Kind: parser.TokenAttribute, Key: "currency", Value: "chf"},
			{Raw: "account:cash", Kind: parser.TokenAttribute, Key: "account", Value: "cash"},
			{Raw: "period:today", Kind: parser.TokenAttribute, Key: "period", Value: "today"},
		},
	}

	dbFilters, runFilters, err := createFilters(parsed, nil, config.Config{})
	if err != nil {
		t.Fatalf("createFilters returned error: %v", err)
	}

	if len(runFilters) != 0 {
		t.Fatalf("len(runFilters) = %d, want 0", len(runFilters))
	}

	if len(dbFilters) != 3 {
		t.Fatalf("len(dbFilters) = %d, want 3", len(dbFilters))
	}

	if _, ok := dbFilters[0].(db.TransactionCurrencyFilter); !ok {
		t.Fatalf("dbFilters[0] is %T, want db.TransactionCurrencyFilter", dbFilters[0])
	}
	if _, ok := dbFilters[1].(db.TransactionAccountNameFilter); !ok {
		t.Fatalf("dbFilters[1] is %T, want db.TransactionAccountNameFilter", dbFilters[1])
	}
	if _, ok := dbFilters[2].(db.TransactionDatetimeFilter); !ok {
		t.Fatalf("dbFilters[2] is %T, want db.TransactionDatetimeFilter", dbFilters[2])
	}
}

func TestCreateFilters_StoreAndDescription(t *testing.T) {
	parsed := parser.ParsedCmdLine{
		Command: "list",
		Filters: []parser.Token{
			{Raw: "desc:%coca%", Kind: parser.TokenAttribute, Key: "desc", Value: "%coca%"},
			{Raw: "store:migros%", Kind: parser.TokenAttribute, Key: "store", Value: "migros%"},
		},
	}

	dbFilters, runFilters, err := createFilters(parsed, nil, config.Config{})
	if err != nil {
		t.Fatalf("createFilters returned error: %v", err)
	}

	if len(runFilters) != 0 {
		t.Fatalf("len(runFilters) = %d, want 0", len(runFilters))
	}

	if len(dbFilters) != 2 {
		t.Fatalf("len(dbFilters) = %d, want 2", len(dbFilters))
	}

	if _, ok := dbFilters[0].(db.TransactionDescriptionFilter); !ok {
		t.Fatalf("dbFilters[0] is %T, want db.TransactionDescriptionFilter", dbFilters[0])
	}
	if _, ok := dbFilters[1].(db.TransactionStoreNameFilter); !ok {
		t.Fatalf("dbFilters[1] is %T, want db.TransactionStoreNameFilter", dbFilters[1])
	}
}

func TestTransactionCurrencyFilter_GenerateSQL_IsCaseInsensitive(t *testing.T) {
	filter := db.TransactionCurrencyFilter{Currency: "chf"}

	condition, args := filter.GenerateSQL()
	if condition != "UPPER(accounts.currency) = UPPER(?)" {
		t.Fatalf("condition = %q, want %q", condition, "UPPER(accounts.currency) = UPPER(?)")
	}
	if len(args) != 1 {
		t.Fatalf("len(args) = %d, want 1", len(args))
	}
	if args[0] != "chf" {
		t.Fatalf("args[0] = %v, want %q", args[0], "chf")
	}
}

func TestTransactionStoreAndDescriptionFilters_GenerateSQL(t *testing.T) {
	storeFilter := db.TransactionStoreNameFilter{Store: "%migros%"}
	storeCondition, storeArgs := storeFilter.GenerateSQL()
	if storeCondition != "places.name LIKE ?" {
		t.Fatalf("store condition = %q, want %q", storeCondition, "places.name LIKE ?")
	}
	if len(storeArgs) != 1 || storeArgs[0] != "%migros%" {
		t.Fatalf("store args = %v, want [%%migros%%]", storeArgs)
	}

	descFilter := db.TransactionDescriptionFilter{Description: "coca%"}
	descCondition, descArgs := descFilter.GenerateSQL()
	if descCondition != "transactions.description LIKE ?" {
		t.Fatalf("desc condition = %q, want %q", descCondition, "transactions.description LIKE ?")
	}
	if len(descArgs) != 1 || descArgs[0] != "coca%" {
		t.Fatalf("desc args = %v, want [coca%%]", descArgs)
	}
}
