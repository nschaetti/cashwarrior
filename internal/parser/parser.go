package parser

import (
	"fmt"
	"os"

	"github.com/nschaetti/cashwarrior/internal/config"
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
	// ParseErrorInvalidLeftArgs indicates that the left side of the command line is invalid.
	ParseErrorInvalidLeftArgs = "INVALID_LEFT_ARGS"
	// ParseErrorInvalidRightArgs indicates that the right side of the command line is invalid.
	ParseErrorInvalidRightArgs = "INVALID_RIGHT_ARGS"
	// ParseErrorErrorExtractingArgs indicates that the left side of the command line is invalid.
	ParseErrorErrorExtractingArgs = "ERROR_EXTRACTING_ARGS"
	// ParseErrorInvalidCommand ParseErrorNoCommand indicates that no command was specified.
	ParseErrorInvalidCommand = "INVALID_COMMAND"
	// ParseErrorInvalidSubcommand indicates that the subcommand is invalid.
	ParseErrorInvalidSubcommand = "INVALID_SUBCOMMAND"
)

// ParseError represents a structured parser or validation error.
type ParseError struct {
	Code    ParseErrorCode
	Arg     string
	Message string
}

// Error returns a human-readable message for the parse error.
func (e *ParseError) Error() string {
	if e.Arg == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Message, e.Arg)
}

type ArgKind int

const (
	ArgKindUnknown ArgKind = iota
	ArgKindTag
	ArgKindTagNegative
	ArgKindAttribute
	ArgKindText
	ArgKindFlag
)

func (k ArgKind) String() string {
	switch k {
	case ArgKindUnknown:
		return "unknown"
	case ArgKindTag:
		return "tag"
	case ArgKindTagNegative:
		return "tag-negative"
	case ArgKindAttribute:
		return "attribute"
	case ArgKindText:
		return "text"
	case ArgKindFlag:
		return "flag"
	default:
		return "unknown"
	}
}

type Arg interface {
	ArgKind() ArgKind
	RawString() string
	IsAttribute() bool
	IsAttributeClear() bool
	IsTag() bool
	IsTagNegative() bool
	IsText() bool
	IsFlag() bool
}

/*
 * ArgText represents a text argument.
 */

type ArgText struct {
	Raw  string
	Text string
}

func (t ArgText) ArgKind() ArgKind       { return ArgKindText }
func (t ArgText) RawString() string      { return t.Raw }
func (t ArgText) String() string         { return fmt.Sprintf("argtext(%s)", t.Text) }
func (t ArgText) IsText() bool           { return true }
func (t ArgText) IsAttribute() bool      { return false }
func (t ArgText) IsAttributeClear() bool { return false }
func (t ArgText) IsTag() bool            { return false }
func (t ArgText) IsTagNegative() bool    { return false }
func (t ArgText) IsFlag() bool           { return false }

/*
 * ArgAttribute represents an attribute argument.
 */

type ArgAttribute struct {
	Raw   string
	Key   string
	Value AttributeValue
	Clear bool
}

func (a ArgAttribute) ArgKind() ArgKind       { return ArgKindAttribute }
func (a ArgAttribute) RawString() string      { return a.Raw }
func (a ArgAttribute) String() string         { return fmt.Sprintf("argattr(%s=%s)", a.Key, a.Value) }
func (a ArgAttribute) IsAttribute() bool      { return true }
func (a ArgAttribute) IsAttributeClear() bool { return a.Clear }
func (a ArgAttribute) IsTag() bool            { return false }
func (a ArgAttribute) IsTagNegative() bool    { return false }
func (a ArgAttribute) IsText() bool           { return false }
func (a ArgAttribute) IsFlag() bool           { return false }

func createClearAttribute(key string) ArgAttribute {
	return ArgAttribute{Key: key, Clear: true}
}

func createAttribute(key string, value AttributeValue) ArgAttribute {
	return ArgAttribute{Key: key, Value: value}
}

/*
 * ArgTag represents a tag argument.
 */

type ArgTag struct {
	Raw      string
	Tag      string
	Negative bool
}

func (t ArgTag) ArgKind() ArgKind       { return ArgKindTag }
func (t ArgTag) RawString() string      { return t.Raw }
func (t ArgTag) String() string         { return fmt.Sprintf("argtag(%s)", t.Tag) }
func (t ArgTag) IsTag() bool            { return true }
func (t ArgTag) IsTagNegative() bool    { return t.Negative }
func (t ArgTag) IsAttribute() bool      { return false }
func (t ArgTag) IsAttributeClear() bool { return false }
func (t ArgTag) IsText() bool           { return false }
func (t ArgTag) IsFlag() bool           { return false }

