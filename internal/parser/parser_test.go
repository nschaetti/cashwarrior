package parser

import (
	"testing"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/domain"
)

func TestTokenKindString(t *testing.T) {
	if TokenAmount.String() != "amount" {
		t.Fatalf("TokenAmount.String() = %q, want %q", TokenAmount.String(), "amount")
	}
	if TokenKind(999).String() != "unknown" {
		t.Fatalf("TokenKind(999).String() = %q, want %q", TokenKind(999).String(), "unknown")
	}
}

func TestTokenString(t *testing.T) {
	if got := (Token{Kind: TokenAmount, Amount: -12.5}).String(); got != "<Token amount: -12.500000>" {
		t.Fatalf("amount token string = %q", got)
	}
	if got := (Token{Kind: TokenID, TransID: domain.TransactionID{Year: 2026, Month: 5, Num: 2}}).String(); got != "<Token id: 2026.05.2>" {
		t.Fatalf("id token string = %q", got)
	}
	if got := (Token{Kind: TokenAttribute, Key: "account", Value: "cash"}).String(); got != "<Token attribute: account:cash>" {
		t.Fatalf("attribute token string = %q", got)
	}
}

func TestClassifyToken_RuleOrder(t *testing.T) {
	flag := ClassifyToken("--help")
	if flag.Kind != TokenFlag || flag.Key != "help" {
		t.Fatalf("ClassifyToken(--help) = (%v,%q), want (flag,help)", flag.Kind, flag.Key)
	}

	negTag := ClassifyToken("-@rent")
	if negTag.Kind != TokenTagNegative {
		t.Fatalf("ClassifyToken(-@rent) kind = %v, want %v", negTag.Kind, TokenTagNegative)
	}

	attr := ClassifyToken("account:cash")
	if attr.Kind != TokenAttribute {
		t.Fatalf("ClassifyToken(account:cash) kind = %v, want %v", attr.Kind, TokenAttribute)
	}

	text := ClassifyToken("hello")
	if text.Kind != TokenText {
		t.Fatalf("ClassifyToken(hello) kind = %v, want %v", text.Kind, TokenText)
	}

	periodLikeText := ClassifyToken("todayx")
	if periodLikeText.Kind != TokenText {
		t.Fatalf("ClassifyToken(todayx) kind = %v, want %v", periodLikeText.Kind, TokenText)
	}
}

func TestParseCmdLine_ExtractsHelpFlagFromArgs(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"accounts", "balance", "main", "--help"})
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if len(parsed.Flags) != 1 || parsed.Flags[0].Key != "help" {
		t.Fatalf("parsed.Flags = %#v, want one help flag", parsed.Flags)
	}
	if len(parsed.Args) != 1 || parsed.Args[0].Raw != "main" {
		t.Fatalf("parsed.Args = %#v, want main", parsed.Args)
	}
}

func TestParseCmdLine_ExtractsShortHelpFlag(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"-h", "accounts"})
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if len(parsed.Flags) != 1 || parsed.Flags[0].Key != "help" {
		t.Fatalf("parsed.Flags = %#v, want one help flag", parsed.Flags)
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

func TestExtractTokens(t *testing.T) {
	tokens := ExtractTokens([]string{"today", "-12.50", "@rent"})
	if len(tokens) != 3 {
		t.Fatalf("len(tokens) = %d, want 3", len(tokens))
	}
	if tokens[0].Kind != TokenPeriod || tokens[1].Kind != TokenAmount || tokens[2].Kind != TokenTag {
		t.Fatalf("unexpected token kinds: %v, %v, %v", tokens[0].Kind, tokens[1].Kind, tokens[2].Kind)
	}
}

func TestParseCmdLine(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"today", "@food", "add", "-12.50", "coffee"})
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
	if parsed.Filters[0].Kind != TokenPeriod || parsed.Filters[1].Kind != TokenTag {
		t.Fatalf("unexpected filter kinds: %v, %v", parsed.Filters[0].Kind, parsed.Filters[1].Kind)
	}
	if parsed.Args[0].Kind != TokenAmount || parsed.Args[1].Kind != TokenText {
		t.Fatalf("unexpected args kinds: %v, %v", parsed.Args[0].Kind, parsed.Args[1].Kind)
	}
}

