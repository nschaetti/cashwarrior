package cmd

import (
	"testing"

	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestParseGroupArgs(t *testing.T) {
	parsed := parser.ParsedCmdLine{
		Command: "group",
		Args: []parser.Token{
			{Kind: parser.TokenText, Raw: "T2026.05.10"},
			{Kind: parser.TokenText, Raw: "T2026.05.11"},
			{Kind: parser.TokenText, Raw: "ticket_12"},
		},
	}

	transactionRefs, groupName, err := parseGroupArgs(parsed)
	if err != nil {
		t.Fatalf("parseGroupArgs returned error: %v", err)
	}
	if len(transactionRefs) != 2 {
		t.Fatalf("len(transactionRefs) = %d, want 2", len(transactionRefs))
	}
	if groupName != "ticket_12" {
		t.Fatalf("groupName = %q, want %q", groupName, "ticket_12")
	}
}

func TestParseGroupArgsRequiresTransactions(t *testing.T) {
	parsed := parser.ParsedCmdLine{
		Command: "group",
		Args:    []parser.Token{{Kind: parser.TokenText, Raw: "ticket_12"}},
	}

	_, _, err := parseGroupArgs(parsed)
	if err == nil || err.Error() != "no transaction given" {
		t.Fatalf("err = %v, want no transaction given", err)
	}
}

func TestParseGroupArgsRequiresGroup(t *testing.T) {
	parsed := parser.ParsedCmdLine{
		Command: "group",
		Args: []parser.Token{
			{Kind: parser.TokenText, Raw: "T2026.05.10"},
			{Kind: parser.TokenText, Raw: "T2026.05.11"},
		},
	}

	_, _, err := parseGroupArgs(parsed)
	if err == nil || err.Error() != "no group given" {
		t.Fatalf("err = %v, want no group given", err)
	}
}

func TestParseGroupArgsRejectsMultipleGroups(t *testing.T) {
	parsed := parser.ParsedCmdLine{
		Command: "group",
		Args: []parser.Token{
			{Kind: parser.TokenText, Raw: "T2026.05.10"},
			{Kind: parser.TokenText, Raw: "ticket_1"},
			{Kind: parser.TokenText, Raw: "ticket_2"},
		},
	}

	_, _, err := parseGroupArgs(parsed)
	if err == nil || err.Error() != "multiple groups given" {
		t.Fatalf("err = %v, want multiple groups given", err)
	}
}

func TestIsTransactionReference(t *testing.T) {
	if !isTransactionReference("T2026.05.10") {
		t.Fatal("isTransactionReference returned false, want true")
	}
	if isTransactionReference("ticket_12") {
		t.Fatal("isTransactionReference returned true, want false")
	}
}
