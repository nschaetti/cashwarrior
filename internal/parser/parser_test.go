package parser

import (
	"testing"

	"github.com/nschaetti/cashwarrior/internal/config"
)

func mustArg(t *testing.T, raw string) Arg {
	t.Helper()
	args, err := ExtractArgs([]string{raw}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("ExtractArgs(%q) returned error: %v", raw, err)
	}
	return args[0]
}

func TestArgKindString(t *testing.T) {
	if ArgKindAttribute.String() != "attribute" {
		t.Fatalf("ArgKindAttribute.String() = %q, want %q", ArgKindAttribute.String(), "attribute")
	}
	if ArgKind(999).String() != "unknown" {
		t.Fatalf("ArgKind(999).String() = %q, want %q", ArgKind(999).String(), "unknown")
	}
}

func TestArgString(t *testing.T) {
	if got := (ArgText{Raw: "coffee", Text: "coffee"}).String(); got != "argtext(coffee)" {
		t.Fatalf("text arg string = %q", got)
	}
	if got := (ArgTag{Raw: "@food", Tag: "food"}).String(); got != "argtag(food)" {
		t.Fatalf("tag arg string = %q", got)
	}
	if got := mustArg(t, "account:cash").(ArgAttribute).String(); got != "argattr(account=single(string(cash)))" {
		t.Fatalf("attribute arg string = %q", got)
	}
}

func TestClassifyArg_RuleOrder(t *testing.T) {
	if got := ClassifyArg("--help"); got != ArgKindFlag {
		t.Fatalf("ClassifyArg(--help) = %v, want flag", got)
	}
	if got := ClassifyArg("-@rent"); got != ArgKindTagNegative {
		t.Fatalf("ClassifyArg(-@rent) = %v, want %v", got, ArgKindTagNegative)
	}
	if got := ClassifyArg("account:cash"); got != ArgKindAttribute {
		t.Fatalf("ClassifyArg(account:cash) = %v, want %v", got, ArgKindAttribute)
	}
	if got := ClassifyArg("hello"); got != ArgKindText {
		t.Fatalf("ClassifyArg(hello) = %v, want %v", got, ArgKindText)
	}
	if got := ClassifyArg("todayx"); got != ArgKindText {
		t.Fatalf("ClassifyArg(todayx) = %v, want %v", got, ArgKindText)
	}
}

func TestParseCmdLine_ExtractsHelpFlagFromArgs(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"tags", "add", "coffee", "--help"}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if len(parsed.Flags) != 1 || parsed.Flags[0].(ArgFlag).Key != "help" {
		t.Fatalf("parsed.Flags = %#v, want one help flag", parsed.Flags)
	}
	if len(parsed.Args) != 1 || parsed.Args[0].RawString() != "coffee" {
		t.Fatalf("parsed.Args = %#v, want coffee", parsed.Args)
	}
}

func TestParseCmdLine_ExtractsShortHelpFlag(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"--help", "accounts"}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if len(parsed.Flags) != 1 || parsed.Flags[0].(ArgFlag).Key != "help" {
		t.Fatalf("parsed.Flags = %#v, want one help flag", parsed.Flags)
	}
}

func TestParseCmdLine_ExtractsOutputFormatFlag(t *testing.T) {
	for _, args := range [][]string{
		{"list", "--format", "json"},
		{"list", "--format=json"},
	} {
		parsed, err := ParseCmdLine(args, config.GetDefaultConfig())
		if err != nil {
			t.Fatalf("ParseCmdLine(%v) returned error: %v", args, err)
		}
		value, ok := parsed.GetFlagString("format")
		if !ok || value != "json" {
			t.Fatalf("format flag from %v = (%q, %t), want (json, true)", args, value, ok)
		}
		if len(parsed.Args) != 0 {
			t.Fatalf("parsed.Args from %v = %#v, want no format value argument", args, parsed.Args)
		}
	}
}

func TestParseCmdLine_ExtractsJSONAndYesFlags(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"--json", "list", "--yes"}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if !parsed.HasFlag("json") || !parsed.HasFlag("yes") {
		t.Fatalf("parsed flags = %#v, want json and yes", parsed.Flags)
	}
}