func TestParseCmdLine_DefaultSubcommand(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"budget"})
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
	parsed, err := ParseCmdLine([]string{"budget", "add", "account:cash", "groceries"})
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
	parsed, err := ParseCmdLine([]string{"budget", "hello"})
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if parsed.Subcommand != "list" {
		t.Fatalf("parsed.Subcommand = %q, want list", parsed.Subcommand)
	}
	if len(parsed.Args) != 1 || parsed.Args[0].Raw != "hello" {
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
		{name: "command at start", args: []string{"add", "today", "-12.50"}, filters: 0, argv: 2},
		{name: "command in middle", args: []string{"today", "add", "-12.50"}, filters: 1, argv: 1},
		{name: "command at end", args: []string{"today", "-12.50", "add"}, filters: 2, argv: 0},
	}

	for _, tt := range tests {
		parsed, err := ParseCmdLine(tt.args)
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
	_, err := ParseCmdLine([]string{"today", "-12.50"})
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
		Filters:    []Token{},
		Args:       []Token{{Raw: "-12.50", Kind: TokenAmount}, {Raw: "coffee", Kind: TokenText}},
	}
	if err := ValidateParsedCmdLine(valid); err != nil {
		t.Fatalf("ValidateParsedCmdLine(valid) returned error: %v", err)
	}

	invalid := ParsedCmdLine{
		Command:    "add",
		Subcommand: "default",
		Filters:    []Token{{Raw: "???", Kind: TokenUnknown}},
		Args:       []Token{},
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
		Filters:    []Token{{Raw: "account:cash,bank", Kind: TokenAttribute, Key: "account", Value: "cash,bank"}},
		Args:       []Token{{Raw: "date:2026/01/01-2026/01/31", Kind: TokenAttribute, Key: "date", Value: "2026/01/01-2026/01/31"}},
	}
	if err := ValidateParsedCmdLine(valid); err != nil {
		t.Fatalf("ValidateParsedCmdLine(valid budget) returned error: %v", err)
	}

	invalid := ParsedCmdLine{
		Command:    "budget",
		Subcommand: "list",
		Filters:    []Token{{Raw: "period:q1,q2", Kind: TokenAttribute, Key: "period", Value: "q1,q2"}},
	}
	err := ValidateParsedCmdLine(invalid)
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(invalid budget) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_AddCommandAllowsHistoricalSyntax(t *testing.T) {
	parsed := ParsedCmdLine{
		Command:    "add",
		Subcommand: "default",
		Args: []Token{
			{Raw: "-12.50", Kind: TokenAmount, Amount: -12.5},
			{Raw: "coffee", Kind: TokenText},
			{Raw: "@food", Kind: TokenTag},
			{Raw: "store:Coop", Kind: TokenAttribute, Key: "store", Value: "Coop"},
			{Raw: "account:cash", Kind: TokenAttribute, Key: "account", Value: "cash"},
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
		Args: []Token{
			{Raw: "-12.50", Kind: TokenAmount, Amount: -12.5},
			{Raw: "account:cash", Kind: TokenAttribute, Key: "account", Value: "cash"},
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
		Args: []Token{
			{Raw: "coffee", Kind: TokenText},
			{Raw: "account:cash", Kind: TokenAttribute, Key: "account", Value: "cash"},
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
		Filters:    []Token{{Raw: "today", Kind: TokenPeriod, Period: domain.PeriodToday}},
		Args:       []Token{{Raw: "-12.50", Kind: TokenAmount, Amount: -12.5}},
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
		Args: []Token{
			{Raw: "-12.50", Kind: TokenAmount, Amount: -12.5},
			{Raw: "from:cash", Kind: TokenAttribute, Key: "from", Value: "cash"},
		},
	}
	err := ValidateParsedCmdLine(parsed)
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(add bad attribute) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_ModifyCommandAllowsHistoricalSyntax(t *testing.T) {
	parsed := ParsedCmdLine{
		Command:    "modify",
		Subcommand: "default",
		Filters:    []Token{{Raw: "date:2026-05-27", Kind: TokenAttribute, Key: "date", Value: "2026-05-27"}},
		Args:       []Token{{Raw: "store:Coop", Kind: TokenAttribute, Key: "store", Value: "Coop"}},
	}
	if err := ValidateParsedCmdLine(parsed); err != nil {
		t.Fatalf("ValidateParsedCmdLine(modify) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_TransferCommandAllowsHistoricalSyntax(t *testing.T) {
	parsed := ParsedCmdLine{
		Command:    "transfer",
		Subcommand: "default",
		Args: []Token{
			{Raw: "+100", Kind: TokenAmount, Amount: 100},
			{Raw: "from:cash", Kind: TokenAttribute, Key: "from", Value: "cash"},
			{Raw: "to:bank", Kind: TokenAttribute, Key: "to", Value: "bank"},
			{Raw: "rent", Kind: TokenText},
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
		Args: []Token{
			{Raw: "from:cash", Kind: TokenAttribute, Key: "from", Value: "cash"},
			{Raw: "to:bank", Kind: TokenAttribute, Key: "to", Value: "bank"},
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
		Args: []Token{
			{Raw: "+100", Kind: TokenAmount, Amount: 100},
			{Raw: "from:cash", Kind: TokenAttribute, Key: "from", Value: "cash"},
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
		Args:       []Token{{Raw: "group:trip", Kind: TokenAttribute, Key: "group", Value: "trip"}},
	}
	err := ValidateParsedCmdLine(parsed)
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(group) expected error, got nil")
	}
}

func TestParseCmdLine_AccountsDefaultSubcommand(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"accounts"})
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if parsed.Command != "accounts" || parsed.Subcommand != "list" {
		t.Fatalf("parsed = (%q, %q), want (accounts, list)", parsed.Command, parsed.Subcommand)
	}
}

func TestParseCmdLine_AccountsBalanceSubcommand(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"accounts", "balance", "main"})
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if parsed.Command != "accounts" || parsed.Subcommand != "balance" {
		t.Fatalf("parsed = (%q, %q), want (accounts, balance)", parsed.Command, parsed.Subcommand)
	}
}

func TestValidateParsedCmdLine_AccountsBalanceAllowsFiltersAndName(t *testing.T) {
	parsed := ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "balance",
		Filters:    []Token{{Raw: "date:2026-05-27", Kind: TokenAttribute, Key: "date", Value: "2026-05-27"}},
		Args:       []Token{{Raw: "main", Kind: TokenText}},
	}
	if err := ValidateParsedCmdLine(parsed); err != nil {
		t.Fatalf("ValidateParsedCmdLine(accounts balance) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_AccountsBalanceAllowsAccountAttribute(t *testing.T) {
	parsed := ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "balance",
		Args:       []Token{{Raw: "account:main", Kind: TokenAttribute, Key: "account", Value: "main"}},
	}
	if err := ValidateParsedCmdLine(parsed); err != nil {
		t.Fatalf("ValidateParsedCmdLine(accounts balance account attr) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_AccountsAddAllowsNameAndCurrency(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "add",
		Args:       []Token{{Raw: "savings", Kind: TokenText}, {Raw: "currency:EUR", Kind: TokenAttribute, Key: "currency", Value: "EUR"}},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(accounts add) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_AccountsAddAllowsInitialBalance(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "add",
		Args:       []Token{{Raw: "savings", Kind: TokenText}, {Raw: "initial_balance:120.50", Kind: TokenAttribute, Key: "initial_balance", Value: "120.50"}},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(accounts add initial_balance) returned error: %v", err)
	}
}

func TestParseCmdLine_CategoriesDefaultSubcommand(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"categories"})
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
		Args:       []Token{{Raw: "travel", Kind: TokenText}, {Raw: "parent:lifestyle", Kind: TokenAttribute, Key: "parent", Value: "lifestyle"}},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(categories add) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_CategoriesModifyRequiresOneTextTarget(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "categories",
		Subcommand: "modify",
		Args:       []Token{{Raw: "parent:lifestyle", Kind: TokenAttribute, Key: "parent", Value: "lifestyle"}},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(categories modify no target) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_CategoriesDeleteAllowsCategoryAttribute(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "categories",
		Subcommand: "delete",
		Args:       []Token{{Raw: "category:travel", Kind: TokenAttribute, Key: "category", Value: "travel"}},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(categories delete attr) returned error: %v", err)
	}
}

func TestParseCmdLine_TagsDefaultSubcommand(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"tags"})
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if parsed.Command != "tags" || parsed.Subcommand != "list" {
		t.Fatalf("parsed = (%q, %q), want (tags, list)", parsed.Command, parsed.Subcommand)
	}
}

func TestParseCmdLine_GroupsDefaultSubcommand(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"groups"})
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if parsed.Subcommand != "list" {
		t.Fatalf("parsed.Subcommand = %q, want list", parsed.Subcommand)
	}
}

func TestParseCmdLine_PlacesDefaultSubcommand(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"places"})
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if parsed.Subcommand != "list" {
		t.Fatalf("parsed.Subcommand = %q, want list", parsed.Subcommand)
	}
}

