package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Help(parsed parser.ParsedCmdLine, config config.Config, db db.DBTX) error {
	printGlobalHelp()
	return nil
}

// printHelp prints the help for the given command.
func printHelp(parsed parser.ParsedCmdLine) {
	if parsed.Command == "" {
		printGlobalHelp()
		return
	}

	commandSpec, ok := parser.GetCommandSpec(parsed.Command)
	if !ok {
		printGlobalHelp()
		return
	}

	printCommandHelp(commandSpec)
	if parsed.Subcommand == "" {
		return
	}
	if subSpec, ok := commandSpec.Subcommands[parsed.Subcommand]; ok {
		printSubcommandHelp(commandSpec.Name, parsed.Subcommand, subSpec)
	}
}

// printGlobalHelp prints the global help.
func printGlobalHelp() {
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  cash <command> [subcommand] [arguments] [--help|-h]")
	fmt.Println()
	fmt.Println("Commands:")
	for _, name := range parser.KnownCommands() {
		fmt.Printf("  %s\n", name)
	}
	fmt.Println()
	fmt.Println("Tip:")
	fmt.Println("  cash <command> --help")
	fmt.Println("  cash <command> <subcommand> --help")
	fmt.Println()
}

func printCommandHelp(spec parser.CommandSpec) {
	fmt.Printf("%s\n", spec.Description)
	fmt.Println()
	fmt.Printf("Usage:\n  cash [filters] %s [subcommand] [arguments] [--help]\n", spec.Name)
	fmt.Println()
	fmt.Println("Subcommands:")
	names := make([]string, 0, len(spec.Subcommands))
	for name, comSpec := range spec.Subcommands {
		if !comSpec.IsAlias {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		commandSpec := spec.Subcommands[name]
		mark := ""
		if name == spec.DefaultSubcommand {
			mark = " (default)"
		}
		fmt.Printf("  %s%s : %s\n", name, mark, commandSpec.Description)
	}
	fmt.Println()
	fmt.Println("Tip:")
	fmt.Printf("  cash %s --help\n", spec.Name)
	fmt.Printf("  cash %s <subcommand> --help\n", spec.Name)
	fmt.Println()
}

func printSubcommandHelp(command string, subcommand string, subSpec parser.SubcommandSpec) {
	// Has filter ?
	filterText := ""
	if subSpec.HasFilter() {
		filterText = "[filters] "
	}

	// Has arguments ?
	argumentText := ""
	if subSpec.HasArgument() {
		argumentText = " [arguments]"
	}

	// Print
	fmt.Printf("Subcommand help:\n  cash %s%s %s%s [--help|-h]\n", filterText, command, subcommand, argumentText)
	fmt.Println()
	printSideHelp("Filters", subSpec.Left)
	printSideHelp("Arguments", subSpec.Right)
	printExamples(command, subcommand)
	fmt.Println("Tip:")
	fmt.Println("  --help and -h can be placed anywhere.")
	fmt.Println()
}

func printSideHelp(title string, side parser.SideSpec) {
	fmt.Println(title + ":")

	// Empty, print 'none'
	if side.IsEmpty() {
		fmt.Println("  none")
		fmt.Println()
		return
	}

	// List allowed tokens
	kinds := make([]string, 0, len(side.AllowedKinds))
	for kind, allowed := range side.AllowedKinds {
		if allowed {
			kinds = append(kinds, kind.String())
		}
	}
	sort.Strings(kinds)
	fmt.Printf("  Accept: %s\n", strings.Join(kinds, ", "))

	if side.MinArgs > 0 || side.MaxArgs > 0 {
		mmax := "unbounded"
		if side.MaxArgs > 0 {
			mmax = fmt.Sprintf("%d", side.MaxArgs)
		}
		fmt.Printf("  count: min=%d max=%s\n", side.MinArgs, mmax)
	}

	if side.AllowAnyAttr {
		fmt.Println("  attributes: any")
	} else if len(side.Attributes) == 0 {
		fmt.Println("  attributes: none")
	} else {
		names := make([]string, 0, len(side.Attributes))
		for name := range side.Attributes {
			names = append(names, name)
		}
		sort.Strings(names)
		items := make([]string, 0, len(names))
		for _, name := range names {
			spec := side.Attributes[name]
			modes := make([]string, 0, 2)
			if spec.Settable {
				modes = append(modes, "set")
			}
			if spec.Clearable {
				modes = append(modes, "clear")
			}
			items = append(items, fmt.Sprintf("%s(%s)", name, strings.Join(modes, "/")))
		}
		fmt.Printf("  attributes: %s\n", strings.Join(items, ", "))
	}
	fmt.Println()
}

func printExamples(command string, subcommand string) {
	examples := helpExamples[command+" "+subcommand]
	if len(examples) == 0 {
		return
	}
	fmt.Println("Examples:")
	for _, ex := range examples {
		fmt.Printf("  %s\n", ex)
	}
	fmt.Println()
}

var helpExamples = map[string][]string{
	"accounts add": {
		"cash accounts add bcv currency:CHF",
		"cash accounts add savings initial_balance:1500",
	},
	"accounts initial-balance": {
		"cash accounts initial-balance bcv 1200",
		"cash accounts initial-balance 1200 bcv",
	},
	"accounts balance": {
		"cash accounts balance main",
		"cash today accounts balance main",
	},
	"add default": {
		"cash add -12.50 Coffee store:coop category:food",
		"cash today add -20 Lunch",
	},
	"modify default": {
		"cash 2026.05.12 modify category:groceries",
		"cash today modify desc:Lunch",
	},
	"list default": {
		"cash list",
		"cash month list account:main category:food",
	},
	"import default": {
		"cash import ./transactions.csv",
	},
	"transfer default": {
		"cash transfer +250 from:main to:savings",
		"cash transfer +75.5 from:cash to:joint date:2026-05-31",
	},
	"summary days": {
		"cash summary days",
		"cash month summary days account:main",
		"cash summary days date:2026-05-01..2026-05-31",
	},
	"groups list": {
		"cash groups",
		"cash groups list",
		"cash order:start_date groups",
		"cash order:end_date desc:true groups",
	},
	"places list": {
		"cash places",
		"cash places list",
	},
	"places rename": {
		"cash places rename Coop Migros",
		"cash places rename \"Coop Pronto Nyon Gare\" \"Coop Pronto Nyon\"",
	},
}
