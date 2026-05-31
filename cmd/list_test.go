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

func TestParseListSortOptionsAcceptsDateAlias(t *testing.T) {
	parsed := parser.ParsedCmdLine{
		Command: "list",
		Filters: []parser.Token{{Kind: parser.TokenAttribute, Key: "order", Value: "date"}},
	}

	_, options, err := parseListSortOptions(parsed)
	if err != nil {
		t.Fatalf("parseListSortOptions returned error: %v", err)
	}
	if options.Field != "date" {
		t.Fatalf("Field = %q, want %q", options.Field, "date")
	}
}

func TestParseListSortOptionsConsumesArgsAttributes(t *testing.T) {
	parsed := parser.ParsedCmdLine{
		Command: "list",
		Args: []parser.Token{
			{Kind: parser.TokenAttribute, Key: "order", Value: "date"},
			{Kind: parser.TokenAttribute, Key: "desc", Value: "false"},
			{Kind: parser.TokenAttribute, Key: "account", Value: "main"},
		},
	}

	filtered, options, err := parseListSortOptions(parsed)
	if err != nil {
		t.Fatalf("parseListSortOptions returned error: %v", err)
	}
	if options.Field != "date" {
		t.Fatalf("Field = %q, want %q", options.Field, "date")
	}
	if options.Desc {
		t.Fatal("Desc = true, want false")
	}
	if len(filtered.Args) != 1 || filtered.Args[0].Key != "account" {
		t.Fatalf("filtered.Args = %#v, want only account attr", filtered.Args)
	}
}

func TestClassifyFilterGroup(t *testing.T) {
	token := parser.Token{Kind: parser.TokenAttribute, Key: "group", Value: "ticket_0001"}
	if got := classifyFilter(token); got != FilterTypeGroup {
		t.Fatalf("classifyFilter(group) = %d, want %d", got, FilterTypeGroup)
	}
}
