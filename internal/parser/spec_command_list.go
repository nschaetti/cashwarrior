package parser

var listCommandSpec = createSubcommandAlias(
	CommandSpec{
		Name:              "list",
		Description:       "List elements such as transactions, accounts, etc.",
		DefaultSubcommand: "transactions",
		Subcommands: subcommands(
			SubcommandSpec{
				Name:        "transactions",
				Description: "List transactions according to date, account, etc.",
				Left: transactionFilterSideSpec(
					settableOnlyAttribute("order").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("desc").SetShapes(AttributeValueShapeSingle),
				),
				Right: genericSideSpec(),
			},
			SubcommandSpec{
				Name:        "accounts",
				Description: "List accounts according to currency, balance, etc.",
				Left: accountFilterSideSpec(
					settableOnlyAttribute("order").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("desc").SetShapes(AttributeValueShapeSingle),
				),
			},
			SubcommandSpec{
				Name:        "groups",
				Description: "List groups of transactions.",
				Left: groupFilterSideSpec(
					settableOnlyAttribute("order").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("desc").SetShapes(AttributeValueShapeSingle),
				),
			},
			SubcommandSpec{
				Name:        "tags",
				Description: "List tags of transactions.",
				Left: tagFilterSideSpec(
					settableOnlyAttribute("order").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("desc").SetShapes(AttributeValueShapeSingle),
				),
			},
		),
	},
	[]SubcommandAlias{{"transactions", "t"}, {"accounts", "a"}, {"groups", "g"}, {"tags", "ta"}},
)
