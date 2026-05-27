package cmd

import (
	"testing"

	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestParseListSortOptions(t *testing.T) {
	parsed := parser.ParsedCmdLine{
		Command: "list",
		Filters: []parser.Token{
			{Kind: parser.TokenAttribute, Key: "order", Value: "datetime"},
			{Kind: parser.TokenAttribute, Key: "desc", Value: "false"},
			{Kind: parser.TokenAttribute, Key: "date", Value: "month"},
		},
	}

	filtered, options, err := parseListSortOptions(parsed)
	if err != nil {
		t.Fatalf("parseListSortOptions returned error: %v", err)
	}
	if options.Field != "datetime" {
		t.Fatalf("Field = %q, want %q", options.Field, "datetime")
	}
	if options.Desc {
		t.Fatal("Desc = true, want false")
	}
	if len(filtered.Filters) != 1 {
		t.Fatalf("len(filtered.Filters) = %d, want 1", len(filtered.Filters))
	}
	if filtered.Filters[0].Key != "date" {
		t.Fatalf("remaining filter key = %q, want %q", filtered.Filters[0].Key, "date")
	}
}

func TestParseListSortOptionsDefaultsToDesc(t *testing.T) {
	parsed := parser.ParsedCmdLine{
		Command: "list",
		Filters: []parser.Token{
			{Kind: parser.TokenAttribute, Key: "order", Value: "description"},
		},
	}

	_, options, err := parseListSortOptions(parsed)
	if err != nil {
		t.Fatalf("parseListSortOptions returned error: %v", err)
	}
	if !options.Desc {
		t.Fatal("Desc = false, want true")
	}
}

func TestParseListSortOptionsRejectsUnsupportedField(t *testing.T) {
	parsed := parser.ParsedCmdLine{
		Command: "list",
		Filters: []parser.Token{
			{Kind: parser.TokenAttribute, Key: "order", Value: "unknown"},
		},
	}

	_, _, err := parseListSortOptions(parsed)
	if err == nil {
		t.Fatal("parseListSortOptions expected error, got nil")
	}
}
