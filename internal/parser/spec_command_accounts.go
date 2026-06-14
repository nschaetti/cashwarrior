package parser

func accountsListSubcommandSpec(subcommandName string) SubcommandSpec {
	return SubcommandSpec{
		Name: subcommandName,
		Left: sideSpec(
			[]ArgKind{ArgKindAttribute, ArgKindTag, ArgKindTagNegative},
			settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList),
			settableOnlyAttribute("name").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList),
			settableOnlyAttribute("currency").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList),
			settableOnlyAttribute("balance").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList|AttributeValueShapeRange),
		),
		Right: emptySideSpec().WithArgs(0, 0),
	}
}

var accountsCommandSpec = createSubcommandAlias(
	CommandSpec{
		Name:              "accounts",
		Description:       "List of accounts",
		DefaultSubcommand: "list",
		Subcommands: subcommands(
			accountsListSubcommandSpec("list"),
			SubcommandSpec{
				Name: "balance",
				Left: transactionFilterSideSpecWithoutAccount(),
				Right: sideSpec(
					[]ArgKind{ArgKindText},
					settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("name").SetShapes(AttributeValueShapeSingle),
				).
					WithAtLeastOneOf(PresenceRule{Kinds: []ArgKind{ArgKindText}, Attributes: []string{"account", "name"}, Message: "accounts balance requires an account name"}).
					WithArgs(1, 1),
			},
			SubcommandSpec{
				Name: "add",
				Left: emptySideSpec(),
				Right: sideSpec(
					[]ArgKind{ArgKindText, ArgKindAttribute},
					settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("currency").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("initial-balance").SetShapes(AttributeValueShapeSingle),
				).
					WithKindRule(ArgKindText, atMostOne()).
					WithArgs(1, 3).
					WithAttributeRule("currency", atMostOne()).
					WithAttributeRule("initial-balance", atMostOne()).
					WithAtLeastOneOf(PresenceRule{Kinds: []ArgKind{ArgKindText}, Attributes: []string{"account"}, Message: "accounts add requires an account name"}),
			},
			SubcommandSpec{
				Name: "modify",
				Left: sideSpec(
					[]ArgKind{ArgKindAttribute},
					settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle),
				),
				Right: sideSpec(
					[]ArgKind{ArgKindText, ArgKindAttribute},
					settableOnlyAttribute("name").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("currency").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("initial-balance").SetShapes(AttributeValueShapeSingle),
				).
					WithArgs(2, 3).
					WithKindRule(ArgKindText, atMostOne()).
					WithAttributeRule("name", atMostOne()).
					WithAttributeRule("currency", atMostOne()).
					WithAttributeRule("initial-balance", atMostOne()).
					WithAtLeastOneOf(PresenceRule{Kinds: []ArgKind{ArgKindText}, Attributes: []string{"name", "currency", "initial-balance"}, Message: "accounts modify requires at least one modification"}),
			},
			SubcommandSpec{
				Name: "initial-balance",
				Left: sideSpec(
					[]ArgKind{ArgKindAttribute},
					settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle),
				),
				Right: sideSpec(
					[]ArgKind{ArgKindAttribute},
					settableOnlyAttribute("amount").SetShapes(AttributeValueShapeSingle),
				).
					WithArgs(1, 1).
					WithAttributeRule("amount", exactlyOne()),
			},
			SubcommandSpec{
				Name: "rename",
				Left: sideSpec(
					[]ArgKind{ArgKindAttribute},
					settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle),
				),
				Right: sideSpec([]ArgKind{ArgKindText}).
					WithArgs(1, 1).
					WithKindRule(ArgKindText, countRule(1, 1)),
			},
			SubcommandSpec{
				Name: "delete",
				Left: sideSpec(
					[]ArgKind{ArgKindAttribute},
					settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle),
				),
				Right: emptySideSpec().WithArgs(0, 0),
			},
		),
	},
	[]SubcommandAlias{{"list", "ls"}, {"add", "create"}, {"modify", "update"}, {"initial-balance", "set-initial-balance"}, {"rename", "mv"}, {"delete", "rm"}},
)
