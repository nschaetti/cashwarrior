package parser

import (
	"testing"

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
		Command: "add",
		Filters: []Token{{Raw: "today", Kind: TokenPeriod}},
		Args:    []Token{{Raw: "-12.50", Kind: TokenAmount}},
	}
	if err := ValidateParsedCmdLine(valid); err != nil {
		t.Fatalf("ValidateParsedCmdLine(valid) returned error: %v", err)
	}

	invalid := ParsedCmdLine{
		Command: "add",
		Filters: []Token{{Raw: "???", Kind: TokenUnknown}},
		Args:    []Token{},
	}
	err := ValidateParsedCmdLine(invalid)
	if err == nil {
		t.Fatal("ValidateParsedCmdLine(invalid) expected error, got nil")
	}
	if !IsParseErrorCode(err, ParseErrorUnknownToken) {
		t.Fatalf("ValidateParsedCmdLine(invalid) code mismatch: %v", err)
	}
}

func TestParseAndValidateCmdLine(t *testing.T) {
	parsed, err := ParseAndValidateCmdLine([]string{"today", "add", "-12.50"})
	if err != nil {
		t.Fatalf("ParseAndValidateCmdLine returned error: %v", err)
	}
	if parsed.Command != "add" {
		t.Fatalf("parsed.Command = %q, want add", parsed.Command)
	}
}

func TestParseCmdLine_SumCommandIsRecognized(t *testing.T) {
	parsed, err := ParseCmdLine([]string{"sum"})
	if err != nil {
		t.Fatalf("ParseCmdLine returned error: %v", err)
	}
	if parsed.Command != "sum" {
		t.Fatalf("parsed.Command = %q, want %q", parsed.Command, "sum")
	}
}

func TestValidateParsedCmdLine_InitHasNoArgs(t *testing.T) {
	err := ValidateParsedCmdLine(ParsedCmdLine{
		Command: "init",
		Args:    []Token{{Raw: "extra", Kind: TokenText}},
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !IsParseErrorCode(err, ParseErrorInvalidInput) {
		t.Fatalf("expected ParseErrorInvalidInput, got %v", err)
	}
}
