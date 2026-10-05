package parser

// countRule creates a CountRule with the given minimum and maximum values.
func countRule(min int, max int) CountRule {
	return CountRule{Min: min, Max: max}
}

// exactlyOne creates a CountRule with eactly one occurrence.
func exactlyOne() CountRule {
	return CountRule{Min: 1, Max: 1}
}

// exactlyZero creates a CountRule with exactly zero occurrences.
func atLeastOne() CountRule {
	return CountRule{Min: 1}
}

// atMostOne creates a CountRule with at most one occurrence.
func atMostOne() CountRule {
	return CountRule{Max: 1}
}

// allowKinds returns a map with TokenKind as key and true as value
func allowKinds(kinds ...ArgKind) map[ArgKind]bool {
	allowed := make(map[ArgKind]bool, len(kinds))
	for _, kind := range kinds {
		allowed[kind] = true
	}
	return allowed
}

// attributes maps attribute names to attribute specs
func attributes(specs ...AttributeSpec) map[string]AttributeSpec {
	attrs := make(map[string]AttributeSpec, len(specs))
	for _, spec := range specs {
		attrs[spec.Name] = spec
	}
	return attrs
}

// subcommands maps subcommand names to subcommand specs
func subcommands(specs ...SubcommandSpec) map[string]SubcommandSpec {
	items := make(map[string]SubcommandSpec, len(specs))
	for _, spec := range specs {
		items[spec.Name] = spec
	}
	return items
}

// allSupportedKinds returns supported all token kinds as supported
func allSupportedKinds() map[ArgKind]bool {
	return allowKinds(
		ArgKindTag,
		ArgKindAttribute,
		ArgKindText,
		ArgKindTagNegative,
	)
}

// createSubcommandAlias creates a subcommand alias for a command.
// This is useful for creating aliases for subcommands.
func createSubcommandAlias(command CommandSpec, aliases []SubcommandAlias) CommandSpec {
	for _, alias := range aliases {
		subCommand, ok := command.Subcommands[alias.SubcommandName]
		if !ok {
			panic("subcommand " + alias.SubcommandName + " not found in command " + command.Name)
		}
		subCommand.IsAlias = true
		command.Subcommands[alias.SubcommandAlias] = subCommand
	}
	return command
}

func defaultCommandSpec(name string) CommandSpec {
	return CommandSpec{
		Name:              name,
		Description:       defaultCommandDescriptions[name],
		DefaultSubcommand: "default",
		Subcommands: subcommands(SubcommandSpec{
			Name:        "default",
			Description: defaultSubcommandDescriptions[name],
			Left:        genericSideSpec(),
			Right:       genericSideSpec(),
		}),
	}
}

var defaultCommandDescriptions = map[string]string{
	"balance":     "Show account balances",
	"by":          "Group or analyze transactions",
	"export":      "Export financial data",
	"init":        "Initialize the Cashwarrior database",
	"report":      "Generate a financial report",
	"set-balance": "Set an account balance",
	"stats":       "Show transaction statistics",
	"sum":         "Calculate transaction totals",
	"undo":        "Undo the last operation",
}

var defaultSubcommandDescriptions = map[string]string{
	"balance":     "Show balances",
	"by":          "Group transactions",
	"export":      "Export data",
	"init":        "Initialize the database",
	"report":      "Generate a report",
	"set-balance": "Set a balance",
	"stats":       "Show statistics",
	"sum":         "Calculate totals",
	"undo":        "Undo the last operation",
}
