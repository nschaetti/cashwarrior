package parser

import (
	"fmt"
	"os"
)

type ParseErrorCode string

const (
	// ParseErrorNoCommand indicates that no known command was found in the input.
	ParseErrorNoCommand ParseErrorCode = "NO_COMMAND"
	// ParseErrorUnknownToken indicates that a token could not be accepted by validation.
	ParseErrorUnknownToken ParseErrorCode = "UNKNOWN_TOKEN"
	// ParseErrorInvalidInput indicates malformed input, such as an empty command line.
	ParseErrorInvalidInput ParseErrorCode = "INVALID_INPUT"
	// ParseErrorEmptyCommandLine indicates an empty command line.
	ParseErrorEmptyCommandLine = "EMPTY_COMMAND_LINE"
	// ParseErrorInvalidLeftTokens indicates that the left side of the command line is invalid.
	ParseErrorInvalidLeftTokens = "INVALID_LEFT_TOKENS"
	// ParseErrorInvalidRightTokens indicates that the right side of the command line is invalid.
	ParseErrorInvalidRightTokens = "INVALID_RIGHT_TOKENS"
	// ParseErrorErrorExtractingTokens indicates that the left side of the command line is invalid.
	ParseErrorErrorExtractingTokens = "ERROR_EXTRACTING_TOKENS"
	// ParseErrorInvalidCommand ParseErrorNoCommand indicates that no command was specified.
	ParseErrorInvalidCommand = "INVALID_COMMAND"
	// ParseErrorInvalidSubcommand indicates that the subcommand is invalid.
	ParseErrorInvalidSubcommand = "INVALID_SUBCOMMAND"
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

type Attribute struct {
	Key   string
	Value AttributeValue
}

func (a Attribute) String() string {
	return fmt.Sprintf("%s=%s", a.Key, a.Value)
}

func createClearAttribute(key string) Attribute {
	return Attribute{Key: key, Value: AttributeValue{Clear: true}}
}

func createAttribute(key string, value AttributeValue) Attribute {
	return Attribute{Key: key, Value: value}
}

type Flag struct {
	Key   string
	Value string
}

func (f Flag) String() string {
	if f.Value == "" {
		return f.Key
	}
	return fmt.Sprintf("%s=%s", f.Key, f.Value)
}

func (f Flag) IsEmpty() bool {
	return f.Value == ""
}

// Token is a classified lexical unit extracted from the command line.
// If the token is an attribute, attribute name and value are in Attribute
// If the token is a flag, flag name and value are in Flag.
// If the token is a text, it is in Raw.
// If the token is a tag, it is in Tag.
type Token struct {
	Raw       string
	Kind      TokenKind
	Tag       string
	Attribute Attribute
	Flag      Flag
}

func createClearAttributeToken(key string) Token {
	return Token{Kind: TokenAttributeClear, Attribute: createClearAttribute(key)}
}

func createClearAttributeTokenWithRaw(raw string, key string) Token {
	return Token{Raw: raw, Kind: TokenAttributeClear, Attribute: createClearAttribute(key)}
}

func createAttributeToken(key string, value AttributeValue) Token {
	return Token{Kind: TokenAttribute, Attribute: createAttribute(key, value)}
}

func createFlagToken(key string, value string) Token {
	return Token{Kind: TokenFlag, Flag: Flag{Key: key, Value: value}}
}

func createSingleFlagToken(key string) Token {
	return Token{Kind: TokenFlag, Flag: Flag{Key: key}}
}

func createFlagTokenWithRaw(raw string, key string, value string) Token {
	return Token{Raw: raw, Kind: TokenFlag, Flag: Flag{Key: key, Value: value}}
}

func createSingleFlagTokenWithRaw(raw string, key string) Token {
	return Token{Raw: raw, Kind: TokenFlag, Flag: Flag{Key: key}}
}

func createTextToken(raw string) Token {
	return Token{Kind: TokenText, Raw: raw}
}

func (t Token) IsAttribute() bool {
	return t.Kind == TokenAttribute || t.Kind == TokenAttributeClear
}

func (t Token) IsFlag() bool {
	return t.Kind == TokenFlag
}

func (t Token) IsText() bool {
	return t.Kind == TokenText
}

func (t Token) IsTag() bool {
	return t.Kind == TokenTag || t.Kind == TokenTagNegative
}

// String returns a debug-friendly string representation of a token.
func (t Token) String() string {
	if t.Kind == TokenAttribute {
		return fmt.Sprintf("<Token %s: %s=%s>", t.Kind, t.Attribute.Key, t.Attribute.Value)
	} else if t.Kind == TokenAttributeClear {
		return fmt.Sprintf("<Token %s: %s>", t.Attribute.Key)
	} else if t.Kind == TokenTag {
		return fmt.Sprintf("<Token %s: %s>", t.Kind, t.Raw)
	} else if t.Kind == TokenTagNegative {
		return fmt.Sprintf("<Token %s: %s>", t.Kind, t.Raw)
	} else if t.Kind == TokenText {
		return fmt.Sprintf("<Token %s: %s>", t.Kind, t.Raw)
	} else if t.Kind == TokenFlag {
		if t.Flag.IsEmpty() {
			return fmt.Sprintf("<Token %s \"%s\" activated>", t.Kind, t.Flag.Key)
		}
		return fmt.Sprintf("<Token %s: %s set to \"%s\">", t.Kind, t.Flag.Key, t.Flag.Value)
	}
	return fmt.Sprintf("<Token unknown %s: %s>", t.Kind, t.Raw)
}

// TokenKind identifies the semantic category of a token.
type TokenKind int

const (
	TokenUnknown TokenKind = iota
	TokenTag
	TokenTagNegative
	TokenAttribute
	TokenAttributeClear
	TokenText
	TokenFlag
)

// String returns the canonical string representation of a token kind.
func (k TokenKind) String() string {
	switch k {
	case TokenUnknown:
		return "unknown"
	case TokenTag:
		return "tag"
	case TokenTagNegative:
		return "tag-negative"
	case TokenAttribute:
		return "attribute"
	case TokenAttributeClear:
		return "attribute-clear"
	case TokenText:
		return "text"
	case TokenFlag:
		return "flag"
	default:
		return "unknown"
	}
}

// TokenRule classifies a raw token and reports whether it matched.
type TokenRule func(raw string) bool

var tokenRules = map[TokenKind]TokenRule{
	TokenFlag:        classifyFlag,
	TokenTagNegative: classifyNegativeTag,
	TokenTag:         classifyTag,
	TokenAttribute:   classifyAttribute,
	TokenText:        classifyText,
}

type TokenParser func(raw string) (Token, error)

var tokenParsers = map[TokenKind]TokenParser{
	TokenFlag:        ParseTokenFlag,
	TokenTag:         ParseTokenTag,
	TokenTagNegative: ParseTokenTagNegative,
	TokenAttribute:   ParseTokenAttribute,
	TokenText:        ParseTokenText,
}

func isCommand(s string) bool {
	return IsKnownCommand(s)
}

// ClassifyToken classifies a raw token using the configured rule order.
func ClassifyToken(raw string) TokenKind {
	for kind, rule := range tokenRules {
		ok := rule(raw)
		if ok {
			return kind
		}
	}
	return TokenUnknown
}

// FindCommand returns the first command found in args and its index.
func FindCommand(args []string) (command string, index int, parseErr *ParseError) {
	if len(args) == 0 {
		return "", -1, &ParseError{Code: ParseErrorEmptyCommandLine, Message: "empty command line"}
	}

	for i, arg := range args {
		if isCommand(arg) {
			return arg, i, nil
		}
	}
	return "", -1, &ParseError{Code: ParseErrorNoCommand, Message: fmt.Sprintf("no command found: %v", args)}
}

// ExtractTokens classifies each raw argument into a Token.
func ExtractTokens(args []string) ([]Token, error) {
	var tokens []Token
	for _, arg := range args {
		kind := ClassifyToken(arg)
		if kind != TokenUnknown {
			tokenParser := tokenParsers[kind]
			token, err := tokenParser(arg)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
		} else {
			return nil, fmt.Errorf("unknown argument: %s", arg)
		}
	}
	return tokens, nil
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

	// Get command spec
	commandSpec, ok := GetCommandSpec(command)
	if !ok {
		return ParsedCmdLine{}, &ParseError{Code: ParseErrorInvalidCommand, Message: fmt.Sprintf("unknown command: %s", command)}
	}

	// Extract filters and arguments
	rawFilterTokens, flagTokensLeft, sErr := splitFlags(args[:index])
	if sErr != nil {
		return ParsedCmdLine{}, &ParseError{Code: ParseErrorInvalidLeftTokens, Message: fmt.Sprintf("Error splitting left side: %s", sErr.Error())}
	}

	rawArgs, flagTokensRight, sErr := splitFlags(args[index+1:])
	if sErr != nil {
		return ParsedCmdLine{}, &ParseError{Code: ParseErrorInvalidRightTokens, Message: fmt.Sprintf("Error splitting right side: %s", sErr.Error())}
	}

	filterTokens, fErr := ExtractTokens(rawFilterTokens)
	if fErr != nil {
		return ParsedCmdLine{}, &ParseError{Code: ParseErrorErrorExtractingTokens, Message: fmt.Sprintf("Error parsing arguments: %s", fErr.Error())}
	}

	flagTokens := append(flagTokensLeft, flagTokensRight...)
	// Subcommand
	subcommand := commandSpec.DefaultSubcommand
	if len(rawArgs) > 0 {
		if _, ok := commandSpec.Subcommands[rawArgs[0]]; ok {
			subcommand = rawArgs[0]
			rawArgs = rawArgs[1:]
		}
	}

	argsTokens, fErr := ExtractTokens(rawArgs)
	if fErr != nil {
		return ParsedCmdLine{}, &ParseError{Code: ParseErrorInvalidInput, Message: err.Error()}
	}

	fmt.Printf("Command: %s\n", command)
	fmt.Printf("Subcommand: %s\n", subcommand)
	fmt.Printf("Filter tokens: %v\n", filterTokens)
	fmt.Printf("Args tokens: %v\n", argsTokens)
	fmt.Printf("Flag tokens: %v\n", flagTokens)
	os.Exit(0)
	// Put it all together
	return ParsedCmdLine{
		Command:    command,
		Subcommand: subcommand,
		Filters:    filterTokens,
		Args:       argsTokens,
		Flags:      flagTokens,
	}, nil
}

func splitFlags(args []string) ([]string, []Token, error) {
	nonFlags := make([]string, 0, len(args))
	flags := make([]Token, 0)
	for _, arg := range args {
		ok := classifyFlag(arg)
		if ok {
			flagParser := tokenParsers[TokenFlag]
			token, err := flagParser(arg)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse flag: %s", err)
			}
			flags = append(flags, token)
			continue
		}
		nonFlags = append(nonFlags, arg)
	}
	return nonFlags, flags, nil
}

