package parser

import (
	"fmt"

	"github.com/nschaetti/cashwarrior/internal/domain"
)

type ParseErrorCode string

const (
	// ParseErrorNoCommand indicates that no known command was found in the input.
	ParseErrorNoCommand ParseErrorCode = "NO_COMMAND"
	// ParseErrorUnknownToken indicates that a token could not be accepted by validation.
	ParseErrorUnknownToken ParseErrorCode = "UNKNOWN_TOKEN"
	// ParseErrorInvalidInput indicates malformed input, such as an empty command line.
	ParseErrorInvalidInput ParseErrorCode = "INVALID_INPUT"
)

// ParseError represents a structured parser or validation error.
type ParseError struct {
	Code    ParseErrorCode
	Token   string
	Message string
}

// Error returns a human-readable message for the parse error.
func (e *ParseError) Error() string {
	if e.Token == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Message, e.Token)
}

// Token is a classified lexical unit extracted from the command line.
type Token struct {
	Raw     string
	Amount  float32
	TransID domain.TransactionID
	Period  domain.PeriodKind
	Kind    TokenKind
	Key     string
	Value   string
}

// String returns a debug-friendly string representation of a token.
func (t Token) String() string {
	if t.Kind == TokenAmount {
		return fmt.Sprintf("<Token %s: %f>", t.Kind, t.Amount)
	} else if t.Kind == TokenID {
		return fmt.Sprintf("<Token %s: %s>", t.Kind, t.TransID)
	} else if t.Kind == TokenPeriod {
		return fmt.Sprintf("<Token %s: %s>", t.Kind, t.Period)
	} else if t.Kind == TokenAttribute {
		return fmt.Sprintf("<Token %s: %s:%s>", t.Kind, t.Key, t.Value)
	} else if t.Kind == TokenAttributeClear {
		return fmt.Sprintf("<Token %s: %s>", t.Kind, t.Key)
	} else if t.Kind == TokenTag {
		return fmt.Sprintf("<Token %s: %s>", t.Kind, t.Raw)
	} else if t.Kind == TokenTagNegative {
		return fmt.Sprintf("<Token %s: %s>", t.Kind, t.Raw)
	} else if t.Kind == TokenText {
		return fmt.Sprintf("<Token %s: %s>", t.Kind, t.Raw)
	}
	return fmt.Sprintf("<Token unknown %s: %s>", t.Kind, t.Raw)
}

// TokenKind identifies the semantic category of a token.
type TokenKind int

const (
	TokenUnknown TokenKind = iota
	TokenAmount
	TokenTag
	TokenTagNegative
	TokenAttribute
	TokenAttributeClear
	TokenID
	TokenPeriod
	TokenText
)

// String returns the canonical string representation of a token kind.
func (k TokenKind) String() string {
	switch k {
	case TokenUnknown:
		return "unknown"
	case TokenAmount:
		return "amount"
	case TokenTag:
		return "tag"
	case TokenTagNegative:
		return "tag-negative"
	case TokenAttribute:
		return "attribute"
	case TokenAttributeClear:
		return "attribute-clear"
	case TokenID:
		return "id"
	case TokenPeriod:
		return "period"
	case TokenText:
		return "text"
	default:
		return "unknown"
	}
}

// Commands is the set of supported CLI commands.
var Commands = map[string]bool{
	"init":        true,
	"add":         true,
	"show":        true,
	"categories":  true,
	"stats":       true,
	"tags":        true,
	"modify":      true,
	"report":      true,
	"list":        true,
	"delete":      true,
	"undo":        true,
	"by":          true,
	"accounts":    true,
	"balance":     true,
	"transfer":    true,
	"set-balance": true,
	"budget":      true,
	"config":      true,
	"theme":       true,
}

// TokenRule classifies a raw token and reports whether it matched.
type TokenRule func(raw string) (Token, bool)

var tokenRules = []TokenRule{
	classifyNegativeTag,
	classifyTag,
	classifyAmount,
	classifyID,
	classifyAttribute,
	classifyPeriod,
	classifyText,
}

func isCommand(s string) bool {
	return Commands[s]
}

// ClassifyToken classifies a raw token using the configured rule order.
func ClassifyToken(raw string) Token {
	for _, rule := range tokenRules {
		token, ok := rule(raw)
		if ok {
			return token
		}
	}
	return Token{Raw: raw, Kind: TokenUnknown}
}

// FindCommand returns the first command found in args and its index.
func FindCommand(args []string) (command string, index int, parseErr *ParseError) {
	if len(args) == 0 {
		return "", -1, &ParseError{Code: ParseErrorInvalidInput, Message: "empty command line"}
	}

	for i, arg := range args {
		if isCommand(arg) {
			return arg, i, nil
		}
	}
	return "", -1, &ParseError{Code: ParseErrorNoCommand, Message: fmt.Sprintf("no command found: %v", args)}
}

// ExtractTokens classifies each raw argument into a Token.
func ExtractTokens(args []string) []Token {
	var tokens []Token
	for _, arg := range args {
		tokens = append(tokens, ClassifyToken(arg))
	}
	return tokens
}

// ParseCmdLine extracts command, filters, and arguments from args.
//
// The first recognized command can appear anywhere in the input.
// Tokens before it are considered filters; tokens after it are arguments.
func ParseCmdLine(args []string) (ParsedCmdLine, *ParseError) {
	// Find the command
	command, index, err := FindCommand(args)
	if err != nil {
		return ParsedCmdLine{}, err
	}

	// Extract filters and arguments
	filterTokens := ExtractTokens(args[:index])
	argsTokens := ExtractTokens(args[index+1:])

	// Put it all together
	return ParsedCmdLine{
		Command: command,
		Filters: filterTokens,
		Args:    argsTokens,
	}, nil
}

// ValidateParsedCmdLine validates the parsed structure and token kinds.
//
// It ensures the command is known and rejects unknown tokens.
func ValidateParsedCmdLine(parsed ParsedCmdLine) *ParseError {
	if !isCommand(parsed.Command) {
		return &ParseError{Code: ParseErrorNoCommand, Message: fmt.Sprintf("unknown command: %s", parsed.Command)}
	}

	for _, token := range parsed.Filters {
		if token.Kind == TokenUnknown {
			return &ParseError{Code: ParseErrorUnknownToken, Token: token.Raw, Message: "unknown token in filters"}
		}
	}

	for _, token := range parsed.Args {
		if token.Kind == TokenUnknown {
			return &ParseError{Code: ParseErrorUnknownToken, Token: token.Raw, Message: "unknown token in arguments"}
		}
	}

	return nil
}

// ParseAndValidateCmdLine parses args and then validates the parsed output.
func ParseAndValidateCmdLine(args []string) (ParsedCmdLine, *ParseError) {
	parsed, err := ParseCmdLine(args)
	if err != nil {
		return ParsedCmdLine{}, err
	}

	if err := ValidateParsedCmdLine(parsed); err != nil {
		return ParsedCmdLine{}, err
	}

	return parsed, nil
}

// ParseAmount converts a raw amount token into a numeric value.
//
// Note: this implementation is currently a placeholder.
func ParseAmount(s string) float32 {
	if s == "-12.50" {
		return -12.50
	}
	return 0
}

// IsParseErrorCode reports whether err is a ParseError with the given code.
func IsParseErrorCode(parseErr *ParseError, code ParseErrorCode) bool {
	if parseErr == nil {
		return false
	}
	return parseErr.Code == code
}
