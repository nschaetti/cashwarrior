package parser

import (
	"testing"

	"github.com/nschaetti/cashwarrior/internal/config"
)

func TestParseCmdLine_BalanceNoFilters(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"balance"}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if parsed.Command != "balance" || parsed.Subcommand != "default" {
		t.Fatalf("parsed = (%q, %q), want (balance, default)", parsed.Command, parsed.Subcommand)
	}
	if len(parsed.Filters) != 0 || len(parsed.Args) != 0 {
		t.Fatalf("parsed filters=%v args=%v, want empty", parsed.Filters, parsed.Args)
	}
}

func TestValidateParsedCmdLine_BalanceAllowsFiltersOnLeft(t *testing.T) {
	parsed := ParsedCmdLine{
		Command:    "balance",
		Subcommand: "default",
		Filters: []Arg{
			mustArg(t, "account:main,savings"),
			mustArg(t, "date:2026-05-01..2026-05-31"),
		},
	}
	if err := ValidateParsedCmdLine(parsed); err != nil {
		t.Fatalf("ValidateParsedCmdLine(balance) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_BalanceAllowsTextShortcutFilter(t *testing.T) {
	parsed := ParsedCmdLine{
		Command:    "balance",
		Subcommand: "default",
		Filters:    []Arg{mustArg(t, "month")},
	}
	if err := ValidateParsedCmdLine(parsed); err != nil {
		t.Fatalf("ValidateParsedCmdLine(balance month) returned error: %v", err)
	}
}

func TestParseCmdLine_BalanceRejectsArgsOnRight(t *testing.T) {
	_, err := ParseAndValidateCmdLine([]string{"balance", "main"}, config.GetDefaultConfig())
	if err == nil {
		t.Fatal("ParseAndValidateCmdLine(balance main) expected error, got nil")
	}
}

func TestParseCmdLine_BalanceRejectsAttributeOnRight(t *testing.T) {
	_, err := ParseAndValidateCmdLine([]string{"balance", "account:main"}, config.GetDefaultConfig())
	if err == nil {
		t.Fatal("ParseAndValidateCmdLine(balance account:main) expected error, got nil")
	}
}

func TestParseAndValidateCmdLine_BalanceAllowsFiltersOnLeft(t *testing.T) {
	_, err := ParseAndValidateCmdLine(
		[]string{"month", "account:main,savings", "balance", "--json"},
		config.GetDefaultConfig(),
	)
	if err != nil {
		t.Fatalf("ParseAndValidateCmdLine returned error: %v", err)
	}
}

