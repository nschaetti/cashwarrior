package parser

import (
	"testing"

	"github.com/nschaetti/cashwarrior/internal/config"
)

func TestParseCmdLine_GroupsAddAcceptsTextGroupAndIdentifiers(t *testing.T) {
	parsed, err := ParseAndValidateCmdLine(
		[]string{"groups", "add", "trip", "identifier:2026.05.1", "identifier:2026.05.2"},
		config.GetDefaultConfig(),
	)
	if err != nil {
		t.Fatalf("ParseAndValidateCmdLine returned error: %v", err)
	}
	if parsed.Command != "groups" || parsed.Subcommand != "add" {
		t.Fatalf("parsed = (%q, %q), want (groups, add)", parsed.Command, parsed.Subcommand)
	}
	if len(parsed.Args) != 3 {
		t.Fatalf("parsed.Args = %v, want 3 args", parsed.Args)
	}
}

func TestParseCmdLine_GroupsAddAcceptsGroupAttribute(t *testing.T) {
	parsed, err := ParseAndValidateCmdLine(
		[]string{"groups", "add", "group:trip", "identifier:2026.05.1"},
		config.GetDefaultConfig(),
	)
	if err != nil {
		t.Fatalf("ParseAndValidateCmdLine returned error: %v", err)
	}
	if len(parsed.Args) != 2 {
		t.Fatalf("parsed.Args = %v, want 2 args", parsed.Args)
	}
}

func TestParseCmdLine_GroupsAddRejectsMissingTransactionID(t *testing.T) {
	_, err := ParseAndValidateCmdLine([]string{"groups", "add", "trip"}, config.GetDefaultConfig())
	if err == nil {
		t.Fatal("ParseAndValidateCmdLine(groups add trip) expected error, got nil")
	}
}

func TestParseCmdLine_GroupsAddRejectsMissingGroupName(t *testing.T) {
	_, err := ParseAndValidateCmdLine([]string{"groups", "add", "identifier:2026.05.1"}, config.GetDefaultConfig())
	if err == nil {
		t.Fatal("ParseAndValidateCmdLine(groups add identifier:...) expected error, got nil")
	}
}

func TestParseCmdLine_GroupsAddRejectsMultipleTextGroups(t *testing.T) {
	_, err := ParseAndValidateCmdLine(
		[]string{"groups", "add", "trip", "journey", "identifier:2026.05.1"},
		config.GetDefaultConfig(),
	)
	if err == nil {
		t.Fatal("ParseAndValidateCmdLine(groups add trip journey ...) expected error, got nil")
	}
}

func TestParseCmdLine_GroupsModifyAcceptsTextAndGroupAttribute(t *testing.T) {
	parsed, err := ParseAndValidateCmdLine(
		[]string{"groups", "modify", "trip", "group:journey"},
		config.GetDefaultConfig(),
	)
	if err != nil {
		t.Fatalf("ParseAndValidateCmdLine returned error: %v", err)
	}
	if parsed.Subcommand != "modify" || len(parsed.Args) != 2 {
		t.Fatalf("parsed = (%q, %v), want (modify, 2 args)", parsed.Subcommand, parsed.Args)
	}
}

func TestParseCmdLine_GroupsModifyRejectsSingleArg(t *testing.T) {
	_, err := ParseAndValidateCmdLine([]string{"groups", "modify", "trip"}, config.GetDefaultConfig())
	if err == nil {
		t.Fatal("ParseAndValidateCmdLine(groups modify trip) expected error, got nil")
	}
}

func TestParseCmdLine_GroupsModifyRejectsGroupNameAsBareText(t *testing.T) {
	_, err := ParseAndValidateCmdLine([]string{"groups", "modify", "trip", "journey"}, config.GetDefaultConfig())
	if err == nil {
		t.Fatal("ParseAndValidateCmdLine(groups modify trip journey) expected error, got nil")
	}
}

func TestParseCmdLine_GroupsDeleteAcceptsName(t *testing.T) {
	parsed, err := ParseAndValidateCmdLine(
		[]string{"groups", "delete", "trip"},
		config.GetDefaultConfig(),
	)
	if err != nil {
		t.Fatalf("ParseAndValidateCmdLine returned error: %v", err)
	}
	if parsed.Subcommand != "delete" || len(parsed.Args) != 1 {
		t.Fatalf("parsed = (%q, %v), want (delete, 1 arg)", parsed.Subcommand, parsed.Args)
	}
}

