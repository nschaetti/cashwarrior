package parser

import (
	"testing"

	"github.com/nschaetti/cashwarrior/internal/config"
)

func TestParseCmdLine_ListDefaultSubcommand(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"list"}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if parsed.Command != "list" || parsed.Subcommand != "transactions" {
		t.Fatalf("parsed = (%q, %q), want (list, transactions)", parsed.Command, parsed.Subcommand)
	}
	if len(parsed.Filters) != 0 || len(parsed.Args) != 0 {
		t.Fatalf("parsed filters=%v args=%v, want empty", parsed.Filters, parsed.Args)
	}
}

func TestParseCmdLine_ListSubcommandsAndAliases(t *testing.T) {
	tests := []struct {
		args   []string
		subcmd string
	}{
		{args: []string{"list", "transactions"}, subcmd: "transactions"},
		{args: []string{"list", "t"}, subcmd: "t"},
		{args: []string{"list", "accounts"}, subcmd: "accounts"},
		{args: []string{"list", "a"}, subcmd: "a"},
		{args: []string{"list", "groups"}, subcmd: "groups"},
		{args: []string{"list", "g"}, subcmd: "g"},
		{args: []string{"list", "tags"}, subcmd: "tags"},
		{args: []string{"list", "ta"}, subcmd: "ta"},
	}

	for _, tt := range tests {
		parsed, err := ParseCmdLine(tt.args, config.GetDefaultConfig())
		if err != nil {
			t.Fatalf("ParseCmdLine(%v) returned error: %v", tt.args, err)
		}
		if parsed.Command != "list" || parsed.Subcommand != tt.subcmd {
			t.Fatalf("ParseCmdLine(%v) = (%q, %q), want (list, %q)", tt.args, parsed.Command, parsed.Subcommand, tt.subcmd)
		}
	}
}

func TestParseAndValidateCmdLine_ListAllowsFiltersOnLeft(t *testing.T) {
	_, err := ParseAndValidateCmdLine(
		[]string{"currency:CHF", "account:main", "list", "accounts", "--json"},
		config.GetDefaultConfig(),
	)
	if err != nil {
		t.Fatalf("ParseAndValidateCmdLine returned error: %v", err)
	}
}

func TestParseCmdLine_ListRejectsArgsOnRight(t *testing.T) {
	subcommands := []string{"transactions", "accounts", "groups", "tags"}
	for _, sub := range subcommands {
		_, err := ParseAndValidateCmdLine([]string{"list", sub, "main"}, config.GetDefaultConfig())
		if err == nil {
			t.Fatalf("ParseAndValidateCmdLine(list %s main) expected error, got nil", sub)
		}
	}
}

func TestParseCmdLine_ListRejectsAttributeOnRight(t *testing.T) {
	_, err := ParseAndValidateCmdLine([]string{"list", "accounts", "currency:CHF"}, config.GetDefaultConfig())
	if err == nil {
		t.Fatal("ParseAndValidateCmdLine(list accounts currency:CHF) expected error, got nil")
	}
}
