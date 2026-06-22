package parser

var addCommandSpec = CommandSpec{
	Name:              "add",
	DefaultSubcommand: "default",
	Subcommands: subcommands(
		SubcommandSpec{
			Name:  "default",
			Left:  emptySideSpec(),
			Right: addCommandRightSideSpec(),
		},
	),
}

var backupCommandSpec = CommandSpec{
	Name:              "backup",
	DefaultSubcommand: "now",
	Subcommands: subcommands(
		SubcommandSpec{
			Name: "now",
			Left: emptySideSpec(),
			Right: sideSpec([]ArgKind{ArgKindAttribute}, settableOnlyAttribute("output").SetShapes(AttributeValueShapeSingle)).
				WithArgs(0, 1).
				WithAttributeRule("output", atMostOne()),
		},
		SubcommandSpec{
			Name: "to",
			Left: emptySideSpec(),
			Right: sideSpec([]ArgKind{ArgKindAttribute}, settableOnlyAttribute("output").SetShapes(AttributeValueShapeSingle)).
				WithArgs(1, 1).
				WithAttributeRule("output", exactlyOne()),
		},
	),
}

var budgetCommandSpec = createSubcommandAlias(
	CommandSpec{
		Name:              "budget",
		DefaultSubcommand: "list",
		Subcommands: subcommands(
			SubcommandSpec{Name: "list", Left: budgetSideSpec(), Right: budgetSideSpec()},
			SubcommandSpec{Name: "add", Left: budgetSideSpec(), Right: budgetSideSpec()},
		),
	},
	[]SubcommandAlias{{"list", "ls"}},
)

var configCommandSpec = CommandSpec{
	Name:              "config",
	DefaultSubcommand: "print",
	Subcommands: subcommands(
		SubcommandSpec{Name: "print", Left: emptySideSpec(), Right: emptySideSpec().WithArgs(0, 0)},
		SubcommandSpec{Name: "get", Left: emptySideSpec(), Right: sideSpec([]ArgKind{ArgKindText}).WithArgs(1, 1).WithKindRule(ArgKindText, countRule(1, 1))},
		SubcommandSpec{Name: "set", Left: emptySideSpec(), Right: sideSpec([]ArgKind{ArgKindText}).WithArgs(2, 2).WithKindRule(ArgKindText, CountRule{Min: 2, Max: 2})},
	),
}

var deleteCommandSpec = CommandSpec{
	Name:              "delete",
	DefaultSubcommand: "transaction",
	Subcommands: subcommands(
		SubcommandSpec{Name: "transaction", Left: transactionFilterSideSpec(), Right: emptySideSpec().WithArgs(0, 0)},
		SubcommandSpec{Name: "list", Left: transactionFilterSideSpec(), Right: emptySideSpec().WithArgs(0, 0)},
	),
}

var helpCommandSpec = CommandSpec{
	Name:              "help",
	DefaultSubcommand: "show",
	Subcommands: subcommands(
		SubcommandSpec{Name: "show", Left: sideSpecAnything(), Right: sideSpec([]ArgKind{ArgKindText}).WithArgs(0, 1).WithKindRule(ArgKindText, atMostOne())},
	),
}

var importCommandSpec = CommandSpec{
	Name:              "import",
	DefaultSubcommand: "csv",
	Subcommands: subcommands(
		SubcommandSpec{Name: "csv", Left: emptySideSpec(), Right: sideSpec([]ArgKind{ArgKindText}).WithArgs(1, 1)},
	),
}

var modifyCommandSpec = CommandSpec{
	Name:              "modify",
	DefaultSubcommand: "transactions",
	Subcommands: subcommands(
		SubcommandSpec{
			Name:  "transactions",
			Left:  transactionFilterSideSpec(),
			Right: modifyTransactionRightSideSpec(),
		},
		SubcommandSpec{
			Name:  "accounts",
			Left:  accountFilterSingleAccountSideSpec(),
			Right: modifyAccountRightSideSpec(),
		},
	),
}

var purgeCommandSpec = CommandSpec{
	Name:              "purge",
	DefaultSubcommand: "default",
	Subcommands: subcommands(
		SubcommandSpec{Name: "default", Left: transactionFilterSideSpec(), Right: emptySideSpec()},
	),
}

var restoreCommandSpec = CommandSpec{
	Name:              "restore",
	DefaultSubcommand: "default",
	Subcommands: subcommands(
		SubcommandSpec{Name: "default", Left: emptySideSpec(), Right: sideSpec([]ArgKind{}).WithArgs(1, 1)},
		SubcommandSpec{Name: "list", Left: emptySideSpec(), Right: emptySideSpec().WithArgs(0, 0)},
	),
}

var summaryCommandSpec = CommandSpec{
	Name:              "summary",
	DefaultSubcommand: "",
	Subcommands: subcommands(
		SubcommandSpec{Name: "days", Left: transactionFilterSideSpec(), Right: transactionFilterSideSpec()},
	),
}

var themeCommandSpec = CommandSpec{
	Name:              "theme",
	DefaultSubcommand: "default",
	Subcommands: subcommands(
		SubcommandSpec{Name: "default", Left: emptySideSpec(), Right: sideSpec([]ArgKind{ArgKindText}).WithArgs(0, 1)},
		SubcommandSpec{Name: "list", Left: emptySideSpec(), Right: emptySideSpec().WithArgs(0, 0)},
		SubcommandSpec{Name: "set", Left: emptySideSpec(), Right: sideSpec([]ArgKind{ArgKindText}).WithArgs(1, 1)},
	),
}