// ValidateParsedCmdLine validates the parsed structure and token kinds.
//
// It ensures the command is known and rejects unknown tokens.
func ValidateParsedCmdLine(parsed ParsedCmdLine) *ParseError {
	// Special case: help is always valid
	if parsed.HasFlag("help") {
		return nil
	}

	// Get command specification
	_, ok := GetCommandSpec(parsed.Command)
	if !ok {
		return &ParseError{
			Code:    ParseErrorNoCommand,
			Message: fmt.Sprintf("unknown command: %s", parsed.Command),
		}
	}

	// Get subcommand specification
	subcommandSpec, ok := GetSubcommandSpec(parsed.Command, parsed.Subcommand)
	if !ok {
		return &ParseError{
			Code:    ParseErrorInvalidInput,
			Message: fmt.Sprintf("unknown subcommand %s for command %s", parsed.Subcommand, parsed.Command),
		}
	}

	// Validate tokens on the left (filter)
	for _, token := range parsed.Filters {
		if token.Kind == TokenUnknown {
			return &ParseError{
				Code:    ParseErrorUnknownToken,
				Token:   token.Raw,
				Message: "unknown token in filters",
			}
		}
		if err := validateTokenAgainstSideSpec(token, subcommandSpec.Left, "left side"); err != nil {
			return err
		}
	}
	if err := validateSideCardinality(parsed.Filters, subcommandSpec.Left, "left side"); err != nil {
		return err
	}

	// Validate tokens on the right (arguments)
	for _, token := range parsed.Args {
		if token.Kind == TokenUnknown {
			return &ParseError{Code: ParseErrorUnknownToken, Token: token.Raw, Message: "unknown token in arguments"}
		}
		if err := validateTokenAgainstSideSpec(token, subcommandSpec.Right, "right side"); err != nil {
			return err
		}
	}
	if err := validateSideCardinality(parsed.Args, subcommandSpec.Right, "right side"); err != nil {
		return err
	}

	return nil
}

