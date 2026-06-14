package parser

var fakeitCommandSpec = CommandSpec{
	Name:              "fake-it",
	DefaultSubcommand: "transactions",
	Subcommands: subcommands(
		SubcommandSpec{Name: "transactions", Left: emptySideSpec(), Right: fakeitTransactionsRightSideSpec()},
		SubcommandSpec{Name: "categories", Left: emptySideSpec(), Right: fakeitCategoriesRightSideSpec()},
		SubcommandSpec{Name: "accounts", Left: emptySideSpec(), Right: fakeitAccountsRightSideSpec()},
		SubcommandSpec{Name: "stores", Left: emptySideSpec(), Right: fakeitStoresRightSideSpec()},
		SubcommandSpec{Name: "groups", Left: emptySideSpec(), Right: fakeitGroupsRightSideSpec()},
		SubcommandSpec{Name: "tags", Left: emptySideSpec(), Right: fakeitTagsRightSideSpec()},
	),
}
