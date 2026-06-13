package cmd

import (
	"fmt"
	"strings"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/db"
	"github.com/nschaetti/cashwarrior/internal/parser"
)

func Budget(parsed parser.ParsedCmdLine, _ config.Config, query db.DBTX) error {
	_ = query
	fmt.Print(FormatBudgetDemo(parsed))
	return nil
}

func FormatBudgetDemo(parsed parser.ParsedCmdLine) string {
	var builder strings.Builder
	builder.WriteString("budget\n")
	builder.WriteString(fmt.Sprintf("command: %s\n", parsed.Command))
	builder.WriteString(fmt.Sprintf("subcommand: %s\n", parsed.Subcommand))
	builder.WriteString("left:\n")
	writeBudgetTokens(&builder, parsed.Left())
	builder.WriteString("right:\n")
	writeBudgetTokens(&builder, parsed.Right())
	return builder.String()
}

func writeBudgetTokens(builder *strings.Builder, args []parser.Arg) {
	if len(args) == 0 {
		builder.WriteString("  - none\n")
		return
	}

	for _, arg := range args {
		builder.WriteString(fmt.Sprintf("  - %s", arg.RawString()))
		attr, ok := arg.(parser.ArgAttribute)
		if ok {
			value, err := parser.ParseAttributeValue(attr.Value.Raw)
			if err == nil {
				builder.WriteString(fmt.Sprintf(" => %s", value.String()))
			}
		}
		builder.WriteString("\n")
	}
}