func TestParseCmdLine_RejectsMissingFormatValue(t *testing.T) {
	_, err := ParseCmdLine([]string{"list", "--format"}, config.GetDefaultConfig())
	if err == nil || err.Code != ParseErrorInvalidFlagValue {
		t.Fatalf("ParseCmdLine returned error = %v, want invalid flag value", err)
	}
}

func TestParseArgAttribute_CanonicalizesIdentifierAliases(t *testing.T) {
	for _, raw := range []string{"identifier:2026.05.1", "id:2026.05.1", "T:2026.05.1"} {
		arg, err := ParseArgAttribute(raw, config.Config{})
		if err != nil {
			t.Fatalf("ParseArgAttribute(%q) returned error: %v", raw, err)
		}

		attr, ok := arg.(ArgAttribute)
		if !ok {
			t.Fatalf("ParseArgAttribute(%q) returned %T, want ArgAttribute", raw, arg)
		}
		if attr.Key != "identifier" {
			t.Fatalf("ParseArgAttribute(%q) key = %q, want identifier", raw, attr.Key)
		}
		if attr.Raw != raw {
			t.Fatalf("ParseArgAttribute(%q) raw = %q, want %q", raw, attr.Raw, raw)
		}
	}
}

func TestFindCommand(t *testing.T) {
	cmd, idx, err := FindCommand([]string{"today", "add", "-12.50"})
	if err != nil {
		t.Fatalf("FindCommand returned error: %v", err)
	}
	if cmd != "add" || idx != 1 {
		t.Fatalf("FindCommand = (%q, %d), want (%q, %d)", cmd, idx, "add", 1)
	}
}

func TestFindCommand_NoCommand(t *testing.T) {
	_, _, err := FindCommand([]string{"today", "-12.50"})
	if err == nil {
		t.Fatal("FindCommand expected error, got nil")
	}
}

func TestExtractArgs(t *testing.T) {
	args, err := ExtractArgs([]string{"date:today", "amount:-12.50", "@rent"}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("ExtractArgs returned error: %v", err)
	}
	if len(args) != 3 {
		t.Fatalf("len(args) = %d, want 3", len(args))
	}
	if args[0].ArgKind() != ArgKindAttribute || args[1].ArgKind() != ArgKindAttribute || args[2].ArgKind() != ArgKindTag {
		t.Fatalf("unexpected arg kinds: %v, %v, %v", args[0].ArgKind(), args[1].ArgKind(), args[2].ArgKind())
	}
}

func TestParseCmdLine(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"date:today", "@food", "add", "amount:-12.50", "coffee"}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}

	if parsed.Command != "add" {
		t.Fatalf("parsed.Command = %q, want %q", parsed.Command, "add")
	}
	if parsed.Subcommand != "default" {
		t.Fatalf("parsed.Subcommand = %q, want %q", parsed.Subcommand, "default")
	}
	if len(parsed.Filters) != 2 || len(parsed.Args) != 2 {
		t.Fatalf("unexpected filters/args lengths: %d/%d", len(parsed.Filters), len(parsed.Args))
	}
	if parsed.Filters[0].ArgKind() != ArgKindAttribute || parsed.Filters[1].ArgKind() != ArgKindTag {
		t.Fatalf("unexpected filter kinds: %v, %v", parsed.Filters[0].ArgKind(), parsed.Filters[1].ArgKind())
	}
	if parsed.Args[0].ArgKind() != ArgKindAttribute || parsed.Args[1].ArgKind() != ArgKindText {
		t.Fatalf("unexpected args kinds: %v, %v", parsed.Args[0].ArgKind(), parsed.Args[1].ArgKind())
	}
}

func TestParseCmdLine_DefaultSubcommand(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"budget"}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if parsed.Command != "budget" {
		t.Fatalf("parsed.Command = %q, want budget", parsed.Command)
	}
	if parsed.Subcommand != "list" {
		t.Fatalf("parsed.Subcommand = %q, want list", parsed.Subcommand)
	}
}

