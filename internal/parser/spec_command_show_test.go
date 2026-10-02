package parser

import (
	"testing"

	"github.com/nschaetti/cashwarrior/internal/config"
)

func TestParseCmdLine_ShowTransactionAcceptsIdentifierFilter(t *testing.T) {
	parsed, err := ParseAndValidateCmdLine(
		[]string{"identifier:2026.05.1", "show", "transaction"},
		config.GetDefaultConfig(),
	)
	if err != nil {
		t.Fatalf("ParseAndValidateCmdLine returned error: %v", err)
	}
	if parsed.Command != "show" || parsed.Subcommand != "transaction" {
		t.Fatalf("parsed = (%q, %q), want (show, transaction)", parsed.Command, parsed.Subcommand)
	}
	if len(parsed.Args) != 0 {
		t.Fatalf("parsed.Args = %v, want empty", parsed.Args)
	}
	if len(parsed.Filters) != 1 {
		t.Fatalf("parsed.Filters = %v, want 1 filter", parsed.Filters)
	}
	attr, ok := parsed.Filters[0].(ArgAttribute)
	if !ok || attr.Key != "identifier" || attr.Value.Raw != "2026.05.1" {
		t.Fatalf("parsed.Filters[0] = %#v, want identifier:2026.05.1", parsed.Filters[0])
	}
}

func TestParseCmdLine_ShowUsesDefaultSubcommand(t *testing.T) {
	parsed, err := ParseAndValidateCmdLine(
		[]string{"identifier:2026.05.1", "show"},
		config.GetDefaultConfig(),
	)
	if err != nil {
		t.Fatalf("ParseAndValidateCmdLine returned error: %v", err)
	}
	if parsed.Subcommand != "transaction" {
		t.Fatalf("parsed.Subcommand = %q, want transaction", parsed.Subcommand)
	}
}

func TestParseCmdLine_ShowTransactionAllowsMissingID(t *testing.T) {
	_, err := ParseAndValidateCmdLine([]string{"show", "transaction"}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("err = %v, want nil (presence checked by handler)", err)
	}
}

func TestParseCmdLine_ShowTransactionRejectsIdentifierOnRight(t *testing.T) {
	_, err := ParseAndValidateCmdLine(
		[]string{"show", "transaction", "identifier:2026.05.1"},
		config.GetDefaultConfig(),
	)
	if err == nil {
		t.Fatal("ParseAndValidateCmdLine(show transaction identifier:...) expected error, got nil")
	}
}

func TestParseCmdLine_ShowTransactionAcceptsTextID(t *testing.T) {
	for _, args := range [][]string{{"show", "transaction", "2026.05.1"}, {"show", "2026.05.1"}} {
		parsed, err := ParseAndValidateCmdLine(args, config.GetDefaultConfig())
		if err != nil {
			t.Fatal(err)
		}
		if parsed.Subcommand != "transaction" || len(parsed.Filters) != 0 || len(parsed.Args) != 1 {
			t.Fatalf("parsed = %#v", parsed)
		}
		text, ok := parsed.Args[0].(ArgText)
		if !ok || text.Text != "2026.05.1" {
			t.Fatalf("arg = %#v", parsed.Args[0])
		}
	}
}

func TestParseCmdLine_ShowTransactionRejectsMultipleTextIDs(t *testing.T) {
	if _, err := ParseAndValidateCmdLine([]string{"show", "transaction", "a", "b"}, config.GetDefaultConfig()); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCmdLine_ShowTransactionRejectsMultipleIdentifiers(t *testing.T) {
	_, err := ParseAndValidateCmdLine(
		[]string{"identifier:2026.05.1", "identifier:2026.05.2", "show", "transaction"},
		config.GetDefaultConfig(),
	)
	if err == nil {
		t.Fatal("ParseAndValidateCmdLine(two identifiers show transaction) expected error, got nil")
	}
}