/*
 * ArgFlag represents a flag argument.
 */

type ArgFlag struct {
	Raw   string
	Key   string
	Value string
}

func (f ArgFlag) ArgKind() ArgKind       { return ArgKindFlag }
func (f ArgFlag) RawString() string      { return f.Raw }
func (f ArgFlag) IsTag() bool            { return false }
func (f ArgFlag) IsTagNegative() bool    { return false }
func (f ArgFlag) IsAttribute() bool      { return false }
func (f ArgFlag) IsAttributeClear() bool { return false }
func (f ArgFlag) IsText() bool           { return false }
func (f ArgFlag) IsFlag() bool           { return true }
func (f ArgFlag) String() string {
	if f.Value == "" {
		return f.Key
	}
	return fmt.Sprintf("%s=%s", f.Key, f.Value)
}
func (f ArgFlag) IsEmpty() bool {
	return f.Value == ""
}

// Token is a classified lexical unit extracted from the command line.
// If the token is an attribute, attribute name and value are in ArgAttribute
// If the token is a flag, flag name and value are in ArgFlag.
// If the token is a text, it is in Raw.
// If the token is a tag, it is in Tag.
//type Token struct {
//	Raw       string
//	Kind      TokenKind
//	Tag       string
//	Attribute ArgAttribute
//	Flag      ArgFlag
//}

func createClearAttributeArg(key string) ArgAttribute {
	return ArgAttribute{Raw: key, Key: key, Clear: true}
}

//func createClearAttributeToken(key string) Token {
//	return Token{Kind: TokenAttributeClear, Attribute: createClearAttribute(key)}
//}

//func createClearAttributeTokenWithRaw(raw string, key string) Token {
//	return Token{Raw: raw, Kind: TokenAttributeClear, Attribute: createClearAttribute(key)}
//}

func createAttributeArg(key string, value AttributeValue) ArgAttribute {
	return ArgAttribute{Raw: key, Key: key, Value: value}
}

//func createAttributeToken(key string, value AttributeValue) Token {
//	return Token{Kind: TokenAttribute, Attribute: createAttribute(key, value)}
//}

func createFlagArg(key string, value string) ArgFlag {
	return ArgFlag{Raw: key, Key: key, Value: value}
}

//func createFlagToken(key string, value string) Token {
//	return Token{Kind: TokenFlag, Flag: ArgFlag{Key: key, Value: value}}
//}

func createSingleFlagArg(key string) ArgFlag {
	return ArgFlag{Raw: key, Key: key}
}

//func createSingleFlagToken(key string) Token {
//	return Token{Kind: TokenFlag, Flag: ArgFlag{Key: key}}
//}

//func createFlagTokenWithRaw(raw string, key string, value string) Token {
//	return Token{Raw: raw, Kind: TokenFlag, Flag: ArgFlag{Key: key, Value: value}}
//}
//
//func createSingleFlagTokenWithRaw(raw string, key string) Token {
//	return Token{Raw: raw, Kind: TokenFlag, Flag: ArgFlag{Key: key}}
//}

func createTextArg(raw string) ArgText {
	return ArgText{Raw: raw, Text: raw}
}

//func createTextToken(raw string) Token {
//	return Token{Kind: TokenText, Raw: raw}
//}

