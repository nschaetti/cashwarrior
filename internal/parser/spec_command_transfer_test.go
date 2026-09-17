package parser

import (
	"testing"

	"github.com/nschaetti/cashwarrior/internal/config"
)

func TestParseCmdLine_TransferDeleteAcceptsIdentifier(t *testing.T) {
	parsed, err := ParseAndValidateCmdLine(
		[]string{"transfer", "delete", "identifier:2026.05.1"},
		config.GetDefaultConfig(),
	)
	if err != nil {
		t.Fatalf("ParseAndValidateCmdLine returned error: %v", err)
	}
	if parsed.Command != "transfer" || parsed.Subcommand != "delete" {
		t.Fatalf("parsed = (%q, %q), want (transfer, delete)", parsed.Command, parsed.Subcommand)
	}
	if len(parsed.Args) != 1 || parsed.Args[0].RawString() != "identifier:2026.05.1" {
		t.Fatalf("parsed.Args = %v, want identifier:2026.05.1", parsed.Args)
	}
}

func TestParseCmdLine_TransferDeleteRejectsMissingIdentifier(t *testing.T) {
	_, err := ParseAndValidateCmdLine([]string{"transfer", "delete"}, config.GetDefaultConfig())
	if err == nil {
		t.Fatal("ParseAndValidateCmdLine(transfer delete) expected error, got nil")
	}
}

func TestParseCmdLine_TransferDeleteRejectsBareText(t *testing.T) {
	_, err := ParseAndValidateCmdLine([]string{"transfer", "delete", "2026.05.1"}, config.GetDefaultConfig())
	if err == nil {
		t.Fatal("ParseAndValidateCmdLine(transfer delete 2026.05.1) expected error, got nil")
	}
}

func TestParseCmdLine_TransferDeleteRejectsMultipleIdentifiers(t *testing.T) {
	_, err := ParseAndValidateCmdLine(
		[]string{"transfer", "delete", "identifier:2026.05.1", "identifier:2026.05.2"},
		config.GetDefaultConfig(),
	)
	if err == nil {
		t.Fatal("ParseAndValidateCmdLine(transfer delete two identifiers) expected error, got nil")
	}
}

func TestParseCmdLine_TransferListNoArgs(t *testing.T) {
	parsed, err := ParseAndValidateCmdLine([]string{"transfer", "list"}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("ParseAndValidateCmdLine returned error: %v", err)
	}
	if parsed.Command != "transfer" || parsed.Subcommand != "list" {
		t.Fatalf("parsed = (%q, %q), want (transfer, list)", parsed.Command, parsed.Subcommand)
	}
	if len(parsed.Filters) != 0 || len(parsed.Args) != 0 {
		t.Fatalf("parsed filters=%v args=%v, want empty", parsed.Filters, parsed.Args)
	}
}

func TestParseCmdLine_TransferListRejectsArgOnRight(t *testing.T) {
	_, err := ParseAndValidateCmdLine([]string{"transfer", "list", "foo"}, config.GetDefaultConfig())
	if err == nil {
		t.Fatal("ParseAndValidateCmdLine(transfer list foo) expected error, got nil")
	}
}

func TestParseCmdLine_TransferListRejectsFiltersOnLeft(t *testing.T) {
	_, err := ParseAndValidateCmdLine([]string{"from:main", "transfer", "list"}, config.GetDefaultConfig())
	if err == nil {
		t.Fatal("ParseAndValidateCmdLine(from:main transfer list) expected error, got nil")
	}
}