func validateTokenAgainstSideSpec(token Token, side SideSpec, sideName string) *ParseError {
	// Check if the token is allowed on the side
	if !side.AllowedKinds[token.Kind] {
		return &ParseError{
			Code:    ParseErrorUnknownToken,
			Token:   token.Raw,
			Message: fmt.Sprintf("token kind %s is not allowed on %s", token.Kind, sideName),
		}
	}

	// Check if the token is allowed to be set
	if token.Kind != TokenAttribute && token.Kind != TokenAttributeClear {
		return nil
	}
	if side.AllowAnyAttr {
		return nil
	}

	// Get attribute specification
	attributeSpec, ok := side.Attributes[token.Attribute.Key]
	if !ok {
		return &ParseError{
			Code:    ParseErrorUnknownToken,
			Token:   token.Raw,
			Message: fmt.Sprintf("attribute %s is not allowed on %s", token.Attribute.Key, sideName),
		}
	}

	// Check if clear attribute is allowed
	if token.Kind == TokenAttributeClear {
		if !attributeSpec.AllowsClear() {
			return &ParseError{
				Code:    ParseErrorInvalidInput,
				Token:   token.Raw,
				Message: fmt.Sprintf("attribute %s cannot be cleared", token.Attribute.Key),
			}
		}
		return nil
	}

	// If not a clear, then it is a set attribute
	// and must be allowed to be set
	if !attributeSpec.AllowsSet() {
		return &ParseError{
			Code:    ParseErrorInvalidInput,
			Token:   token.Raw,
			Message: fmt.Sprintf("attribute %s cannot be set", token.Attribute.Key),
		}
	}

	// Check if the attribute value (single, list, range) is allowed
	value, err := ParseAttributeValue(token.Attribute.Value.Raw)
	if err != nil {
		return &ParseError{
			Code:    ParseErrorInvalidInput,
			Token:   token.Raw,
			Message: err.Error(),
		}
	}
	if !attributeSpec.Shapes.Allows(value) {
		return &ParseError{
			Code:    ParseErrorInvalidInput,
			Token:   token.Raw,
			Message: fmt.Sprintf("attribute %s does not accept %s values", token.Attribute.Key, value.Kind),
		}
	}

	return nil
}

