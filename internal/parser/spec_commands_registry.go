package parser

import "sort"

// CommandSpecs is a map of command names to their specifications.
var CommandSpecs = map[string]CommandSpec{
	"accounts":    accountsCommandSpec,
	"add":         addCommandSpec,
	"backup":      backupCommandSpec,
	"balance":     defaultCommandSpec("balance"),
	"budget":      budgetCommandSpec,
	"by":          defaultCommandSpec("by"),
	"categories":  categoriesCommandSpec,
	"config":      configCommandSpec,
	"delete":      deleteCommandSpec,
	"export":      defaultCommandSpec("export"),
	"fakeit":      fakeitCommandSpec,
	"groups":      groupsCommandSpec,
	"help":        helpCommandSpec,
	"import":      importCommandSpec,
	"list":        listCommandSpec,
	"modify":      modifyCommandSpec,
	"stores":      storesCommandSpec,
	"purge":       purgeCommandSpec,
	"report":      defaultCommandSpec("report"),
	"restore":     restoreCommandSpec,
	"set-balance": defaultCommandSpec("set-balance"),
	"show":        showCommandSpec,
	"stats":       defaultCommandSpec("stats"),
	"sum":         defaultCommandSpec("sum"),
	"summary":     summaryCommandSpec,
	"tags":        tagsCommandSpec,
	"theme":       themeCommandSpec,
	"transfer":    transferCommandSpec,
	"undo":        defaultCommandSpec("undo"),
}

func GetCommandSpec(name string) (CommandSpec, bool) {
	spec, ok := CommandSpecs[name]
	return spec, ok
}

func IsKnownCommand(command string) bool {
	_, ok := CommandSpecs[command]
	return ok
}

func KnownCommands() []string {
	commands := make([]string, 0, len(CommandSpecs))
	for name := range CommandSpecs {
		commands = append(commands, name)
	}
	sort.Strings(commands)
	return commands
}

func GetSubcommandSpec(command string, subcommand string) (SubcommandSpec, bool) {
	commandSpec, ok := GetCommandSpec(command)
	if !ok {
		return SubcommandSpec{}, false
	}
	subcommandSpec, ok := commandSpec.Subcommands[subcommand]
	return subcommandSpec, ok
}
