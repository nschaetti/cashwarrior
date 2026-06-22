package parser

// genericSideSpec specifies a side that accepts any token kind and any attribute.
func genericSideSpec() SideSpec {
	return SideSpec{
		AllowedKinds:   allSupportedKinds(),
		AllowAnyAttr:   true,
		KindRules:      map[ArgKind]CountRule{},
		AttributeRules: map[string]CountRule{},
		AtLeastOneOf:   []PresenceRule{},
	}
}

// emptySideSpec specifies a side that accepts no tokens.
func emptySideSpec() SideSpec {
	return SideSpec{
		AllowedKinds:   map[ArgKind]bool{},
		Attributes:     map[string]AttributeSpec{},
		KindRules:      map[ArgKind]CountRule{},
		AttributeRules: map[string]CountRule{},
		AtLeastOneOf:   []PresenceRule{},
	}
}

// sideSpec specifies a side that accepts a list of token kinds and a list of attributes.
func sideSpec(kinds []ArgKind, attrs ...AttributeSpec) SideSpec {
	return SideSpec{
		AllowedKinds:   allowKinds(kinds...),
		Attributes:     attributes(attrs...),
		KindRules:      map[ArgKind]CountRule{},
		AttributeRules: map[string]CountRule{},
		AtLeastOneOf:   []PresenceRule{},
	}
}

func sideSpecAnything() SideSpec {
	return SideSpec{
		AllowedKinds:   allSupportedKinds(),
		AllowAnyAttr:   true,
		KindRules:      map[ArgKind]CountRule{},
		AttributeRules: map[string]CountRule{},
		AtLeastOneOf:   []PresenceRule{},
	}
}

// sideSpecWithAttributes specifies a side that accepts a list of token kinds and a list of attributes.
func sideSpecWithAnyAttributes(kinds ...ArgKind) SideSpec {
	return SideSpec{
		AllowedKinds:   allowKinds(kinds...),
		AllowAnyAttr:   true,
		KindRules:      map[ArgKind]CountRule{},
		AttributeRules: map[string]CountRule{},
		AtLeastOneOf:   []PresenceRule{},
	}
}

// transactionFilterSideSpec specifies the left side as a filter for transactions.
func transactionFilterSideSpec(extraAttrs ...AttributeSpec) SideSpec {
	base := []AttributeSpec{
		settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
		settableOnlyAttribute("category").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
		settableOnlyAttribute("date").SetShapes(AttributeValueShapeSingle | AttributeValueShapeRange | AttributeValueShapeList | AttributeValueShapeShortcut),
		settableOnlyAttribute("desc").SetShapes(AttributeValueShapeSingle),
		settableOnlyAttribute("store").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
		settableOnlyAttribute("identifier").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange),
		settableOnlyAttribute("group").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
	}
	base = append(base, extraAttrs...)
	return sideSpec([]ArgKind{ArgKindTag, ArgKindTagNegative, ArgKindAttribute}, base...)
}

// transactionFilterSideSpecWithoutAccount specifies the left side as a filter for transactions.
func transactionFilterSideSpecWithoutAccount(extraAttrs ...AttributeSpec) SideSpec {
	base := []AttributeSpec{
		settableOnlyAttribute("category").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
		settableOnlyAttribute("date").SetShapes(AttributeValueShapeSingle | AttributeValueShapeRange | AttributeValueShapeList),
		settableOnlyAttribute("desc").SetShapes(AttributeValueShapeSingle),
		settableOnlyAttribute("store").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
		settableOnlyAttribute("identifier").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange),
		settableOnlyAttribute("group").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
	}
	return sideSpec([]ArgKind{ArgKindAttribute}, base...)
}

// accountFilterSideSpec specifies the left side as a filter for accounts.
func accountFilterSideSpec(extraAttrs ...AttributeSpec) SideSpec {
	base := []AttributeSpec{
		settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
		settableOnlyAttribute("currency").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
		settableOnlyAttribute("initial-balance").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange),
		settableOnlyAttribute("balance").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange),
	}
	base = append(base, extraAttrs...)
	return sideSpec([]ArgKind{ArgKindAttribute}, base...)
}

// groupFilterSideSpec specifies the left side as a filter for groups.
func groupFilterSideSpec(extraAttrs ...AttributeSpec) SideSpec {
	base := []AttributeSpec{
		settableOnlyAttribute("group").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
		settableOnlyAttribute("size").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange),
	}
	base = append(base, extraAttrs...)
	return sideSpec([]ArgKind{ArgKindText, ArgKindAttribute}, base...)
}

// tagFilterSideSpec specifies the left side as a filter for tags.
func tagFilterSideSpec(extraAttrs ...AttributeSpec) SideSpec {
	base := []AttributeSpec{
		buildAttributeSpec("tag").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
		buildAttributeSpec("size").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange),
	}
	base = append(base, extraAttrs...)
	return sideSpec([]ArgKind{ArgKindTag, ArgKindAttribute}, base...)
}