func TestParseCmdLine_ExplicitSubcommand(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"budget", "add", "account:cash", "groceries"}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if parsed.Subcommand != "add" {
		t.Fatalf("parsed.Subcommand = %q, want add", parsed.Subcommand)
	}
	if len(parsed.Args) != 2 {
		t.Fatalf("len(args) = %d, want 2", len(parsed.Args))
	}
}

func TestParseCmdLine_NonSubcommandStaysArgument(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"budget", "hello"}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if parsed.Subcommand != "list" {
		t.Fatalf("parsed.Subcommand = %q, want list", parsed.Subcommand)
	}
	if len(parsed.Args) != 1 || parsed.Args[0].RawString() != "hello" {
		t.Fatalf("args = %#v, want hello text token", parsed.Args)
	}
}

func TestParseCmdLine_CommandAnywhere(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		filters int
		argv    int
	}{
		{name: "command at start", args: []string{"add", "date:today", "amount:-12.50"}, filters: 0, argv: 2},
		{name: "command in middle", args: []string{"date:today", "add", "amount:-12.50"}, filters: 1, argv: 1},
		{name: "command at end", args: []string{"date:today", "amount:-12.50", "add"}, filters: 2, argv: 0},
	}

	for _, tt := range tests {
		parsed, err := ParseCmdLine(tt.args, config.GetDefaultConfig())
		if err != nil {
			t.Fatalf("%s: ParseCmdLine returned error: %v", tt.name, err)
		}
		if parsed.Command != "add" {
			t.Fatalf("%s: parsed.Command = %q, want add", tt.name, parsed.Command)
		}
		if len(parsed.Filters) != tt.filters {
			t.Fatalf("%s: len(filters) = %d, want %d", tt.name, len(parsed.Filters), tt.filters)
		}
		if len(parsed.Args) != tt.argv {
			t.Fatalf("%s: len(args) = %d, want %d", tt.name, len(parsed.Args), tt.argv)
		}
	}
}

func TestParseCmdLine_NoCommand(t *testing.T) {
	_, err := ParseCmdLine([]string{"date:today", "amount:-12.50"}, config.GetDefaultConfig())
	if err == nil {
		t.Fatal("ParseCmdLine expected error, got nil")
	}
	if err.Code != ParseErrorNoCommand {
		t.Fatalf("ParseCmdLine error code = %q, want %q", err.Code, ParseErrorNoCommand)
	}
}

func TestParseAmount(t *testing.T) {
	if got := ParseAmount("-12.50"); got != float32(-12.5) {
		t.Fatalf("ParseAmount(-12.50) = %v, want %v", got, float32(-12.5))
	}
	if got := ParseAmount("10.00"); got != 0 {
		t.Fatalf("ParseAmount(10.00) = %v, want 0", got)
	}
}

func TestValidateParsedCmdLine(t *testing.T) {
	valid := ParsedCmdLine{
		Command:    "add",
		Subcommand: "default",
		Filters:    []Arg{},
		Args:       []Arg{mustArg(t, "amount:"+"-12.50"), mustArg(t, "coffee"), mustArg(t, "store:coop")},
	}
	if err := ValidateParsedCmdLine(valid); err != nil {
		t.Fatalf("ValidateParsedCmdLine(valid) returned error: %v", err)
	}

	invalid := ParsedCmdLine{
		Command:    "add",
		Subcommand: "default",
		Filters:    []Arg{mustArg(t, "date:today")},
		Args:       []Arg{mustArg(t, "amount:-12.50"), mustArg(t, "coffee"), mustArg(t, "store:coop")},
	}
	err := ValidateParsedCmdLine(invalid)
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(invalid) expected error, got nil")
	}
	if !IsParseErrorCode(err, ParseErrorUnknownToken) {
		t.Fatalf("ValidateParsedCmdLine(invalid) code mismatch: %v", err)
	}
}

