package parser

var showCommandSpec = CommandSpec{
	Name:              "show",
	DefaultSubcommand: "transaction",
	Subcommands: subcommands(
		SubcommandSpec{
			Name:  "transaction",
			Left:  emptySideSpec(),
			Right: sideSpec([]ArgKind{}).WithArgs(1, 1).WithAttributeRule("identifier", atMostOne()).WithAtLeastOneOf(PresenceRule{Kinds: []ArgKind{ArgKindAttribute}, Attributes: []string{"identifier"}, Message: "show requires an id"}),
		},
		categoriesListSubcommandSpec("categories"),
		accountsListSubcommandSpec("accounts"),
	),
}

var tagsCommandSpec = CommandSpec{
	Name:              "tags",
	DefaultSubcommand: "list",
	Subcommands: subcommands(
		SubcommandSpec{Name: "list", Left: emptySideSpec(), Right: emptySideSpec().WithArgs(0, 0)},
		SubcommandSpec{Name: "add", Left: emptySideSpec(), Right: sideSpec([]ArgKind{ArgKindText, ArgKindAttribute}, settableOnlyAttribute("tag").SetShapes(AttributeValueShapeSingle)).WithArgs(1, 1).WithAtLeastOneOf(PresenceRule{Kinds: []ArgKind{ArgKindText}, Attributes: []string{"tag"}, Message: "tags add requires a tag name"})},
		SubcommandSpec{Name: "modify", Left: emptySideSpec(), Right: sideSpec([]ArgKind{ArgKindText, ArgKindAttribute}, settableOnlyAttribute("tag").SetShapes(AttributeValueShapeSingle)).WithKindRule(ArgKindText, exactlyOne()).WithArgs(2, 2).WithAttributeRule("tag", exactlyOne()).WithAtLeastOneOf(PresenceRule{Attributes: []string{"tag"}, Message: "tags modify requires a new tag name"})},
		SubcommandSpec{Name: "delete", Left: emptySideSpec(), Right: sideSpec([]ArgKind{ArgKindText, ArgKindAttribute}, settableOnlyAttribute("tag").SetShapes(AttributeValueShapeSingle)).WithArgs(1, 1).WithAtLeastOneOf(PresenceRule{Kinds: []ArgKind{ArgKindText}, Attributes: []string{"tag"}, Message: "tags delete requires a tag name"})},
	),
}

var storesCommandSpec = createSubcommandAlias(
	CommandSpec{
		Name:              "stores",
		DefaultSubcommand: "list",
		Subcommands: subcommands(
			SubcommandSpec{Name: "list", Left: emptySideSpec(), Right: emptySideSpec().WithArgs(0, 0)},
			SubcommandSpec{Name: "add", Left: emptySideSpec(), Right: sideSpec([]ArgKind{ArgKindText}).WithArgs(1, 1)},
			SubcommandSpec{Name: "rename", Left: emptySideSpec(), Right: sideSpec([]ArgKind{ArgKindText}).WithArgs(2, 2).WithKindRule(ArgKindText, countRule(2, 2))},
			SubcommandSpec{Name: "delete", Left: emptySideSpec(), Right: sideSpec([]ArgKind{ArgKindText}).WithArgs(1, 1).WithKindRule(ArgKindText, exactlyOne())},
		),
	},
	[]SubcommandAlias{{"list", "ls"}, {"rename", "rn"}, {"delete", "rm"}},
)

var transferCommandSpec = CommandSpec{
	Name:              "transfer",
	DefaultSubcommand: "add",
	Subcommands: subcommands(
		SubcommandSpec{Name: "add", Left: emptySideSpec(), Right: transferRightSideSpec()},
		SubcommandSpec{Name: "delete", Left: emptySideSpec(), Right: sideSpec([]ArgKind{ArgKindAttribute}).WithArgs(1, 1).WithKindRule(ArgKindAttribute, exactlyOne()).WithAttributeRule("identifier", atMostOne()).WithAtLeastOneOf(PresenceRule{Kinds: []ArgKind{ArgKindAttribute}, Attributes: []string{"identifier"}, Message: "transfer delete requires an id"})},
		SubcommandSpec{Name: "list", Left: emptySideSpec(), Right: emptySideSpec().WithArgs(0, 0)},
	),
}
