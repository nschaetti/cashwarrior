package cmd

import (
	"testing"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestCreateDatetimeFilterFromPeriodTokenWeekNumber(t *testing.T) {
	filter, err := createDatetimeFilter(parser.Token{Kind: parser.TokenPeriod, Raw: "week1"}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("createDatetimeFilter returned error: %v", err)
	}

	dateFilter, ok := filter.(db.TransactionDateFilter)
	if !ok {
		t.Fatalf("filter type = %T, want db.TransactionDateFilter", filter)
	}

	now := time.Now()
	want := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())
	for want.Weekday() != time.Monday {
		want = want.AddDate(0, 0, 1)
	}

	if dateFilter.From != want.Format("2006-01-02") {
		t.Fatalf("From = %v, want %v", dateFilter.From, want.Format("2006-01-02"))
	}
}
