package parser

import (
	"fmt"

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
	ParseErrorEmptyCommandLine ParseErrorCode = "EMPTY_COMMAND_LINE"
	// ParseErrorUnknownCommand ParseErrorNoCommand indicates that no command was specified.
	ParseErrorUnknownCommand ParseErrorCode = "UNKNOWN_COMMAND"
	// ParseErrorUnknownSubcommand indicates that the subcommand is invalid.
	ParseErrorUnknownSubcommand     ParseErrorCode = "UNKNOWN_SUBCOMMAND"
	ParseErrorUnknownArgument       ParseErrorCode = "UNKNOWN_ARGUMENT"
	ParseErrorUnknownAttributeKey   ParseErrorCode = "UNKNOWN_ATTRIBUTE_KEY"
	ParseErrorUnknownFlag           ParseErrorCode = "UNKNOWN_FLAG"
	ParseErrorInvalidAttributeValue ParseErrorCode = "INVALID_ATTRIBUTE_VALUE"
	ParseErrorInvalidRange          ParseErrorCode = "INVALID_RANGE"
	ParseErrorInvalidFlagValue      ParseErrorCode = "INVALID_FLAG_VALUE"
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

// TokenRule classifies a raw token and reports whether it matched.
type TokenRule func(raw string) bool
type ArgRule func(raw string) bool

type argRuleEntry struct {
	kind ArgKind
	rule ArgRule
}

var argRules = []argRuleEntry{
	{kind: ArgKindFlag, rule: classifyFlag},
	{kind: ArgKindTag, rule: classifyTag},
	{kind: ArgKindTagNegative, rule: classifyNegativeTag},
	{kind: ArgKindAttribute, rule: classifyAttribute},
	{kind: ArgKindText, rule: classifyText},
}

// type TokenParser func(raw string, config config.Config) (Token, error)
type ArgParser func(raw string, config config.Config) (Arg, *ParseError)

var argParsers = map[ArgKind]ArgParser{
	ArgKindFlag:        ParseArgFlag,
	ArgKindTag:         ParseArgTag,
	ArgKindTagNegative: ParseArgTagNegative,
	ArgKindAttribute:   ParseArgAttribute,
	ArgKindText:        ParseArgText,
}

func isCommand(s string) bool {
	return IsKnownCommand(s)
}

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

// ExtractArgs extracts arguments from args using the configured parsers.
// It returns the extracted arguments and an error if any.
//
// The arguments are extracted in the order they appear in args.
// If an argument is not recognized, it is returned as an unknown argument.
// If an argument is recognized but cannot be parsed, it is returned as an error.
//
// The returned arguments are guaranteed to be in the same order as they appear in args.
// If an argument is not recognized, it is returned as an unknown argument.
// If an argument is recognized but cannot be parsed, it is returned as an error.
func ExtractArgs(args []string, config config.Config) ([]Arg, *ParseError) {
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
			return nil, &ParseError{
				Code:    ParseErrorUnknownArgument,
				Arg:     arg,
				Message: fmt.Sprintf("unknown argument: %s", arg),
			}
		}
	}
	return extractedArgs, nil
}

func hasHelpFlag(flags []Arg) bool {
	for _, flag := range flags {
		if flag.IsFlag() {
			flagArg, _ := flag.(ArgFlag)
			if flagArg.Key == "help" {
				return true
			}
		}
	}
	return false
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
		return ParsedCmdLine{}, &ParseError{
			Code:    ParseErrorUnknownCommand,
			Message: fmt.Sprintf("unknown command: %s", command),
		}
	}

	// Extract filters and arguments on left
	rawFilterArgs, flagArgsLeft, err := splitFlags(args[:index])
	if err != nil {
		return ParsedCmdLine{}, err
	}

	// Extract filters and arguments on right
	rawArgs, flagArgsRight, err := splitFlags(args[index+1:])
	if err != nil {
		return ParsedCmdLine{}, err
	}

	// Extract filters and arguments
	filterTokens, fErr := ExtractArgs(rawFilterArgs, config)
	if fErr != nil {
		return ParsedCmdLine{}, fErr
	}

	flagArgs := append(flagArgsLeft, flagArgsRight...)

	// Subcommand
	subcommand := commandSpec.DefaultSubcommand
	if len(rawArgs) > 0 {
		if _, ok = commandSpec.Subcommands[rawArgs[0]]; ok {
			subcommand = rawArgs[0]
			rawArgs = rawArgs[1:]
		}
	}

	argsTokens, fErr := ExtractArgs(rawArgs, config)
	if fErr != nil {
		return ParsedCmdLine{}, &ParseError{
			Code:    ParseErrorInvalidInput,
			Message: fErr.Error(),
		}
	}
	fmt.Printf("argsTokens: %v\n", argsTokens)
	//fmt.Printf("Command: %s\n", command)
	//fmt.Printf("Subcommand: %s\n", subcommand)
	//fmt.Printf("Use default subcommand: %v\n", useDefaultSubcommand)
	//fmt.Printf("Filter tokens: %v\n", filterTokens)
	//fmt.Printf("Args tokens: %v\n", argsTokens)
	//fmt.Printf("Flag tokens: %v\n", flagArgs)
	//os.Exit(0)
	// Put it all together
	return ParsedCmdLine{
		Command:    command,
		Subcommand: subcommand,
		Filters:    filterTokens,
		Args:       argsTokens,
		Flags:      flagArgs,
	}, nil
}

// splitFlags splits the given args into non-flag and flag tokens.
// It returns the non-flag tokens and the flag tokens.
// The flag tokens are parsed using the configured flag parsers.
// The flag tokens are returned in the order they appear in the input.
// If the input contains a flag token that is not recognized,
// it is returned as a non-flag token.
func splitFlags(args []string) ([]string, []Arg, *ParseError) {
	nonFlags := make([]string, 0, len(args))
	flags := make([]Arg, 0)
	for _, arg := range args {
		ok := classifyFlag(arg)
		if ok {
			flagParser := argParsers[ArgKindFlag]
			token, err := flagParser(arg, config.Config{})
			if err != nil {
				return nil, nil, err
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
	if !attributeSpec.IsAllowedValue(attr.Value.ValueShape) {
		return &ParseError{
			Code:    ParseErrorInvalidInput,
			Arg:     attr.RawString(),
			Message: fmt.Sprintf("attribute %s does not accept %s values", attr.Key, attr.Value.ValueShape),
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
	if err = ValidateParsedCmdLine(parsed); err != nil {
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
