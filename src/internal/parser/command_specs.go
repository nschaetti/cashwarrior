package parser

import "sort"

// CommandSpec defines parser-level validation rules for a command.
type CommandSpec struct {
	Name               string
	MinArgs            int
	MaxArgs            int
	AllowedFilterKinds map[TokenKind]bool
	AllowedArgKinds    map[TokenKind]bool
}

func allowKinds(kinds ...TokenKind) map[TokenKind]bool {
	allowed := make(map[TokenKind]bool, len(kinds))
	for _, kind := range kinds {
		allowed[kind] = true
	}
	return allowed
}

var allFilterKinds = allowKinds(
	TokenAmount,
	TokenTag,
	TokenTagNegative,
	TokenAttribute,
	TokenAttributeClear,
	TokenID,
	TokenPeriod,
	TokenText,
)

var allArgKinds = allowKinds(
	TokenAmount,
	TokenTag,
	TokenTagNegative,
	TokenAttribute,
	TokenAttributeClear,
	TokenID,
	TokenPeriod,
	TokenText,
)

var commandSpecs = map[string]CommandSpec{
	"init":        {Name: "init", MinArgs: 0, MaxArgs: 0, AllowedFilterKinds: allFilterKinds, AllowedArgKinds: allArgKinds},
	"add":         {Name: "add", MinArgs: 0, MaxArgs: -1, AllowedFilterKinds: allFilterKinds, AllowedArgKinds: allArgKinds},
	"show":        {Name: "show", MinArgs: 0, MaxArgs: -1, AllowedFilterKinds: allFilterKinds, AllowedArgKinds: allArgKinds},
	"categories":  {Name: "categories", MinArgs: 0, MaxArgs: -1, AllowedFilterKinds: allFilterKinds, AllowedArgKinds: allArgKinds},
	"stats":       {Name: "stats", MinArgs: 0, MaxArgs: -1, AllowedFilterKinds: allFilterKinds, AllowedArgKinds: allArgKinds},
	"tags":        {Name: "tags", MinArgs: 0, MaxArgs: -1, AllowedFilterKinds: allFilterKinds, AllowedArgKinds: allArgKinds},
	"modify":      {Name: "modify", MinArgs: 0, MaxArgs: -1, AllowedFilterKinds: allFilterKinds, AllowedArgKinds: allArgKinds},
	"report":      {Name: "report", MinArgs: 0, MaxArgs: -1, AllowedFilterKinds: allFilterKinds, AllowedArgKinds: allArgKinds},
	"list":        {Name: "list", MinArgs: 0, MaxArgs: -1, AllowedFilterKinds: allFilterKinds, AllowedArgKinds: allArgKinds},
	"delete":      {Name: "delete", MinArgs: 0, MaxArgs: -1, AllowedFilterKinds: allFilterKinds, AllowedArgKinds: allArgKinds},
	"undo":        {Name: "undo", MinArgs: 0, MaxArgs: -1, AllowedFilterKinds: allFilterKinds, AllowedArgKinds: allArgKinds},
	"by":          {Name: "by", MinArgs: 0, MaxArgs: -1, AllowedFilterKinds: allFilterKinds, AllowedArgKinds: allArgKinds},
	"accounts":    {Name: "accounts", MinArgs: 0, MaxArgs: -1, AllowedFilterKinds: allFilterKinds, AllowedArgKinds: allArgKinds},
	"balance":     {Name: "balance", MinArgs: 0, MaxArgs: -1, AllowedFilterKinds: allFilterKinds, AllowedArgKinds: allArgKinds},
	"transfer":    {Name: "transfer", MinArgs: 0, MaxArgs: -1, AllowedFilterKinds: allFilterKinds, AllowedArgKinds: allArgKinds},
	"set-balance": {Name: "set-balance", MinArgs: 0, MaxArgs: -1, AllowedFilterKinds: allFilterKinds, AllowedArgKinds: allArgKinds},
	"budget":      {Name: "budget", MinArgs: 0, MaxArgs: -1, AllowedFilterKinds: allFilterKinds, AllowedArgKinds: allArgKinds},
	"config":      {Name: "config", MinArgs: 0, MaxArgs: -1, AllowedFilterKinds: allFilterKinds, AllowedArgKinds: allArgKinds},
	"theme":       {Name: "theme", MinArgs: 0, MaxArgs: -1, AllowedFilterKinds: allFilterKinds, AllowedArgKinds: allArgKinds},
	"sum":         {Name: "sum", MinArgs: 0, MaxArgs: -1, AllowedFilterKinds: allFilterKinds, AllowedArgKinds: allArgKinds},
}

// IsKnownCommand reports whether the command is registered.
func IsKnownCommand(command string) bool {
	_, ok := commandSpecs[command]
	return ok
}

// GetCommandSpec returns the spec for a command.
func GetCommandSpec(command string) (CommandSpec, bool) {
	spec, ok := commandSpecs[command]
	return spec, ok
}

// KnownCommands returns all registered command names in sorted order.
func KnownCommands() []string {
	commands := make([]string, 0, len(commandSpecs))
	for name := range commandSpecs {
		commands = append(commands, name)
	}
	sort.Strings(commands)
	return commands
}