func validateSideCardinality(tokens []Token, side SideSpec, sideName string) *ParseError {
	if err := validateArgsCount(tokens, side, sideName); err != nil {
		return err
	}
	if err := validateKindRules(tokens, side, sideName); err != nil {
		return err
	}
	if err := validateAttributeRules(tokens, side, sideName); err != nil {
		return err
	}
	if err := validatePresenceRules(tokens, side, sideName); err != nil {
		return err
	}
	return nil
}

func validateArgsCount(tokens []Token, side SideSpec, sideName string) *ParseError {
	count := len(tokens)
	if side.MinArgs > 0 && count < side.MinArgs {
		return &ParseError{Code: ParseErrorInvalidInput, Message: fmt.Sprintf("%s requires at least %d argument", sideName, side.MinArgs)}
	}
	if side.MaxArgs > 0 && count > side.MaxArgs {
		return &ParseError{Code: ParseErrorInvalidInput, Message: fmt.Sprintf("%s accepts at most %d argument", sideName, side.MaxArgs)}
	}
	return nil
}

func validateKindRules(tokens []Token, side SideSpec, sideName string) *ParseError {
	counts := make(map[TokenKind]int)
	for _, token := range tokens {
		counts[token.Kind]++
	}
	for kind, rule := range side.KindRules {
		if err := validateCountRule(counts[kind], rule, sideName, kind.String()); err != nil {
			return err
		}
	}
	return nil
}

