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
		Filters: []parser.Token{
			{Raw: "account:cash,bank", Kind: parser.TokenAttribute, Key: "account", Value: "cash,bank"},
		},
		Args: []parser.Token{
			{Raw: "date:2026/01/01-2026/01/31", Kind: parser.TokenAttribute, Key: "date", Value: "2026/01/01-2026/01/31"},
			{Raw: "@planned", Kind: parser.TokenTag},
		},
	}

	output := FormatBudgetDemo(parsed)
	checks := []string{
		"budget",
		"command: budget",
		"subcommand: add",
		"<Token attribute: account:cash,bank> => list(cash,bank)",
		"<Token attribute: date:2026/01/01-2026/01/31> => range(2026/01/01-2026/01/31)",
		"<Token tag: @planned>",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("output does not contain %q:\n%s", check, output)
		}
	}
}