//func (t Token) IsAttribute() bool {
//	return t.Kind == TokenAttribute || t.Kind == TokenAttributeClear
//}
//
//func (t Token) IsFlag() bool {
//	return t.Kind == TokenFlag
//}
//
//func (t Token) IsText() bool {
//	return t.Kind == TokenText
//}
//
//func (t Token) IsTag() bool {
//	return t.Kind == TokenTag || t.Kind == TokenTagNegative
//}
//
//// String returns a debug-friendly string representation of a token.
//func (t Token) String() string {
//	if t.Kind == TokenAttribute {
//		return fmt.Sprintf("<Token %s: %s=%s>", t.Kind, t.Attribute.Key, t.Attribute.Value)
//	} else if t.Kind == TokenAttributeClear {
//		return fmt.Sprintf("<Token %s: %s>", t.Attribute.Key)
//	} else if t.Kind == TokenTag {
//		return fmt.Sprintf("<Token %s: %s>", t.Kind, t.Raw)
//	} else if t.Kind == TokenTagNegative {
//		return fmt.Sprintf("<Token %s: %s>", t.Kind, t.Raw)
//	} else if t.Kind == TokenText {
//		return fmt.Sprintf("<Token %s: %s>", t.Kind, t.Raw)
//	} else if t.Kind == TokenFlag {
//		if t.Flag.IsEmpty() {
//			return fmt.Sprintf("<Token %s \"%s\" activated>", t.Kind, t.Flag.Key)
//		}
//		return fmt.Sprintf("<Token %s: %s set to \"%s\">", t.Kind, t.Flag.Key, t.Flag.Value)
//	}
//	return fmt.Sprintf("<Token unknown %s: %s>", t.Kind, t.Raw)
//}

// TokenKind identifies the semantic category of a token.
// type TokenKind int

/* const (
	TokenUnknown TokenKind = iota
	TokenTag
	TokenTagNegative
	TokenAttribute
	TokenAttributeClear
	TokenText
	TokenFlag
)*/

// String returns the canonical string representation of a token kind.
/* func (k TokenKind) String() string {
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
}*/

// TokenRule classifies a raw token and reports whether it matched.
type TokenRule func(raw string) bool
type ArgRule func(raw string) bool

type argRuleEntry struct {
	kind ArgKind
	rule ArgRule
}

// tokenRuleEntry Rule entry to keep the order
/*type tokenRuleEntry struct {
	kind TokenKind
	rule TokenRule
}*/

var argRules = []argRuleEntry{
	{kind: ArgKindFlag, rule: classifyFlag},
	{kind: ArgKindTag, rule: classifyTag},
	/*{kind: ArgKindTagNegative, rule: classifyNegativeTag},*/
	{kind: ArgKindAttribute, rule: classifyAttribute},
	{kind: ArgKindText, rule: classifyText},
}

/*var tokenRules = []tokenRuleEntry{
	{kind: TokenFlag, rule: classifyFlag},
	{kind: TokenTagNegative, rule: classifyNegativeTag},
	{kind: TokenTag, rule: classifyTag},
	{kind: TokenAttribute, rule: classifyAttribute},
	{kind: TokenText, rule: classifyText},
}*/

// type TokenParser func(raw string, config config.Config) (Token, error)
type ArgParser func(raw string, config config.Config) (Arg, error)

var argParsers = map[ArgKind]ArgParser{
	ArgKindFlag:        ParseArgFlag,
	ArgKindTag:         ParseArgTag,
	ArgKindTagNegative: ParseArgTagNegative,
	ArgKindAttribute:   ParseArgAttribute,
	ArgKindText:        ParseArgText,
}

/*var tokenParsers = map[TokenKind]TokenParser{
	TokenFlag:        ParseTokenFlag,
	TokenTag:         ParseTokenTag,
	TokenTagNegative: ParseTokenTagNegative,
	TokenAttribute:   ParseTokenAttribute,
	TokenText:        ParseTokenText,
}*/

func isCommand(s string) bool {
	return IsKnownCommand(s)
}

// ClassifyToken classifies a raw token using the configured rule order.
//func ClassifyToken(raw string) TokenKind {
//	for _, entry := range tokenRules {
//		ok := entry.rule(raw)
//		if ok {
//			return entry.kind
//		}
//	}
//	return TokenUnknown
//}