// placesFilterSideSpec specifies the left side as a filter for places.
func placesFilterSideSpec(extraAttrs ...AttributeSpec) SideSpec {
	base := []AttributeSpec{
		buildAttributeSpec("place").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
		buildAttributeSpec("size").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange),
	}
	base = append(base, extraAttrs...)
	return sideSpec([]ArgKind{ArgKindText, ArgKindAttribute}, base...)
}

// addCommandRightSideSpec specifies the right side of the add command.
func addCommandRightSideSpec() SideSpec {
	side := sideSpec(
		[]ArgKind{ArgKindTag, ArgKindAttribute, ArgKindText},
		settableOnlyAttribute("date").SetShapes(AttributeValueShapeSingle|AttributeValueShapeShortcut),
		settableOnlyAttribute("store").SetShapes(AttributeValueShapeSingle),
		settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle),
		settableOnlyAttribute("amount").SetShapes(AttributeValueShapeSingle),
		setOrClearAttribute("category").SetShapes(AttributeValueShapeSingle),
		setOrClearAttribute("group").SetShapes(AttributeValueShapeSingle),
	).WithArgs(2, 0).
		WithAttributeRule("amount", exactlyOne()).
		WithAttributeRule("store", exactlyOne())
	for _, name := range []string{"date", "account", "category", "group"} {
		side = side.WithAttributeRule(name, atMostOne())
	}
	side = side.WithAtLeastOneOf(PresenceRule{Kinds: []ArgKind{ArgKindText}, Message: "add requires a description"})
	return side
}

func modifyRightSideSpec() SideSpec {
	side := sideSpec(
		[]ArgKind{ArgKindAttribute, ArgKindTag, ArgKindTagNegative},
		settableOnlyAttribute("identifier").SetShapes(AttributeValueShapeSingle),
		settableOnlyAttribute("amount").SetShapes(AttributeValueShapeSingle),
		settableOnlyAttribute("description").SetShapes(AttributeValueShapeSingle),
		settableOnlyAttribute("date").SetShapes(AttributeValueShapeSingle),
		settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle),
		setOrClearAttribute("category").SetShapes(AttributeValueShapeSingle),
		settableOnlyAttribute("store").SetShapes(AttributeValueShapeSingle),
		setOrClearAttribute("group").SetShapes(AttributeValueShapeSingle),
	).WithArgs(1, 0)
	for _, name := range []string{"identifier", "amount", "desc", "date", "time", "datetime", "account", "category", "store", "group"} {
		side = side.WithAttributeRule(name, atMostOne())
	}
	return side
}

func transferRightSideSpec() SideSpec {
	side := sideSpec(
		[]ArgKind{ArgKindAttribute, ArgKindText},
		settableOnlyAttribute("from").SetShapes(AttributeValueShapeSingle),
		settableOnlyAttribute("to").SetShapes(AttributeValueShapeSingle),
		settableOnlyAttribute("date").SetShapes(AttributeValueShapeSingle),
	)
	side.WithAttributeRule("amount", exactlyOne()).
		WithAttributeRule("from", exactlyOne()).
		WithAttributeRule("to", exactlyOne()).
		WithAttributeRule("date", atMostOne())
	return side
}

func fakeitTransactionsRightSideSpec() SideSpec {
	side := sideSpec(
		[]ArgKind{ArgKindAttribute, ArgKindText},
		settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList),
		settableOnlyAttribute("category").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList),
		settableOnlyAttribute("year").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList|AttributeValueShapeRange),
		settableOnlyAttribute("month").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList|AttributeValueShapeRange),
	).WithKindRule(ArgKindText, atMostOne())
	for _, name := range []string{"account", "category", "year", "month"} {
		side = side.WithAttributeRule(name, atMostOne())
	}
	return side
}

func fakeitStoresRightSideSpec() SideSpec     { return emptySideSpec() }
func fakeitAccountsRightSideSpec() SideSpec   { return emptySideSpec() }
func fakeitGroupsRightSideSpec() SideSpec     { return emptySideSpec() }
func fakeitTagsRightSideSpec() SideSpec       { return emptySideSpec() }
func fakeitCategoriesRightSideSpec() SideSpec { return emptySideSpec() }

func budgetSideSpec() SideSpec {
	return SideSpec{
		AllowedKinds: allSupportedKinds(),
		Attributes: attributes(
			settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList),
			settableOnlyAttribute("category").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList),
			settableOnlyAttribute("currency").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList),
			settableOnlyAttribute("date").SetShapes(AttributeValueShapeSingle|AttributeValueShapeRange),
			settableOnlyAttribute("desc").SetShapes(AttributeValueShapeSingle),
			settableOnlyAttribute("group").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList),
			settableOnlyAttribute("store").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList),
		),
		KindRules:      map[ArgKind]CountRule{},
		AttributeRules: map[string]CountRule{},
		AtLeastOneOf:   []PresenceRule{},
	}
}