func TestValidateParsedCmdLine_BudgetAttributeShapes(t *testing.T) {
	valid := ParsedCmdLine{
		Command:    "budget",
		Subcommand: "list",
		Filters:    []Arg{mustArg(t, "account:cash,bank")},
		Args:       []Arg{mustArg(t, "date:2026-01-01..2026-01-31")},
	}
	if err := ValidateParsedCmdLine(valid); err != nil {
		t.Fatalf("ValidateParsedCmdLine(valid budget) returned error: %v", err)
	}

	invalid := ParsedCmdLine{
		Command:    "budget",
		Subcommand: "list",
		Filters:    []Arg{mustArg(t, "amount:1,2")},
	}
	err := ValidateParsedCmdLine(invalid)
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(invalid budget) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_AddCommandAllowsCurrentSyntax(t *testing.T) {
	parsed := ParsedCmdLine{
		Command:    "add",
		Subcommand: "default",
		Args: []Arg{
			mustArg(t, "amount:"+"-12.50"),
			mustArg(t, "coffee"),
			mustArg(t, "@food"),
			mustArg(t, "store:Coop"),
			mustArg(t, "account:cash"),
		},
	}
	if err := ValidateParsedCmdLine(parsed); err != nil {
		t.Fatalf("ValidateParsedCmdLine(add) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_AddCommandRequiresDescriptionText(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "add",
		Subcommand: "default",
		Args: []Arg{
			mustArg(t, "amount:"+"-12.50"),
			mustArg(t, "account:cash"),
		},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(add no description) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_AddCommandRequiresAmount(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "add",
		Subcommand: "default",
		Args: []Arg{
			mustArg(t, "coffee"),
			mustArg(t, "account:cash"),
		},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(add no amount) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_AddCommandRejectsFilters(t *testing.T) {
	parsed := ParsedCmdLine{
		Command:    "add",
		Subcommand: "default",
		Filters:    []Arg{mustArg(t, "date:"+"today")},
		Args:       []Arg{mustArg(t, "amount:"+"-12.50")},
	}
	err := ValidateParsedCmdLine(parsed)
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(add filters) expected error, got nil")
	}
	if err.Code != ParseErrorUnknownToken {
		t.Fatalf("error code = %q, want %q", err.Code, ParseErrorUnknownToken)
	}
}

func TestValidateParsedCmdLine_AddCommandRejectsUnsupportedAttribute(t *testing.T) {
	parsed := ParsedCmdLine{
		Command:    "add",
		Subcommand: "default",
		Args: []Arg{
			mustArg(t, "amount:"+"-12.50"),
			mustArg(t, "from:cash"),
		},
	}
	err := ValidateParsedCmdLine(parsed)
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(add bad attribute) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_ModifyCommandAllowsCurrentSyntax(t *testing.T) {
	parsed := ParsedCmdLine{
		Command:    "modify",
		Subcommand: "transactions",
		Filters:    []Arg{mustArg(t, "date:2026-05-27")},
		Args:       []Arg{mustArg(t, "store:Coop")},
	}
	if err := ValidateParsedCmdLine(parsed); err != nil {
		t.Fatalf("ValidateParsedCmdLine(modify) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_TransferCommandAllowsCurrentSyntax(t *testing.T) {
	parsed := ParsedCmdLine{
		Command:    "transfer",
		Subcommand: "add",
		Args: []Arg{
			mustArg(t, "amount:"+"+100"),
			mustArg(t, "from:cash"),
			mustArg(t, "to:bank"),
			mustArg(t, "rent"),
		},
	}
	if err := ValidateParsedCmdLine(parsed); err != nil {
		t.Fatalf("ValidateParsedCmdLine(transfer) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_TransferRequiresExactlyOneAmount(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "transfer",
		Subcommand: "default",
		Args: []Arg{
			mustArg(t, "from:cash"),
			mustArg(t, "to:bank"),
		},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(transfer no amount) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_TransferRequiresFromAndTo(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "transfer",
		Subcommand: "default",
		Args: []Arg{
			mustArg(t, "amount:"+"+100"),
			mustArg(t, "from:cash"),
		},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(transfer missing to) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_GroupCommandRejectsNonTextArgs(t *testing.T) {
	parsed := ParsedCmdLine{
		Command:    "group",
		Subcommand: "default",
		Args:       []Arg{mustArg(t, "group:trip")},
	}
	err := ValidateParsedCmdLine(parsed)
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(group) expected error, got nil")
	}
}

func TestParseCmdLine_AccountsDefaultSubcommand(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"accounts"}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if parsed.Command != "accounts" || parsed.Subcommand != "list" {
		t.Fatalf("parsed = (%q, %q), want (accounts, list)", parsed.Command, parsed.Subcommand)
	}
}

func TestValidateParsedCmdLine_AccountsAddAllowsNameAndCurrency(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "add",
		Args:       []Arg{mustArg(t, "savings"), mustArg(t, "currency:EUR")},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(accounts add) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_AccountsAddAllowsInitialBalance(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "add",
		Args:       []Arg{mustArg(t, "savings"), mustArg(t, "initial-balance:120.50")},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(accounts add initial_balance) returned error: %v", err)
	}
}

func TestParseCmdLine_CategoriesDefaultSubcommand(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"categories"}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if parsed.Command != "categories" || parsed.Subcommand != "list" {
		t.Fatalf("parsed = (%q, %q), want (categories, list)", parsed.Command, parsed.Subcommand)
	}
}

func TestValidateParsedCmdLine_CategoriesAddAllowsNameAndParent(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "categories",
		Subcommand: "add",
		Args:       []Arg{mustArg(t, "travel"), mustArg(t, "parent:lifestyle")},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(categories add) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_CategoriesModifyRequiresOneTextTarget(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "categories",
		Subcommand: "modify",
		Args:       []Arg{mustArg(t, "parent:lifestyle")},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(categories modify no target) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_CategoriesDeleteAllowsCategoryAttribute(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "categories",
		Subcommand: "delete",
		Args:       []Arg{mustArg(t, "category:travel")},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(categories delete attr) returned error: %v", err)
	}
}

func TestParseCmdLine_TagsDefaultSubcommand(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"tags"}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if parsed.Command != "tags" || parsed.Subcommand != "list" {
		t.Fatalf("parsed = (%q, %q), want (tags, list)", parsed.Command, parsed.Subcommand)
	}
}

func TestParseCmdLine_GroupsDefaultSubcommand(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"groups"}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if parsed.Subcommand != "list" {
		t.Fatalf("parsed.Subcommand = %q, want list", parsed.Subcommand)
	}
}

func TestParseCmdLine_StoresDefaultSubcommand(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"stores"}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if parsed.Subcommand != "list" {
		t.Fatalf("parsed.Subcommand = %q, want list", parsed.Subcommand)
	}
}

func TestParseCmdLine_SummaryHasNoDefaultSubcommand(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"summary"}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if parsed.Subcommand != "" {
		t.Fatalf("parsed.Subcommand = %q, want empty", parsed.Subcommand)
	}
}

func TestValidateParsedCmdLine_SummaryRequiresSubcommand(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{Command: "summary", Subcommand: ""})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(summary no subcommand) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_SummaryDaysAllowsTransactionFilters(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "summary",
		Subcommand: "days",
		Filters: []Arg{
			mustArg(t, "date:"+"month"),
			mustArg(t, "account:main"),
			mustArg(t, "identifier:2026.05.1"),
		},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(summary days) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_SummaryDaysAllowsRightSideFilters(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "summary",
		Subcommand: "days",
		Filters:    []Arg{mustArg(t, "date:"+"month")},
		Args:       []Arg{mustArg(t, "account:main")},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(summary days right-side filter) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_TagsAddAllowsName(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{Command: "tags", Subcommand: "add", Args: []Arg{mustArg(t, "travel")}})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(tags add) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_TagsModifyRequiresNewName(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{Command: "tags", Subcommand: "modify", Args: []Arg{mustArg(t, "travel")}})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(tags modify no new name) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_TagsDeleteAllowsTagAttribute(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{Command: "tags", Subcommand: "delete", Args: []Arg{mustArg(t, "tag:travel")}})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(tags delete attr) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_AccountsModifyRequiresOneTextTarget(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "modify",
		Args:       []Arg{mustArg(t, "currency:EUR")},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(accounts modify no target) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_AccountsModifyAllowsInitialBalance(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "modify",
		Args:       []Arg{mustArg(t, "savings"), mustArg(t, "initial-balance:10")},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(accounts modify initial_balance) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_AccountsDeleteAllowsAccountAttribute(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "delete",
		Filters:    []Arg{mustArg(t, "account:main")},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(accounts delete attr) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_AccountsRenameAllowsNewName(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "rename",
		Filters:    []Arg{mustArg(t, "account:main")},
		Args:       []Arg{mustArg(t, "brokerage")},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(accounts rename) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_AccountsRenameRequiresNewName(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "rename",
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(accounts rename no name) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_AccountsInitialBalanceAllowsAttributes(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "initial-balance",
		Filters:    []Arg{mustArg(t, "account:main")},
		Args:       []Arg{mustArg(t, "amount:100")},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(accounts initial-balance attrs) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_AccountsListRejectsArgs(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "list",
		Args:       []Arg{mustArg(t, "main")},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(accounts list args) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_FakeitCommandAllowsCurrentSyntax(t *testing.T) {
	parsed := ParsedCmdLine{
		Command:    "fakeit",
		Subcommand: "transactions",
		Args: []Arg{
			mustArg(t, "year:2026"),
			mustArg(t, "month:may"),
			mustArg(t, "50"),
		},
	}
	if err := ValidateParsedCmdLine(parsed); err != nil {
		t.Fatalf("ValidateParsedCmdLine(fakeit) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_FakeitCommandRejectsFilters(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "fakeit",
		Subcommand: "transactions",
		Filters:    []Arg{mustArg(t, "date:"+"today")},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(fakeit filters) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_FakeitRejectsMultipleTextArgs(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "fakeit",
		Subcommand: "transactions",
		Args:       []Arg{mustArg(t, "10"), mustArg(t, "20")},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(fakeit multiple counts) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_ConfigAllowsSetArguments(t *testing.T) {
	parsed := ParsedCmdLine{
		Command:    "config",
		Subcommand: "set",
		Args:       []Arg{mustArg(t, "gui.theme"), mustArg(t, "neon-noir")},
	}
	if err := ValidateParsedCmdLine(parsed); err != nil {
		t.Fatalf("ValidateParsedCmdLine(config) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_ConfigRejectsMultipleArgs(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "config",
		Subcommand: "set",
		Args: []Arg{
			mustArg(t, "gui.theme"),
			mustArg(t, "neon-noir"),
			mustArg(t, "extra"),
		},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(config multiple args) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_ConfigRejectsFilters(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "config",
		Subcommand: "set",
		Filters:    []Arg{mustArg(t, "date:"+"today")},
		Args:       []Arg{mustArg(t, "gui.theme"), mustArg(t, "neon-noir")},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(config filters) expected error, got nil")
	}
}

func TestParseAndValidateConfigSetGetAndLegacySyntax(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantSub   string
		wantCount int
	}{
		{name: "set", args: []string{"config", "set", "gui.date_format", "2006-01-02"}, wantSub: "set", wantCount: 2},
		{name: "get", args: []string{"config", "get", "gui.date_format"}, wantSub: "get", wantCount: 1},
		{name: "legacy", args: []string{"config", "database:/data/cashwarrior/cash.db"}, wantSub: "set", wantCount: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := ParseAndValidateCmdLine(test.args, config.GetDefaultConfig())
			if err != nil {
				t.Fatalf("ParseAndValidateCmdLine returned error: %v", err)
			}
			if parsed.Subcommand != test.wantSub || len(parsed.Args) != test.wantCount {
				t.Fatalf("parsed = (%q, %d args), want (%q, %d args)", parsed.Subcommand, len(parsed.Args), test.wantSub, test.wantCount)
			}
		})
	}
}

func TestValidateParsedCmdLine_ThemeAllowsZeroOrOneTextArg(t *testing.T) {
	if err := ValidateParsedCmdLine(ParsedCmdLine{Command: "theme", Subcommand: "default"}); err != nil {
		t.Fatalf("ValidateParsedCmdLine(theme no args) returned error: %v", err)
	}
	if err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "theme",
		Subcommand: "default",
		Args:       []Arg{mustArg(t, "neon-noir")},
	}); err != nil {
		t.Fatalf("ValidateParsedCmdLine(theme one arg) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_ThemeRejectsMultipleArgs(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "theme",
		Subcommand: "default",
		Args:       []Arg{mustArg(t, "one"), mustArg(t, "two")},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(theme multiple args) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_ModifyRequiresAtLeastOneArg(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{Command: "modify", Subcommand: "default"})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(modify no args) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_ModifyRejectsClearingNonClearableAttribute(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "modify",
		Subcommand: "default",
		Args:       []Arg{mustArg(t, "desc:")},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(modify clear desc) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_ModifyAllowsClearingCategory(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "modify",
		Subcommand: "transactions",
		Args:       []Arg{mustArg(t, "category:")},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(modify clear category) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_ModifyAllowsTagChanges(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "modify",
		Subcommand: "transactions",
		Filters:    []Arg{mustArg(t, "T2026.05.1")},
		Args: []Arg{
			mustArg(t, "@food"),
			mustArg(t, "-@travel"),
		},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(modify tag changes) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_AddRejectsDuplicateSingletonAttribute(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "add",
		Subcommand: "default",
		Args: []Arg{
			mustArg(t, "amount:"+"-12.50"),
			mustArg(t, "coffee"),
			mustArg(t, "account:cash"),
			mustArg(t, "account:bank"),
		},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(add duplicate account) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_ThemeRejectsFilters(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "theme",
		Subcommand: "default",
		Filters:    []Arg{mustArg(t, "date:"+"today")},
		Args:       []Arg{mustArg(t, "neon-noir")},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(theme filters) expected error, got nil")
	}
}

func TestParseCmdLine_DeleteListSubcommand(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"delete", "list"}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if parsed.Command != "delete" || parsed.Subcommand != "list" {
		t.Fatalf("parsed = (%q, %q), want (delete, list)", parsed.Command, parsed.Subcommand)
	}
}

func TestValidateParsedCmdLine_DeleteRequiresTransactionID(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{Command: "delete", Subcommand: "default"})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(delete no id) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_RestoreRequiresTransactionID(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{Command: "restore", Subcommand: "default"})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(restore no id) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_PurgeRejectsArgOnRight(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{Command: "purge", Subcommand: "default", Args: []Arg{mustArg(t, "main")}})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(purge with arg) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_ShowRequiresTransactionID(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{Command: "show", Subcommand: "default"})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(show no id) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_ImportRequiresPath(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{Command: "import", Subcommand: "default"})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(import no path) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_BackupAllowsNoArgs(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{Command: "backup", Subcommand: "now"})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(backup no args) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_BackupAllowsOutputAttribute(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{Command: "backup", Subcommand: "now", Args: []Arg{mustArg(t, "output:/tmp/cash.db")}})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(backup output) returned error: %v", err)
	}
}

func TestParseAndValidateCmdLine(t *testing.T) {
	parsed, err := ParseAndValidateCmdLine([]string{"add", "amount:-12.50", "coffee", "store:coop"}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("ParseAndValidateCmdLine returned error: %v", err)
	}
	if parsed.Command != "add" {
		t.Fatalf("parsed.Command = %q, want add", parsed.Command)
	}
}

func TestParseAndValidateCmdLine_AllowsHelpWithoutRequiredArgs(t *testing.T) {
	parsed, err := ParseAndValidateCmdLine([]string{"transfer", "--help"}, config.GetDefaultConfig())
	if err != nil {
		t.Fatalf("ParseAndValidateCmdLine returned error: %v", err)
	}
	if parsed.Command != "transfer" {
		t.Fatalf("parsed.Command = %q, want transfer", parsed.Command)
	}
	if !parsed.HasFlag("help") {
		t.Fatal("expected help flag to be set")
	}
}