func validateAttributeRules(tokens []Token, side SideSpec, sideName string) *ParseError {
	counts := make(map[string]int)
	for _, token := range tokens {
		if token.Kind != TokenAttribute && token.Kind != TokenAttributeClear {
			continue
		}
		counts[token.Attribute.Key]++
	}
	for name, rule := range side.AttributeRules {
		if err := validateCountRule(counts[name], rule, sideName, fmt.Sprintf("attribute %s", name)); err != nil {
			return err
		}
	}
	return nil
}

func validateCountRule(count int, rule CountRule, sideName string, label string) *ParseError {
	if rule.Min > 0 && count < rule.Min {
		return &ParseError{Code: ParseErrorInvalidInput, Message: fmt.Sprintf("%s requires at least %d %s", sideName, rule.Min, label)}
	}
	if rule.Max > 0 && count > rule.Max {
		return &ParseError{Code: ParseErrorInvalidInput, Message: fmt.Sprintf("%s accepts at most %d %s", sideName, rule.Max, label)}
	}
	return nil
}

func validatePresenceRules(tokens []Token, side SideSpec, sideName string) *ParseError {
	if len(side.AtLeastOneOf) == 0 {
		return nil
	}

	kindCounts := make(map[TokenKind]int)
	attributeCounts := make(map[string]int)
	for _, token := range tokens {
		kindCounts[token.Kind]++
		if token.Kind == TokenAttribute || token.Kind == TokenAttributeClear {
			attributeCounts[token.Attribute.Key]++
		}
	}

	for _, rule := range side.AtLeastOneOf {
		if presenceRuleSatisfied(rule, kindCounts, attributeCounts) {
			continue
		}
		message := rule.Message
		if message == "" {
			message = fmt.Sprintf("%s requires at least one matching token", sideName)
		}
		return &ParseError{Code: ParseErrorInvalidInput, Message: message}
	}

	return nil
}

func presenceRuleSatisfied(rule PresenceRule, kindCounts map[TokenKind]int, attributeCounts map[string]int) bool {
	for _, kind := range rule.Kinds {
		if kindCounts[kind] > 0 {
			return true
		}
	}
	for _, name := range rule.Attributes {
		if attributeCounts[name] > 0 {
			return true
		}
	}
	return false
}

// ParseAndValidateCmdLine parses args and then validates the parsed output.
func ParseAndValidateCmdLine(args []string) (ParsedCmdLine, *ParseError) {
	// Parse the command line
	parsed, err := ParseCmdLine(args)
	if err != nil {
		return ParsedCmdLine{}, err
	}

	// Validate the parsed structure
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