func TestParseCmdLine_SummaryHasNoDefaultSubcommand(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"summary"})
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
		Filters: []Token{
			{Raw: "month", Kind: TokenPeriod, Period: domain.PeriodMonth},
			{Raw: "account:main", Kind: TokenAttribute, Key: "account", Value: "main"},
			{Raw: "identifier:2026.05.1", Kind: TokenAttribute, Key: "identifier", Value: "2026.05.1"},
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
		Filters:    []Token{{Raw: "month", Kind: TokenPeriod, Period: domain.PeriodMonth}},
		Args:       []Token{{Raw: "account:main", Kind: TokenAttribute, Key: "account", Value: "main"}},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(summary days right-side filter) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_TagsAddAllowsName(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{Command: "tags", Subcommand: "add", Args: []Token{{Raw: "travel", Kind: TokenText}}})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(tags add) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_TagsModifyRequiresNewName(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{Command: "tags", Subcommand: "modify", Args: []Token{{Raw: "travel", Kind: TokenText}}})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(tags modify no new name) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_TagsDeleteAllowsTagAttribute(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{Command: "tags", Subcommand: "delete", Args: []Token{{Raw: "tag:travel", Kind: TokenAttribute, Key: "tag", Value: "travel"}}})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(tags delete attr) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_AccountsModifyRequiresOneTextTarget(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "modify",
		Args:       []Token{{Raw: "currency:EUR", Kind: TokenAttribute, Key: "currency", Value: "EUR"}},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(accounts modify no target) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_AccountsModifyAllowsInitialBalance(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "modify",
		Args:       []Token{{Raw: "savings", Kind: TokenText}, {Raw: "initial_balance:10", Kind: TokenAttribute, Key: "initial_balance", Value: "10"}},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(accounts modify initial_balance) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_AccountsDeleteAllowsAccountAttribute(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "delete",
		Args:       []Token{{Raw: "account:main", Kind: TokenAttribute, Key: "account", Value: "main"}},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(accounts delete attr) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_AccountsRenameAllowsTwoTextArgs(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "rename",
		Args: []Token{
			{Raw: "main", Kind: TokenText},
			{Raw: "brokerage", Kind: TokenText},
		},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(accounts rename) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_AccountsRenameRequiresTwoTextArgs(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "rename",
		Args:       []Token{{Raw: "main", Kind: TokenText}},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(accounts rename one arg) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_AccountsInitialBalanceAllowsTextArgs(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "initial-balance",
		Args: []Token{
			{Raw: "main", Kind: TokenText},
			{Raw: "100", Kind: TokenText},
		},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(accounts initial-balance text args) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_AccountsInitialBalanceAllowsAttributes(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "initial-balance",
		Args: []Token{
			{Raw: "account:main", Kind: TokenAttribute, Key: "account", Value: "main"},
			{Raw: "amount:100", Kind: TokenAttribute, Key: "amount", Value: "100"},
		},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(accounts initial-balance attrs) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_AccountsBalanceRejectsAccountFilter(t *testing.T) {
	parsed := ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "balance",
		Filters:    []Token{{Raw: "account:main", Kind: TokenAttribute, Key: "account", Value: "main"}},
		Args:       []Token{{Raw: "main", Kind: TokenText}},
	}
	err := ValidateParsedCmdLine(parsed)
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(accounts balance account-filter) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_AccountsBalanceRequiresExactlyOneText(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{Command: "accounts", Subcommand: "balance"})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(accounts balance no arg) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_AccountsListRejectsArgs(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "list",
		Args:       []Token{{Raw: "main", Kind: TokenText}},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(accounts list args) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_AccountsBalanceRejectsUnexpectedAttribute(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "accounts",
		Subcommand: "balance",
		Args:       []Token{{Raw: "store:coop", Kind: TokenAttribute, Key: "store", Value: "coop"}},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(accounts balance store attr) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_FakeitCommandAllowsHistoricalSyntax(t *testing.T) {
	parsed := ParsedCmdLine{
		Command:    "fakeit",
		Subcommand: "default",
		Args: []Token{
			{Raw: "year:2026", Kind: TokenAttribute, Key: "year", Value: "2026"},
			{Raw: "month:may", Kind: TokenAttribute, Key: "month", Value: "may"},
			{Raw: "50", Kind: TokenText},
		},
	}
	if err := ValidateParsedCmdLine(parsed); err != nil {
		t.Fatalf("ValidateParsedCmdLine(fakeit) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_FakeitCommandRejectsFilters(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "fakeit",
		Subcommand: "default",
		Filters:    []Token{{Raw: "today", Kind: TokenPeriod, Period: domain.PeriodToday}},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(fakeit filters) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_FakeitRejectsMultipleTextArgs(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "fakeit",
		Subcommand: "default",
		Args:       []Token{{Raw: "10", Kind: TokenText}, {Raw: "20", Kind: TokenText}},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(fakeit multiple counts) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_ConfigAllowsAttributeOnly(t *testing.T) {
	parsed := ParsedCmdLine{
		Command:    "config",
		Subcommand: "default",
		Args:       []Token{{Raw: "gui.theme:neon-noir", Kind: TokenAttribute, Key: "gui.theme", Value: "neon-noir"}},
	}
	if err := ValidateParsedCmdLine(parsed); err != nil {
		t.Fatalf("ValidateParsedCmdLine(config) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_ConfigRejectsMultipleArgs(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "config",
		Subcommand: "default",
		Args: []Token{
			{Raw: "gui.theme:neon-noir", Kind: TokenAttribute, Key: "gui.theme", Value: "neon-noir"},
			{Raw: "gui.show_currency:true", Kind: TokenAttribute, Key: "gui.show_currency", Value: "true"},
		},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(config multiple args) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_ConfigRejectsFilters(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "config",
		Subcommand: "default",
		Filters:    []Token{{Raw: "today", Kind: TokenPeriod, Period: domain.PeriodToday}},
		Args:       []Token{{Raw: "gui.theme:neon-noir", Kind: TokenAttribute, Key: "gui.theme", Value: "neon-noir"}},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(config filters) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_ThemeAllowsZeroOrOneTextArg(t *testing.T) {
	if err := ValidateParsedCmdLine(ParsedCmdLine{Command: "theme", Subcommand: "default"}); err != nil {
		t.Fatalf("ValidateParsedCmdLine(theme no args) returned error: %v", err)
	}
	if err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "theme",
		Subcommand: "default",
		Args:       []Token{{Raw: "neon-noir", Kind: TokenText}},
	}); err != nil {
		t.Fatalf("ValidateParsedCmdLine(theme one arg) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_ThemeRejectsMultipleArgs(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "theme",
		Subcommand: "default",
		Args:       []Token{{Raw: "one", Kind: TokenText}, {Raw: "two", Kind: TokenText}},
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
		Args:       []Token{{Raw: "desc:", Kind: TokenAttributeClear, Key: "desc"}},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(modify clear desc) expected error, got nil")
	}
}

func TestValidateParsedCmdLine_ModifyAllowsClearingCategory(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "modify",
		Subcommand: "default",
		Args:       []Token{{Raw: "category:", Kind: TokenAttributeClear, Key: "category"}},
	})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(modify clear category) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_ModifyAllowsTagChanges(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command:    "modify",
		Subcommand: "default",
		Filters:    []Token{{Raw: "T2026.05.1", Kind: TokenText}},
		Args: []Token{
			{Raw: "@food", Kind: TokenTag},
			{Raw: "-@travel", Kind: TokenTagNegative},
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
		Args: []Token{
			{Raw: "-12.50", Kind: TokenAmount, Amount: -12.5},
			{Raw: "coffee", Kind: TokenText},
			{Raw: "account:cash", Kind: TokenAttribute, Key: "account", Value: "cash"},
			{Raw: "account:bank", Kind: TokenAttribute, Key: "account", Value: "bank"},
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
		Filters:    []Token{{Raw: "today", Kind: TokenPeriod, Period: domain.PeriodToday}},
		Args:       []Token{{Raw: "neon-noir", Kind: TokenText}},
	})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(theme filters) expected error, got nil")
	}
}

func TestParseCmdLine_DeleteListSubcommand(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"delete", "list"})
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

func TestValidateParsedCmdLine_PurgeRequiresTransactionID(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{Command: "purge", Subcommand: "default"})
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(purge no id) expected error, got nil")
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
	err := ValidateParsedCmdLine(ParsedCmdLine{Command: "backup", Subcommand: "default"})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(backup no args) returned error: %v", err)
	}
}

func TestValidateParsedCmdLine_BackupAllowsOutputAttribute(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{Command: "backup", Subcommand: "default", Args: []Token{{Raw: "output:/tmp/cash.db", Kind: TokenAttribute, Key: "output", Value: "/tmp/cash.db"}}})
	if err != nil {
		t.Fatalf("ValidateParsedCmdLine(backup output) returned error: %v", err)
	}
}

func TestParseAndValidateCmdLine(t *testing.T) {
	parsed, err := ParseAndValidateCmdLine([]string{"add", "-12.50", "coffee"})
	if err != nil {
		t.Fatalf("ParseAndValidateCmdLine returned error: %v", err)
	}
	if parsed.Command != "add" {
		t.Fatalf("parsed.Command = %q, want add", parsed.Command)
	}
}

func TestParseAndValidateCmdLine_AllowsHelpWithoutRequiredArgs(t *testing.T) {
	parsed, err := ParseAndValidateCmdLine([]string{"transfer", "--help"})
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
