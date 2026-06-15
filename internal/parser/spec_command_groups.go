package parser

var groupsCommandSpec = createSubcommandAlias(
	CommandSpec{
		Name:              "groups",
		DefaultSubcommand: "list",
		Subcommands: subcommands(
			SubcommandSpec{
				Name: "list",
				Left: sideSpec(
					[]ArgKind{ArgKindAttribute},
					settableOnlyAttribute("order").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("desc").SetShapes(AttributeValueShapeSingle),
				),
				Right: emptySideSpec().WithArgs(0, 0),
			},
			SubcommandSpec{
				Name: "add",
				Left: emptySideSpec(),
				Right: sideSpec([]ArgKind{ArgKindText, ArgKindAttribute}).WithArgs(2, 0).
					WithKindRule(ArgKindText, atMostOne()).
					WithAttributeRule("group", atMostOne()).
					WithAtLeastOneOf(PresenceRule{Kinds: []ArgKind{ArgKindText}, Attributes: []string{"group"}, Message: "groups add requires a group name"}).
					WithAtLeastOneOf(PresenceRule{Kinds: []ArgKind{ArgKindAttribute}, Attributes: []string{"id", "identifier", "T"}, Message: "groups add requires at least one transaction id"}),
			},
			SubcommandSpec{
				Name: "modify",
				Left: emptySideSpec(),
				Right: sideSpec([]ArgKind{ArgKindText, ArgKindAttribute}).WithArgs(2, 2).
					WithKindRule(ArgKindAttribute, atMostOne()).
					WithKindRule(ArgKindText, atMostOne()),
			},
			SubcommandSpec{Name: "delete", Left: emptySideSpec(), Right: sideSpec([]ArgKind{ArgKindText}).WithArgs(1, 1).WithKindRule(ArgKindText, exactlyOne())},
			SubcommandSpec{
				Name: "remove",
				Left: emptySideSpec(),
				Right: sideSpec([]ArgKind{ArgKindAttribute}).WithArgs(2, 2).
					WithAttributeRule("id", atMostOne()).
					WithAttributeRule("identifier", atMostOne()).
					WithAttributeRule("T", atMostOne()).
					WithAttributeRule("group", exactlyOne()).
					WithAtLeastOneOf(PresenceRule{Kinds: []ArgKind{ArgKindAttribute}, Attributes: []string{"id", "identifier", "T", "group"}, Message: "groups remove requires at least one transaction id"}),
			},
		),
	},
	[]SubcommandAlias{{SubcommandName: "list", SubcommandAlias: "ls"}, {SubcommandName: "modify", SubcommandAlias: "rename"}, {SubcommandName: "modify", SubcommandAlias: "rn"}, {SubcommandName: "delete", SubcommandAlias: "rm"}},
)

var groupCommandSpec = CommandSpec{
	Name:              "group",
	DefaultSubcommand: "add",
	Subcommands: subcommands(
		SubcommandSpec{
			Name: "add",
			Left: emptySideSpec(),
			Right: sideSpec(
				[]ArgKind{ArgKindAttribute},
				settableOnlyAttribute("group").SetShapes(AttributeValueShapeSingle),
				settableOnlyAttribute("id").SetShapes(AttributeValueShapeSingle),
				settableOnlyAttribute("identifier").SetShapes(AttributeValueShapeSingle),
				settableOnlyAttribute("T").SetShapes(AttributeValueShapeSingle),
			).
				WithArgs(2, 0).
				WithAttributeRule("group", exactlyOne()).
				WithAtLeastOneOf(PresenceRule{Kinds: []ArgKind{ArgKindAttribute}, Attributes: []string{"id", "identifier", "T"}, Message: "groups add requires at least one transaction id"}),
		},
	),
}