func TestParseCmdLine_GroupsDeleteRejectsMissingName(t *testing.T) {
	_, err := ParseAndValidateCmdLine([]string{"groups", "delete"}, config.GetDefaultConfig())
	if err == nil {
		t.Fatal("ParseAndValidateCmdLine(groups delete) expected error, got nil")
	}
}

func TestParseCmdLine_GroupsRemoveAcceptsIdentifierAndGroup(t *testing.T) {
	parsed, err := ParseAndValidateCmdLine(
		[]string{"groups", "remove", "identifier:2026.05.1", "group:trip"},
		config.GetDefaultConfig(),
	)
	if err != nil {
		t.Fatalf("ParseAndValidateCmdLine returned error: %v", err)
	}
	if parsed.Subcommand != "remove" || len(parsed.Args) != 2 {
		t.Fatalf("parsed = (%q, %v), want (remove, 2 args)", parsed.Subcommand, parsed.Args)
	}
}

func TestParseCmdLine_GroupsRemoveRejectsMissingGroup(t *testing.T) {
	_, err := ParseAndValidateCmdLine(
		[]string{"groups", "remove", "identifier:2026.05.1"},
		config.GetDefaultConfig(),
	)
	if err == nil {
		t.Fatal("ParseAndValidateCmdLine(groups remove identifier:...) expected error, got nil")
	}
}

func TestParseCmdLine_GroupsRemoveRejectsBareText(t *testing.T) {
	_, err := ParseAndValidateCmdLine(
		[]string{"groups", "remove", "2026.05.1", "group:trip"},
		config.GetDefaultConfig(),
	)
	if err == nil {
		t.Fatal("ParseAndValidateCmdLine(groups remove 2026.05.1 group:trip) expected error, got nil")
	}
}

func TestParseCmdLine_GroupsAliases(t *testing.T) {
	for _, alias := range []struct {
		subcommand string
		want       string
	}{
		{"ls", "ls"},
		{"rn", "rn"},
		{"rename", "rename"},
		{"rm", "rm"},
	} {
		args := []string{"groups", alias.subcommand}
		if alias.subcommand == "rn" || alias.subcommand == "rename" {
			args = append(args, "trip", "group:journey")
		}
		if alias.subcommand == "rm" {
			args = append(args, "trip")
		}
		parsed, err := ParseAndValidateCmdLine(args, config.GetDefaultConfig())
		if err != nil {
			t.Fatalf("ParseAndValidateCmdLine(%v) returned error: %v", args, err)
		}
		if parsed.Subcommand != alias.want {
			t.Fatalf("parsed.Subcommand = %q, want %q", parsed.Subcommand, alias.want)
		}
	}
}

func TestParseCmdLine_GroupsListRejectsFiltersOnLeft(t *testing.T) {
	_, err := ParseAndValidateCmdLine([]string{"group:alpha", "groups", "list"}, config.GetDefaultConfig())
	if err == nil {
		t.Fatal("ParseAndValidateCmdLine(group:alpha groups list) expected error, got nil")
	}
}

func TestParseCmdLine_GroupsAddRejectsFiltersOnLeft(t *testing.T) {
	_, err := ParseAndValidateCmdLine(
		[]string{"amount:-10", "groups", "add", "trip", "identifier:2026.05.1"},
		config.GetDefaultConfig(),
	)
	if err == nil {
		t.Fatal("ParseAndValidateCmdLine(amount:-10 groups add ...) expected error, got nil")
	}
}

func TestParseCmdLine_GroupsListRejectsArgOnRight(t *testing.T) {
	_, err := ParseAndValidateCmdLine([]string{"groups", "list", "foo"}, config.GetDefaultConfig())
	if err == nil {
		t.Fatal("ParseAndValidateCmdLine(groups list foo) expected error, got nil")
	}
}

func TestParseCmdLine_GroupsUnknownSubcommand(t *testing.T) {
	_, err := ParseAndValidateCmdLine([]string{"groups", "frobnicate"}, config.GetDefaultConfig())
	if err == nil {
		t.Fatal("ParseAndValidateCmdLine(groups frobnicate) expected error, got nil")
	}
}