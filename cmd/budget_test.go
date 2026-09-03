package cmd

import (
	"strings"
	"testing"

	"github.com/nschaetti/cashwarrior/internal/parser"
)

func TestFormatBudgetDemo(t *testing.T) {
	parsed := parser.ParsedCmdLine{
		Command:    "budget",
		Subcommand: "add",
		Filters: []parser.Arg{
			testArg(t, "account:cash,bank"),
		},
		Args: []parser.Arg{
			testArg(t, "date:2026-01-01..2026-01-31"),
			testArg(t, "@planned"),
		},
	}

	output := FormatBudgetDemo(parsed)
	checks := []string{
		"budget",
		"command: budget",
		"subcommand: add",
		"account:cash,bank => list(string(cash),string(bank))",
		"date:2026-01-01..2026-01-31 => range(date(2026-01-01),date(2026-01-31))",
		"@planned",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("output does not contain %q:\n%s", check, output)
		}
	}
}
