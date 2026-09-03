package parser

var balanceCommandSpec = CommandSpec{
	Name:              "balance",
	Description:       "Show account balances for a period",
	DefaultSubcommand: "default",
	Subcommands: subcommands(SubcommandSpec{
		Name: "default",
		Left: sideSpec(
			[]ArgKind{ArgKindText, ArgKindAttribute},
			buildAttributeSpec("account").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList),
			buildAttributeSpec("currency").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList),
			buildAttributeSpec("date").SetShapes(AttributeValueShapeSingle|AttributeValueShapeRange|AttributeValueShapeShortcut),
		).
			WithAttributeRule("account", atMostOne()).
			WithAttributeRule("currency", atMostOne()).
			WithAttributeRule("date", atMostOne()),
		Right: emptySideSpec().WithArgs(0, 0),
	}),
}
