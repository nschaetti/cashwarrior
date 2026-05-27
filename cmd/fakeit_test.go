package cmd

import (
	"testing"
	"time"

	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestParseFakeitOptionsMonthAndCount(t *testing.T) {
	parsed := parser.ParsedCmdLine{
		Command: "fakeit",
		Args: []parser.Token{
			{Kind: parser.TokenAttribute, Key: "year", Value: "2026"},
			{Kind: parser.TokenAttribute, Key: "month", Value: "may"},
			{Kind: parser.TokenText, Raw: "50"},
		},
	}

	options, err := parseFakeitOptions(parsed)
	if err != nil {
		t.Fatalf("parseFakeitOptions returned error: %v", err)
	}

	if options.Year != 2026 {
		t.Fatalf("Year = %d, want 2026", options.Year)
	}
	if options.Month != time.May {
		t.Fatalf("Month = %v, want %v", options.Month, time.May)
	}
	if options.Count != 50 {
		t.Fatalf("Count = %d, want 50", options.Count)
	}
}

func TestResolveFakeitDateRangeForMonth(t *testing.T) {
	from, to := resolveFakeitDateRange(fakeitOptions{Year: 2026, Month: time.May})

	if from.Year() != 2026 || from.Month() != time.May || from.Day() != 1 {
		t.Fatalf("from = %v, want 2026-05-01", from)
	}
	if to.Year() != 2026 || to.Month() != time.May || to.Day() != 31 {
		t.Fatalf("to = %v, want 2026-05-31", to)
	}
}
