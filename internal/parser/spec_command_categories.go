package parser

// categoriesListSubcommandSpec creates a subcommand for categories listening
func categoriesListSubcommandSpec(commandName string) SubcommandSpec {
	return SubcommandSpec{
		Name:  commandName,
		Left:  emptySideSpec(),
		Right: emptySideSpec().WithArgs(0, 0),
	}
}

var categoriesCommandSpec = createSubcommandAlias(
	CommandSpec{
		Name:              "categories",
		DefaultSubcommand: "list",
		Subcommands: subcommands(
			categoriesListSubcommandSpec("list"),
			SubcommandSpec{
				Name: "add",
				Left: emptySideSpec(),
				Right: sideSpec(
					[]ArgKind{ArgKindText, ArgKindAttribute},
					settableOnlyAttribute("category").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("parent").SetShapes(AttributeValueShapeSingle),
				).
					WithArgs(1, 2).
					WithAttributeRule("parent", atMostOne()).
					WithAtLeastOneOf(PresenceRule{Kinds: []ArgKind{ArgKindText}, Attributes: []string{"category"}, Message: "categories add requires a category name"}),
			},
			SubcommandSpec{
				Name: "modify",
				Left: emptySideSpec(),
				Right: sideSpec(
					[]ArgKind{ArgKindText, ArgKindAttribute},
					settableOnlyAttribute("category").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("parent").SetShapes(AttributeValueShapeSingle),
				).
					WithArgs(2, 3).
					WithKindRule(ArgKindText, exactlyOne()).
					WithAttributeRule("category", atMostOne()).
					WithAttributeRule("parent", atMostOne()).
					WithAtLeastOneOf(PresenceRule{Attributes: []string{"category", "parent"}, Message: "categories modify requires at least one modification"}),
			},
			SubcommandSpec{
				Name: "delete",
				Left: emptySideSpec(),
				Right: sideSpec(
					[]ArgKind{ArgKindText, ArgKindAttribute},
					settableOnlyAttribute("category").SetShapes(AttributeValueShapeSingle),
				).
					WithArgs(1, 1).
					WithAttributeRule("category", atMostOne()).
					WithAtLeastOneOf(PresenceRule{Kinds: []ArgKind{ArgKindText}, Attributes: []string{"category"}, Message: "categories delete requires a category name"}),
			},
		),
	},
	[]SubcommandAlias{{"list", "ls"}, {"modify", "update"}, {"delete", "rm"}},
)