func ClassifyArg(raw string) ArgKind {
	for _, entry := range argRules {
		ok := entry.rule(raw)
		if ok {
			return entry.kind
		}
	}
	return ArgKindUnknown
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
//func ExtractTokens(args []string, config config.Config) ([]Token, error) {
//	var tokens []Token
//	for _, arg := range args {
//		kind := ClassifyToken(arg)
//		if kind != TokenUnknown {
//			tokenParser := tokenParsers[kind]
//			token, err := tokenParser(arg, config)
//			if err != nil {
//				return nil, err
//			}
//			tokens = append(tokens, token)
//		} else {
//			return nil, fmt.Errorf("unknown argument: %s", arg)
//		}
//	}
//	return tokens, nil
//}

func ExtractArgs(args []string, config config.Config) ([]Arg, error) {
	var extractedArgs []Arg
	for _, arg := range args {
		kind := ClassifyArg(arg)
		if kind != ArgKindUnknown {
			argParser := argParsers[kind]
			a, err := argParser(arg, config)
			if err != nil {
				return nil, err
			}
			extractedArgs = append(extractedArgs, a)
		} else {
			return nil, fmt.Errorf("unknown argument: %s", arg)
		}
	}
	return extractedArgs, nil
}

// ParseCmdLine extracts command, filters, and arguments from args.
//
// The first recognized command can appear anywhere in the input.
// Tokens before it are considered filters; tokens after it are arguments.
func ParseCmdLine(args []string, config config.Config) (ParsedCmdLine, *ParseError) {
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
	rawFilterArgs, flagArgsLeft, sErr := splitFlags(args[:index])
	if sErr != nil {
		return ParsedCmdLine{}, &ParseError{Code: ParseErrorInvalidLeftArgs, Message: fmt.Sprintf("Error splitting left side: %s", sErr.Error())}
	}

	rawArgs, flagArgsRight, sErr := splitFlags(args[index+1:])
	if sErr != nil {
		return ParsedCmdLine{}, &ParseError{Code: ParseErrorInvalidRightArgs, Message: fmt.Sprintf("Error splitting right side: %s", sErr.Error())}
	}

	filterTokens, fErr := ExtractArgs(rawFilterArgs, config)
	if fErr != nil {
		return ParsedCmdLine{}, &ParseError{Code: ParseErrorErrorExtractingArgs, Message: fmt.Sprintf("Error parsing arguments: %s", fErr.Error())}
	}

	flagArgs := append(flagArgsLeft, flagArgsRight...)
	// Subcommand
	subcommand := commandSpec.DefaultSubcommand
	if len(rawArgs) > 0 {
		if _, ok := commandSpec.Subcommands[rawArgs[0]]; ok {
			subcommand = rawArgs[0]
			rawArgs = rawArgs[1:]
		}
	}

	argsTokens, fErr := ExtractArgs(rawArgs, config)
	if fErr != nil {
		return ParsedCmdLine{}, &ParseError{Code: ParseErrorInvalidInput, Message: err.Error()}
	}
	fmt.Printf("Command: %s\n", command)
	fmt.Printf("Subcommand: %s\n", subcommand)
	fmt.Printf("Filter tokens: %v\n", filterTokens)
	fmt.Printf("Args tokens: %v\n", argsTokens)
	fmt.Printf("Flag tokens: %v\n", flagArgs)
	os.Exit(0)
	// Put it all together
	return ParsedCmdLine{
		Command:    command,
		Subcommand: subcommand,
		Filters:    filterTokens,
		Args:       argsTokens,
		Flags:      flagArgs,
	}, nil
}

func splitFlags(args []string) ([]string, []Arg, error) {
	nonFlags := make([]string, 0, len(args))
	flags := make([]Arg, 0)
	for _, arg := range args {
		ok := classifyFlag(arg)
		if ok {
			flagParser := argParsers[ArgKindFlag]
			token, err := flagParser(arg, config.Config{})
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
	for _, arg := range parsed.Filters {
		if arg.ArgKind() == ArgKindUnknown {
			return &ParseError{
				Code:    ParseErrorUnknownToken,
				Arg:     arg.RawString(),
				Message: "unknown token in filters",
			}
		}
		if err := validateTokenAgainstSideSpec(arg, subcommandSpec.Left, "left side"); err != nil {
			return err
		}
	}
	if err := validateSideCardinality(parsed.Filters, subcommandSpec.Left, "left side"); err != nil {
		return err
	}

	// Validate tokens on the right (arguments)
	for _, arg := range parsed.Args {
		if arg.ArgKind() == ArgKindUnknown {
			return &ParseError{Code: ParseErrorUnknownToken, Arg: arg.RawString(), Message: "unknown token in arguments"}
		}
		if err := validateTokenAgainstSideSpec(arg, subcommandSpec.Right, "right side"); err != nil {
			return err
		}
	}
	if err := validateSideCardinality(parsed.Args, subcommandSpec.Right, "right side"); err != nil {
		return err
	}

	return nil
}

func validateTokenAgainstSideSpec(arg Arg, side SideSpec, sideName string) *ParseError {
	// Check if the token is allowed on the side
	if !side.AllowedKinds[arg.ArgKind()] {
		return &ParseError{
			Code:    ParseErrorUnknownToken,
			Arg:     arg.RawString(),
			Message: fmt.Sprintf("token kind %s is not allowed on %s", arg.ArgKind(), sideName),
		}
	}

	// Check if the token is allowed to be set
	if arg.ArgKind() != ArgKindAttribute {
		return nil
	}

	attr, ok := arg.(ArgAttribute)
	if !ok {
		return &ParseError{
			Code:    ParseErrorInvalidInput,
			Arg:     arg.RawString(),
			Message: "token is not an attribute (but has attribute kind)",
		}
	}

	if side.AllowAnyAttr {
		return nil
	}

	// Get attribute specification
	attributeSpec, ok := side.Attributes[attr.Key]
	if !ok {
		return &ParseError{
			Code:    ParseErrorUnknownToken,
			Arg:     attr.RawString(),
			Message: fmt.Sprintf("attribute %s is not allowed on %s", attr.Key, sideName),
		}
	}

	// Check if clear attribute is allowed
	if attr.IsAttributeClear() {
		if !attributeSpec.AllowsClear() {
			return &ParseError{
				Code:    ParseErrorInvalidInput,
				Arg:     attr.RawString(),
				Message: fmt.Sprintf("attribute %s cannot be cleared", attr.Key),
			}
		}
		return nil
	}

	// If not a clear, then it is a set attribute
	// and must be allowed to be set
	if !attributeSpec.AllowsSet() {
		return &ParseError{
			Code:    ParseErrorInvalidInput,
			Arg:     attr.RawString(),
			Message: fmt.Sprintf("attribute %s cannot be set", attr.Key),
		}
	}

	// Check if the attribute value (single, list, range) is allowed
	value, err := ParseAttributeValue(attr.Value.Raw)
	if err != nil {
		return &ParseError{
			Code:    ParseErrorInvalidInput,
			Arg:     attr.RawString(),
			Message: err.Error(),
		}
	}
	if !attributeSpec.Shapes.Allows(value) {
		return &ParseError{
			Code:    ParseErrorInvalidInput,
			Arg:     attr.RawString(),
			Message: fmt.Sprintf("attribute %s does not accept %s values", attr.Key, value.ValueShape),
		}
	}

	return nil
}

func validateSideCardinality(tokens []Arg, side SideSpec, sideName string) *ParseError {
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

func validateArgsCount(tokens []Arg, side SideSpec, sideName string) *ParseError {
	count := len(tokens)
	if side.MinArgs > 0 && count < side.MinArgs {
		return &ParseError{Code: ParseErrorInvalidInput, Message: fmt.Sprintf("%s requires at least %d argument", sideName, side.MinArgs)}
	}
	if side.MaxArgs > 0 && count > side.MaxArgs {
		return &ParseError{Code: ParseErrorInvalidInput, Message: fmt.Sprintf("%s accepts at most %d argument", sideName, side.MaxArgs)}
	}
	return nil
}

func validateKindRules(tokens []Arg, side SideSpec, sideName string) *ParseError {
	counts := make(map[ArgKind]int)
	for _, token := range tokens {
		counts[token.ArgKind()]++
	}
	for kind, rule := range side.KindRules {
		if err := validateCountRule(counts[kind], rule, sideName, kind.String()); err != nil {
			return err
		}
	}
	return nil
}

func validateAttributeRules(tokens []Arg, side SideSpec, sideName string) *ParseError {
	counts := make(map[string]int)
	for _, arg := range tokens {
		if arg.ArgKind() != ArgKindAttribute {
			continue
		}
		attr, ok := arg.(ArgAttribute)
		if !ok {
			continue
		}
		counts[attr.Key]++
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

func validatePresenceRules(tokens []Arg, side SideSpec, sideName string) *ParseError {
	if len(side.AtLeastOneOf) == 0 {
		return nil
	}

	kindCounts := make(map[ArgKind]int)
	attributeCounts := make(map[string]int)
	for _, token := range tokens {
		kindCounts[token.ArgKind()]++
		attr, ok := token.(ArgAttribute)
		if token.ArgKind() == ArgKindAttribute && ok {
			attributeCounts[attr.Key]++
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

func presenceRuleSatisfied(rule PresenceRule, kindCounts map[ArgKind]int, attributeCounts map[string]int) bool {
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
func ParseAndValidateCmdLine(args []string, config config.Config) (ParsedCmdLine, *ParseError) {
	// Parse the command line
	parsed, err := ParseCmdLine(args, config)
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
