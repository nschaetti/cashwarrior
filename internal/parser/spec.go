package parser

import (
	"fmt"
	"sort"
)

// region AttributeSpecs

// AttributeValueShape are the shapes of an attribute value (single, list, range, operator)
type AttributeValueShape uint8

const (
	AttributeValueShapeSingle AttributeValueShape = 1 << iota // Attribute value can be a single value
	AttributeValueShapeList                                   // Attribute value can be a list of values
	AttributeValueShapeRange                                  // Attribute value can be a range of values
)

// IsSingle returns true if the attribute value is a single value
func (s AttributeValueShape) IsSingle() bool {
	return s&AttributeValueShapeSingle != 0
}

// IsList returns true if the attribute value is a list of values
func (s AttributeValueShape) IsList() bool {
	return s&AttributeValueShapeList != 0
}

// IsRange returns true if the attribute value is a range of values
func (s AttributeValueShape) IsRange() bool {
	return s&AttributeValueShapeRange != 0
}

func (s AttributeValueShape) String() string {
	switch s {
	case AttributeValueShapeSingle:
		return "single"
	case AttributeValueShapeList:
		return "list"
	case AttributeValueShapeRange:
		return "range"
	default:
		return "unknown"
	}
}

// AttributeValueType are the value types of an attribute token (string, int, float, etc)
type AttributeValueType uint8

const (
	AttributeValueTypeString AttributeValueType = iota
	AttributeValueTypeInteger
	AttributeValueTypeFloat
	AttributeValueTypeDate
	AttributeValueTypeBool
	AttributeValueTypeFile
)

// String returns a string representation of the attribute value type.
func (t AttributeValueType) String() string {
	switch t {
	case AttributeValueTypeString:
		return "string"
	case AttributeValueTypeInteger:
		return "integer"
	case AttributeValueTypeFloat:
		return "float"
	case AttributeValueTypeDate:
		return "date"
	default:
		panic("unhandled default case")
	}
}

// AttributeSpec is the specifications of an attribute token
type AttributeSpec struct {
	Name          string              // Name of the attribute
	AllowedShapes AttributeValueShape // Allowed Shapes of the attributes (single, list, range, operator)
	Shapes        AttributeValueShape // Shapes of the attribute (single, list, range, operator)
	Type          AttributeValueType  // Type of the value (String, int, float, etc)
	AllowSettable bool                // Settable allowed
	AllowClear    bool                // AllowClear allowed
	Settable      bool                // Can be set (modify) or only for filters
	Clearable     bool                // Can be cleared (modify)
}

// String returns a string representation of the attribute spec.
func (s AttributeSpec) String() string {
	return fmt.Sprintf(
		"<AttributeSpec name=%s, allowed_shapes=%s, shapes=%s, type=%s, settable=%t, clearable=%t>",
		s.Name,
		s.Shapes,
		s.Type,
		s.Settable,
		s.Clearable,
	)
}

func (s AttributeSpec) SetShapes(shapes AttributeValueShape) AttributeSpec {
	s.Shapes = s.AllowedShapes & shapes
	return s
}

// AllowsSet AllowSet returns true if the attribute can be set (modify) or only for filters
func (s AttributeSpec) AllowsSet() bool {
	if !s.Settable && !s.Clearable {
		return true
	}
	return s.Settable
}

// AllowsClear returns true if the attribute can be cleared (modify)
func (s AttributeSpec) AllowsClear() bool {
	return s.Clearable
}

// getAttributeSpec returns the AttributeSpec for the given name.
func getAttributeSpec(name string) (AttributeSpec, error) {
	attrSpec, ok := AttributeSpecs[name]
	if !ok {
		return AttributeSpec{}, fmt.Errorf("unknown attribute %s", name)
	}
	return attrSpec, nil
}

func buildCompleteAttributeSpec(
	name string,
	shapes AttributeValueShape,
	settable bool,
	clearable bool,
) AttributeSpec {
	attrSpec, err := getAttributeSpec(name)
	if err != nil {
		panic(err)
	}
	if settable && !attrSpec.AllowSettable {
		panic(fmt.Errorf("attribute %s is not settable", name))
	}
	if clearable && !attrSpec.AllowClear {
		panic(fmt.Errorf("attribute %s is not clearable", name))
	}
	attrSpec.Settable = settable
	attrSpec.Clearable = clearable
	attrSpec.Shapes = shapes
	return attrSpec
}

func buildAttributeSpec(name string) AttributeSpec {
	attrSpec, err := getAttributeSpec(name)
	if err != nil {
		panic(err)
	}
	return attrSpec
}

// settableOnlyAttribute creates an AttributeSpec with the given name and shapes and sets the attribute as settable.
func settableOnlyAttribute(name string) AttributeSpec {
	attrSpec, err := getAttributeSpec(name)
	if err != nil {
		panic(err)
	}
	if !attrSpec.AllowSettable {
		panic(fmt.Errorf("attribute %s is not settable", name))
	}
	attrSpec.Settable = true
	return attrSpec
}

// settableOnlyAttribute creates an AttributeSpec with the given name and shapes and sets the attribute as clearable.
func setOrClearAttribute(name string) AttributeSpec {
	attrSpec := settableOnlyAttribute(name)
	if !attrSpec.AllowClear {
		panic(fmt.Errorf("attribute %s is not clearable", name))
	}
	attrSpec.Clearable = true
	return attrSpec
}

// AttributeSpecs is a map of all the attributes that can be used in a transaction.
var AttributeSpecs = map[string]AttributeSpec{
	"amount": {
		Name:   "amount",
		Shapes: AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange,
		Type:   AttributeValueTypeFloat,
	},
	"account": {
		Name:   "account",
		Shapes: AttributeValueShapeSingle | AttributeValueShapeList,
		Type:   AttributeValueTypeString,
	},
	"category": {
		Name:   "category",
		Shapes: AttributeValueShapeSingle | AttributeValueShapeList,
		Type:   AttributeValueTypeString,
	},
	"currency": {
		Name:   "currency",
		Shapes: AttributeValueShapeSingle | AttributeValueShapeList,
		Type:   AttributeValueTypeString,
	},
	"date": {
		Name:   "date",
		Shapes: AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange,
		Type:   AttributeValueTypeDate,
	},
	"desc": {
		Name:   "desc",
		Shapes: AttributeValueShapeSingle,
		Type:   AttributeValueTypeBool,
	},
	"from": {
		Name:   "from",
		Shapes: AttributeValueShapeSingle,
		Type:   AttributeValueTypeString,
	},
	"group": {
		Name:   "group",
		Shapes: AttributeValueShapeSingle | AttributeValueShapeList,
		Type:   AttributeValueTypeString,
	},
	"identifier": {
		Name:   "identifier",
		Shapes: AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange,
		Type:   AttributeValueTypeString,
	},
	"initial-balance": {
		Name:   "initial-balance",
		Shapes: AttributeValueShapeSingle,
		Type:   AttributeValueTypeFloat,
	},
	"output": {
		Name:   "output",
		Shapes: AttributeValueShapeSingle,
		Type:   AttributeValueTypeFile,
	},
	"order": {
		Name:   "order",
		Shapes: AttributeValueShapeSingle | AttributeValueShapeList,
		Type:   AttributeValueTypeString,
	},
	"store": {
		Name:   "store",
		Shapes: AttributeValueShapeSingle | AttributeValueShapeList,
		Type:   AttributeValueTypeString,
	},
	"to": {
		Name:   "to",
		Shapes: AttributeValueShapeSingle,
		Type:   AttributeValueTypeString,
	},
}

var transactionFilterAttributes = []string{"account", "currency", "store", "desc", "date", "group", "identifier"}

// endregion AttributeSpecs

// region SideSpec

// CountRule specifies the number of times a token of a given kind or attribute
// must appear in a sequence.
type CountRule struct {
	Min int // Min number of times the token must appear.
	Max int // Max number of times the token may appear.
}

// countRule creates a CountRule with the given minimum and maximum values.
func countRule(min int, max int) CountRule {
	return CountRule{Min: min, Max: max}
}

// exactlyOne creates a CountRule with eactly one occurrence.
func exactlyOne() CountRule {
	return CountRule{Min: 1, Max: 1}
}

// exactlyZero creates a CountRule with exactly zero occurrences.
func atLeastOne() CountRule {
	return CountRule{Min: 1}
}

// atMostOne creates a CountRule with at most one occurrence.
func atMostOne() CountRule {
	return CountRule{Max: 1}
}

// PresenceRule specifies that at least one of the given kinds or attributes
// must appear in a sequence.
// We can give :
// - a list of token kinds, in which case the rule is satisfied if any of the given kinds appear.
// - a list of attribute names, in which case the rule is satisfied if any of the given attributes appear.
// - a list of token kinds and attribute names, in which case the rule is satisfied if any of the given kinds or attributes appear.
type PresenceRule struct {
	Kinds      []TokenKind // List of kinds that must appear.
	Attributes []string    // List of attributes that must appear.
	Message    string      // Error message to display if the rule is not met.
}

// SideSpec specifies the allowed tokens on either side of a command.
// We can specify :
// - a list of allowed token kinds.
// - a list of allowed attributes.
// - a minimum and maximum number of arguments.
// - a list of CountRules for each kind.
// - a list of CountRules for each attribute.
// - a list of PresenceRules.
type SideSpec struct {
	AllowedKinds   map[TokenKind]bool       // Allowed token kinds.
	AllowAnyAttr   bool                     // Allow any attribute.
	Attributes     map[string]AttributeSpec // List of allowed attributes.
	MinArgs        int                      // Minimum number of arguments.
	MaxArgs        int                      // Maximum number of arguments.
	KindRules      map[TokenKind]CountRule  // Count rules for each kind.
	AttributeRules map[string]CountRule     // Count rules for each attribute.
	AtLeastOneOf   []PresenceRule           // List of PresenceRules.
}

// genericSideSpec specifies a side that accepts any token kind and any attribute.
func genericSideSpec() SideSpec {
	return SideSpec{
		AllowedKinds:   allSupportedKinds(),       // all token kinds are allowed
		AllowAnyAttr:   true,                      // any attribute is allowed
		KindRules:      map[TokenKind]CountRule{}, // no count rules for token kinds
		AttributeRules: map[string]CountRule{},    // no count rules for attributes
		AtLeastOneOf:   []PresenceRule{},          // no PresenceRules
	}
}

// emptySideSpec specifies a side that accepts no tokens.
func emptySideSpec() SideSpec {
	return SideSpec{
		AllowedKinds:   map[TokenKind]bool{},       // no token kinds are allowed
		Attributes:     map[string]AttributeSpec{}, // no attributes are allowed
		KindRules:      map[TokenKind]CountRule{},  // no count rules for token kinds
		AttributeRules: map[string]CountRule{},     // no count rules for attributes
		AtLeastOneOf:   []PresenceRule{},           // no PresenceRules
	}
}

// sideSpec specifies a side that accepts a list of token kinds and a list of attributes.
func sideSpec(kinds []TokenKind, attrs ...AttributeSpec) SideSpec {
	return SideSpec{
		AllowedKinds:   allowKinds(kinds...),      // allowed token kinds are given by the list of kinds
		Attributes:     attributes(attrs...),      // allowed attributes are given by the list of attributes
		KindRules:      map[TokenKind]CountRule{}, // no count rules for token kinds
		AttributeRules: map[string]CountRule{},    // no count rules for attributes
		AtLeastOneOf:   []PresenceRule{},          // no PresenceRules
	}
}

// sideSpecWithAttributes specifies a side that accepts a list of token kinds and a list of attributes.
func sideSpecWithAnyAttributes(kinds ...TokenKind) SideSpec {
	return SideSpec{
		AllowedKinds:   allowKinds(kinds...),      // allowed token kinds are given by the list of kinds
		AllowAnyAttr:   true,                      // any attribute is allowed
		KindRules:      map[TokenKind]CountRule{}, // no count rules for token kinds
		AttributeRules: map[string]CountRule{},    // no count rules for attributes
		AtLeastOneOf:   []PresenceRule{},          // no PresenceRules
	}
}

// WithArgs specifies the minimum and maximum number of arguments in a SideSpec.
func (s SideSpec) WithArgs(min int, max int) SideSpec {
	s.MinArgs = min
	s.MaxArgs = max
	return s
}

func (s SideSpec) WithKindRule(kind TokenKind, rule CountRule) SideSpec {
	if s.KindRules == nil {
		s.KindRules = map[TokenKind]CountRule{}
	}
	s.KindRules[kind] = rule
	return s
}

// WithAttributeRule specifies a CountRule for an attribute in a SideSpec.
func (s SideSpec) WithAttributeRule(name string, rule CountRule) SideSpec {
	if s.AttributeRules == nil {
		s.AttributeRules = map[string]CountRule{}
	}
	s.AttributeRules[name] = rule
	return s
}

// WithAtLeastOneOf specifies a PresenceRule in a SideSpec.
func (s SideSpec) WithAtLeastOneOf(rule PresenceRule) SideSpec {
	s.AtLeastOneOf = append(s.AtLeastOneOf, rule)
	return s
}

// endregion SideSpec

// SubcommandSpec specifies a subcommand of a command.
// We can specify :
// - the left side of the subcommand.
// - the right side of the subcommand.
// - the default subcommand.
type SubcommandSpec struct {
	Name  string
	Left  SideSpec
	Right SideSpec
}

// CommandSpec specifies a command.
// We can specify :
// - the name of the command.
// - the default subcommand.
// - a list of subcommands.
type CommandSpec struct {
	Name              string
	DefaultSubcommand string
	Subcommands       map[string]SubcommandSpec
}

// String returns a string representation of the command spec.
func (c CommandSpec) String() string {
	return fmt.Sprintf(
		"<CommandSpec name=%s, default=%s, subcommands=%v>",
		c.Name,
		c.DefaultSubcommand,
		c.Subcommands,
	)
}

// allowKinds returns a map with TokenKind as key and true as value
func allowKinds(kinds ...TokenKind) map[TokenKind]bool {
	allowed := make(map[TokenKind]bool, len(kinds))
	for _, kind := range kinds {
		allowed[kind] = true
	}
	return allowed
}

// attributes maps attribute names to attribute specs
func attributes(specs ...AttributeSpec) map[string]AttributeSpec {
	attrs := make(map[string]AttributeSpec, len(specs))
	for _, spec := range specs {
		attrs[spec.Name] = spec
	}
	return attrs
}

// subcommands maps subcommand names to subcommand specs
func subcommands(specs ...SubcommandSpec) map[string]SubcommandSpec {
	items := make(map[string]SubcommandSpec, len(specs))
	for _, spec := range specs {
		items[spec.Name] = spec
	}
	return items
}

// allSupportedKinds returns supported all token kinds as supported
func allSupportedKinds() map[TokenKind]bool {
	return allowKinds(
		TokenTag,
		TokenTagNegative,
		TokenAttribute,
		TokenAttributeClear,
		TokenText,
	)
}

// transactionFilterSideSpec specifies the left side as a filter for transactions.
func transactionFilterSideSpec(extraAttrs ...AttributeSpec) SideSpec {
	base := []AttributeSpec{
		buildAttributeSpec("account").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
		buildAttributeSpec("category").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
		buildAttributeSpec("date").SetShapes(AttributeValueShapeSingle | AttributeValueShapeRange | AttributeValueShapeList),
		buildAttributeSpec("desc").SetShapes(AttributeValueShapeSingle),
		buildAttributeSpec("store").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
		buildAttributeSpec("identifier").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange),
		buildAttributeSpec("group").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
	}
	base = append(base, extraAttrs...)
	return sideSpec([]TokenKind{TokenText, TokenAttribute}, base...)
}

// transactionFilterSideSpecWithoutAccount specifies the left side as a filter for transactions.
func transactionFilterSideSpecWithoutAccount(extraAttrs ...AttributeSpec) SideSpec {
	base := []AttributeSpec{
		buildAttributeSpec("category").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
		buildAttributeSpec("date").SetShapes(AttributeValueShapeSingle | AttributeValueShapeRange | AttributeValueShapeList),
		buildAttributeSpec("desc").SetShapes(AttributeValueShapeSingle),
		buildAttributeSpec("store").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
		buildAttributeSpec("identifier").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange),
		buildAttributeSpec("group").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
	}
	return sideSpec([]TokenKind{TokenText, TokenAttribute}, base...)
}

// accountFilterSideSpec specifies the left side as a filter for accounts.
func accountFilterSideSpec(extraAttrs ...AttributeSpec) SideSpec {
	base := []AttributeSpec{
		buildAttributeSpec("account").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
		buildAttributeSpec("currency").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
		buildAttributeSpec("initial-balance").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange),
		buildAttributeSpec("balance").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange),
	}
	base = append(base, extraAttrs...)
	return sideSpec([]TokenKind{TokenText, TokenAttribute}, base...)
}

// groupFilterSideSpec specifies the left side as a filter for groups.
func groupFilterSideSpec(extraAttrs ...AttributeSpec) SideSpec {
	base := []AttributeSpec{
		buildAttributeSpec("group").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
		buildAttributeSpec("size").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange),
	}
	base = append(base, extraAttrs...)
	return sideSpec([]TokenKind{TokenText, TokenAttribute}, base...)
}

// tagFilterSideSpec specifies the left side as a filter for tags.
func tagFilterSideSpec(extraAttrs ...AttributeSpec) SideSpec {
	base := []AttributeSpec{
		buildAttributeSpec("tag").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
		buildAttributeSpec("size").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange),
	}
	base = append(base, extraAttrs...)
	return sideSpec([]TokenKind{TokenText, TokenAttribute}, base...)
}

// placesFilterSideSpec specifies the left side as a filter for places.
func placesFilterSideSpec(extraAttrs ...AttributeSpec) SideSpec {
	base := []AttributeSpec{
		buildAttributeSpec("place").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList),
		buildAttributeSpec("size").SetShapes(AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange),
	}
	base = append(base, extraAttrs...)
	return sideSpec([]TokenKind{TokenText, TokenAttribute}, base...)
}

// addCommandRightSideSpec specifies the right side of the add command.
func addCommandRightSideSpec() SideSpec {
	side := sideSpec(
		[]TokenKind{TokenTag, TokenAttribute, TokenText},
		settableOnlyAttribute("date").SetShapes(AttributeValueShapeSingle),
		settableOnlyAttribute("store").SetShapes(AttributeValueShapeSingle),
		settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle),
		setOrClearAttribute("category").SetShapes(AttributeValueShapeSingle),
		setOrClearAttribute("group").SetShapes(AttributeValueShapeSingle),
	).WithArgs(2, 0).
		WithAttributeRule("amount", exactlyOne()).
		WithAttributeRule("store", exactlyOne())
	for _, name := range []string{"date", "account", "category", "group"} {
		side.WithAttributeRule(name, atMostOne())
	}
	side.WithAtLeastOneOf(PresenceRule{Kinds: []TokenKind{TokenText}, Message: "add requires a description"})
	return side
}

func modifyRightSideSpec() SideSpec {
	side := sideSpec(
		[]TokenKind{TokenAttribute, TokenAttributeClear, TokenTag, TokenTagNegative},
		settableOnlyAttribute("identifier").SetShapes(AttributeValueShapeSingle),
		settableOnlyAttribute("amount").SetShapes(AttributeValueShapeSingle),
		settableOnlyAttribute("desc").SetShapes(AttributeValueShapeSingle),
		settableOnlyAttribute("date").SetShapes(AttributeValueShapeSingle),
		settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle),
		setOrClearAttribute("category").SetShapes(AttributeValueShapeSingle),
		settableOnlyAttribute("store").SetShapes(AttributeValueShapeSingle),
		setOrClearAttribute("group").SetShapes(AttributeValueShapeSingle),
	).WithArgs(1, 0)
	for _, name := range []string{"identifier", "amount", "desc", "date", "time", "datetime", "account", "category", "store", "group"} {
		side.WithAttributeRule(name, atMostOne())
	}
	return side
}

func transferRightSideSpec() SideSpec {
	side := sideSpec(
		[]TokenKind{TokenAttribute, TokenText},
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
		[]TokenKind{TokenAttribute, TokenText},
		settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList),
		settableOnlyAttribute("category").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList),
		settableOnlyAttribute("year").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList|AttributeValueShapeRange),
		settableOnlyAttribute("month").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList|AttributeValueShapeRange),
	).WithKindRule(TokenText, atMostOne())
	for _, name := range []string{"account", "category", "year", "month"} {
		side.WithAttributeRule(name, atMostOne())
	}
	return side
}

func fakeitStoresRightSideSpec() SideSpec {
	return emptySideSpec()
}

func fakeitAccountsRightSideSpec() SideSpec {
	return emptySideSpec()
}

func fakeitGroupsRightSideSpec() SideSpec {
	return emptySideSpec()
}

func fakeitTagsRightSideSpec() SideSpec {
	return emptySideSpec()
}

func fakeitCategoriesRightSideSpec() SideSpec {
	return emptySideSpec()
}

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
		KindRules:      map[TokenKind]CountRule{},
		AttributeRules: map[string]CountRule{},
		AtLeastOneOf:   []PresenceRule{},
	}
}

func defaultCommandSpec(name string) CommandSpec {
	return CommandSpec{
		Name:              name,
		DefaultSubcommand: "default",
		Subcommands: subcommands(SubcommandSpec{
			Name:  "default",
			Left:  genericSideSpec(),
			Right: genericSideSpec(),
		}),
	}
}

func accountsListSubcommandSpec(subcommandName string) SubcommandSpec {
	// cash [balance] [name] [currency] [account] accounts list
	return SubcommandSpec{
		Name: subcommandName,
		// cash accounts list
		Left: sideSpec(
			[]TokenKind{TokenAttribute, TokenTag, TokenTagNegative},
			settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList),
			settableOnlyAttribute("name").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList),
			settableOnlyAttribute("currency").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList),
			settableOnlyAttribute("balance").SetShapes(AttributeValueShapeSingle|AttributeValueShapeList|AttributeValueShapeRange),
		),
		Right: emptySideSpec().WithArgs(0, 0), // No arguments on the right
	}
}

type SubcommandAlias struct {
	SubcommandName  string
	SubcommandAlias string
}

// createSubcommandAlias creates a subcommand alias for a command.
// This is useful for creating aliases for subcommands.
func createSubcommandAlias(
	command CommandSpec,
	aliases []SubcommandAlias,
) CommandSpec {
	for _, alias := range aliases {
		subCommand, ok := command.Subcommands[alias.SubcommandName]
		if !ok {
			panic(fmt.Sprintf("subcommand %s not found in command %s", alias.SubcommandName, command.Name))
		}
		command.Subcommands[alias.SubcommandAlias] = subCommand
	}
	return command
}

// region CommandSpec

// accountsCommandSpec specifies the command spec for the "accounts" command.
//
// Subcommand:
// - list
// - add
// - modify
// - initial-balance
// - rename
// - delete
var accountsCommandSpec = createSubcommandAlias(
	CommandSpec{
		Name:              "accounts",
		DefaultSubcommand: "list",
		Subcommands: subcommands(
			// > Subcommand: list
			// $ cash [account-filter] accounts list
			// [balance] is a filter on the balance attribute
			// [name] is a filter on the account name field
			// [currency] is a filter on the currency attribute
			// [account] is a filter on the account name field
			accountsListSubcommandSpec("list"),
			// > Subcommand: balance
			// cash [transaction filter] accounts balance <account name>
			// category, date, desc, store, identifier, group are all optional filter attributes.
			SubcommandSpec{
				Name: "balance",
				Left: transactionFilterSideSpecWithoutAccount(),
				Right: sideSpec(
					[]TokenKind{TokenText},
					settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("name").SetShapes(AttributeValueShapeSingle),
				).
					WithAtLeastOneOf(
						PresenceRule{
							Kinds:      []TokenKind{TokenText},
							Attributes: []string{"account", "name"},
							Message:    "accounts balance requires an account name",
						},
					).
					WithArgs(1, 1),
			},
			// > Subcommand: add
			// $ cash accounts add <account name> [currency] [initial balance]
			// <account name> is the name of the account to add
			// [currency] is the currency of the account to add (default is default currency)
			// [initial balance] is the initial balance of the account to add (default is 0)
			SubcommandSpec{
				Name: "add",
				Left: emptySideSpec(),
				Right: sideSpec(
					[]TokenKind{TokenText, TokenAttribute},
					settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("currency").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("initial-balance").SetShapes(AttributeValueShapeSingle),
				).
					WithKindRule(TokenText, atMostOne()).
					WithArgs(1, 3).
					WithAttributeRule("currency", atMostOne()).
					WithAttributeRule("initial_balance", atMostOne()).
					WithAtLeastOneOf(
						PresenceRule{
							Kinds:      []TokenKind{TokenText},
							Attributes: []string{"account"},
							Message:    "accounts add requires an account name",
						},
					),
			},
			// > Subcommand: modify
			// $ cash <account> accounts modify [currency] [initial balance] [name]
			// <account> is the account to modify
			// [currency] is the currency of the account to modify (default is default currency)
			// [initial balance] is the initial balance of the account to modify (default is 0)
			// [name] is the new name of the account
			SubcommandSpec{
				Name: "modify",
				// cash accounts modify <account name> [currency] [initial balance]
				Left: sideSpec(
					[]TokenKind{TokenAttribute},
					settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle),
				),
				Right: sideSpec(
					[]TokenKind{TokenText, TokenAttribute},
					settableOnlyAttribute("name").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("currency").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("initial_balance").SetShapes(AttributeValueShapeSingle),
				).
					WithArgs(2, 3).
					WithKindRule(TokenText, atMostOne()).
					WithAttributeRule("name", atMostOne()).
					WithAttributeRule("currency", atMostOne()).
					WithAttributeRule("initial_balance", atMostOne()).
					WithAtLeastOneOf(
						PresenceRule{
							Kinds:      []TokenKind{TokenText},
							Attributes: []string{"name", "currency", "initial_balance"},
							Message:    "accounts modify requires at least one modification",
						},
					),
			},
			// > Subcommand: initial-balance
			// $ cash <account> accounts initial-balance [initial balance]
			// initial-balance subcommand (set initial balance of an account)
			// cash accounts initial-balance <account name> [initial-balance|amount]
			SubcommandSpec{
				Name: "initial-balance",
				Left: sideSpec(
					[]TokenKind{TokenAttribute},
					settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle),
				),
				Right: sideSpec(
					[]TokenKind{TokenAttribute},
					settableOnlyAttribute("amount").SetShapes(AttributeValueShapeSingle),
				).
					WithArgs(1, 1).
					WithAttributeRule("amount", exactlyOne()),
			},
			// > Subcommand: rename
			// $ cash <account> accounts rename <new account name>
			SubcommandSpec{
				Name: "rename",
				Left: sideSpec(
					[]TokenKind{TokenAttribute},
					settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle),
				),
				Right: sideSpec(
					[]TokenKind{TokenText},
				).
					WithArgs(1, 1).
					WithKindRule(TokenText, countRule(1, 1)),
			},
			// > Subcommand: delete
			// $ cash <account> accounts delete
			// delete subcommand (delete an account)
			SubcommandSpec{
				Name: "delete",
				Left: sideSpec(
					[]TokenKind{TokenAttribute},
					settableOnlyAttribute("account").SetShapes(AttributeValueShapeSingle),
				),
				Right: emptySideSpec().WithArgs(0, 0),
			},
		),
	},
	[]SubcommandAlias{
		{"list", "ls"},
		{"add", "create"},
		{"modify", "update"},
		{"initial-balance", "set-initial-balance"},
		{"rename", "mv"},
		{"delete", "rm"},
	},
)

// addCommandSpec specifies the command spec for the "add" command.
// cash add <amount> <store> <desc...> [date] [account] [category] [group]
// amount is required
// store is required
// desc is required
// date is optional (current date)
// account is optional (default account)
// category is optional (no category)
// group is optional (no group)
//
// Examples:
// cash add amount:-100 store:SuperM category:groceries date:2020-01-01 Bought some groceries at SuperM
var addCommandSpec = CommandSpec{
	Name:              "add",
	DefaultSubcommand: "default",
	Subcommands: subcommands(
		// > Subcommand: default (add)
		// cash add <amount> <store> <desc...> [date] [account] [category] [group]
		SubcommandSpec{
			Name:  "default",
			Left:  emptySideSpec(),
			Right: addCommandRightSideSpec(),
		},
	),
}

// backupCommandSpec specifies the command spec for the "backup" command.
// cash backup [output]
// output is optional (default is stdout)
//
// Examples:
// cash backup output:backup.db
// cash backup
var backupCommandSpec = CommandSpec{
	Name:              "backup",
	DefaultSubcommand: "now",
	Subcommands: subcommands(
		// > Subcommand: now (backup)
		// cash backup now [output]
		SubcommandSpec{
			Name: "now",
			Left: emptySideSpec(),
			Right: sideSpec(
				[]TokenKind{TokenAttribute},
				settableOnlyAttribute("output").SetShapes(AttributeValueShapeSingle),
			).
				WithArgs(0, 1).
				WithAttributeRule("output", atMostOne()),
		},
		// > Subcommand: to (backup)
		// cash backup to <output>
		SubcommandSpec{
			Name: "to",
			Left: emptySideSpec(),
			Right: sideSpec(
				[]TokenKind{TokenAttribute},
				settableOnlyAttribute("output").SetShapes(AttributeValueShapeSingle),
			).
				WithArgs(1, 1).
				WithAttributeRule("output", exactlyOne()),
		},
	),
}

// budgetCommandSpec specifies the command spec for the "budget" command.
// cash budget list
// cash budget add <account> <category> <amount> <date> <desc> [group] [store]
// account is required
// category is required
// amount is required
// date is required
// desc is required
// group is optional (no group)
// store is optional (no store)
var budgetCommandSpec = createSubcommandAlias(
	CommandSpec{
		Name:              "budget",
		DefaultSubcommand: "list",
		Subcommands: subcommands(
			// cash budget list
			SubcommandSpec{
				Name:  "list",
				Left:  budgetSideSpec(),
				Right: budgetSideSpec(),
			},
			// cash budget add <account> <category> <amount> <date> <desc> [group] [store]
			SubcommandSpec{
				Name:  "add",
				Left:  budgetSideSpec(),
				Right: budgetSideSpec(),
			},
		),
	},
	[]SubcommandAlias{
		{"list", "ls"},
	},
)

// categoriesListSubcommandSpec creates a subcommand for categories listening
func categoriesListSubcommandSpec(commandName string) SubcommandSpec {
	return SubcommandSpec{
		Name:  commandName,
		Left:  emptySideSpec(),
		Right: emptySideSpec().WithArgs(0, 0),
	}
}

// categoriesCommandSpec specifies the command spec for the "categories" command.
// cash categories list
// cash categories add <category name> [parent]
// cash categories modify <category name> [parent]
// cash categories delete <category name>
//
// Examples:
// cash categories add groceries
// cash categories modify groceries parent:groceries
// cash categories delete groceries
var categoriesCommandSpec = createSubcommandAlias(
	CommandSpec{
		Name:              "categories",
		DefaultSubcommand: "list",
		Subcommands: subcommands(
			// cash categories list
			categoriesListSubcommandSpec("list"),
			// cash categories add <category name> [parent]
			SubcommandSpec{
				Name: "add",
				Left: emptySideSpec(),
				Right: sideSpec(
					[]TokenKind{TokenText, TokenAttribute},
					settableOnlyAttribute("category").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("parent").SetShapes(AttributeValueShapeSingle),
				).
					WithArgs(1, 2).
					WithAttributeRule("parent", atMostOne()).
					WithAtLeastOneOf(PresenceRule{Kinds: []TokenKind{TokenText}, Attributes: []string{"category"}, Message: "categories add requires a category name"}),
			},
			// cash categories modify <category name> [parent]
			SubcommandSpec{
				Name: "modify",
				Left: emptySideSpec(),
				Right: sideSpec(
					[]TokenKind{TokenText, TokenAttribute},
					settableOnlyAttribute("category").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("parent").SetShapes(AttributeValueShapeSingle),
				).
					WithArgs(2, 3).
					WithKindRule(TokenText, exactlyOne()).
					WithAttributeRule("category", atMostOne()).
					WithAttributeRule("parent", atMostOne()).
					WithAtLeastOneOf(
						PresenceRule{
							Attributes: []string{"category", "parent"},
							Message:    "categories modify requires at least one modification",
						},
					),
			},
			// cash categories delete <category name>
			SubcommandSpec{
				Name: "delete",
				Left: emptySideSpec(),
				Right: sideSpec(
					[]TokenKind{TokenText, TokenAttribute},
					settableOnlyAttribute("category").SetShapes(AttributeValueShapeSingle),
				).
					WithArgs(1, 1).
					WithAttributeRule("category", atMostOne()).
					WithAtLeastOneOf(
						PresenceRule{
							Kinds:      []TokenKind{TokenText},
							Attributes: []string{"category"},
							Message:    "categories delete requires a category name",
						},
					),
			},
		),
	},
	[]SubcommandAlias{
		{"list", "ls"},
		{"modify", "update"},
		{"delete", "rm"},
	},
)

// configCommandSpec specifies the command spec for the "config" command.
// cash config default <attribute> <value>
// cash config default <attribute>
//
// Examples:
// cash config default currency:USD
// cash config default currency
var configCommandSpec = CommandSpec{
	Name:              "config",
	DefaultSubcommand: "print",
	Subcommands: subcommands(
		// > Subcommand: print (config)
		// $ cash config print
		SubcommandSpec{
			Name:  "print",
			Left:  emptySideSpec(),
			Right: emptySideSpec().WithArgs(0, 0),
		},
		// > Subcommand: get
		// $ cash config get <key>
		SubcommandSpec{
			Name: "get",
			Left: emptySideSpec(),
			Right: sideSpec(
				[]TokenKind{TokenText},
			).
				WithArgs(1, 1).
				WithKindRule(TokenText, countRule(1, 1)),
		},
		// > Subcommand: set
		// $ cash config set <key> <value>
		SubcommandSpec{
			Name: "set",
			Left: emptySideSpec(),
			Right: sideSpec(
				[]TokenKind{TokenText},
			).
				WithArgs(2, 2).
				WithKindRule(TokenText, CountRule{Min: 2, Max: 2}),
		},
	),
}

// deleteCommandSpec specifies the command spec for the "delete" command.
// $ cash <filter> delete [transaction|list]
// id is required
// identifier is optional
// T is optional
//
// Examples:
// cash delete identifier:2026-05-12
// cash delete T:2026-05-12
var deleteCommandSpec = CommandSpec{
	Name:              "delete",
	DefaultSubcommand: "transaction",
	Subcommands: subcommands(
		// > Subcommand: transaction (delete)
		// Delete a transaction.
		// $ cash <filter> delete
		SubcommandSpec{
			Name:  "transaction",
			Left:  transactionFilterSideSpec(),
			Right: emptySideSpec().WithArgs(0, 0),
		},
		// > Subcommand: list (delete)
		// List deleted transactions.
		// $ cash <filter> delete list
		SubcommandSpec{
			Name:  "list",
			Left:  transactionFilterSideSpec(),
			Right: emptySideSpec().WithArgs(0, 0),
		},
	),
}

// fakeitCommandSpec specifies the command spec for the "fake-it" command.
// cash fake-it <account> <category> <type> <year> <month>
// account is optional
// category is optional
// type is optional
// year is optional
// month is optional
//
// Examples:
// cash fake-it account:groceries category:groceries type:groceries year:2020 month:1
var fakeitCommandSpec = CommandSpec{
	Name:              "fake-it",
	DefaultSubcommand: "transactions",
	Subcommands: subcommands(
		// > Subcommand: generate (fake-it)
		// Generate fake transactions.
		SubcommandSpec{
			Name:  "transactions",
			Left:  emptySideSpec(),
			Right: fakeitTransactionsRightSideSpec(),
		},
		// > Subcommand: categories (fake-it)
		// Generate fake categories.
		SubcommandSpec{
			Name:  "categories",
			Left:  emptySideSpec(),
			Right: fakeitCategoriesRightSideSpec(),
		},
		// > Subcommand: accounts (fake-it)
		// Generate fake accounts.
		SubcommandSpec{
			Name:  "accounts",
			Left:  emptySideSpec(),
			Right: fakeitAccountsRightSideSpec(),
		},
		// > Subcommand: stores (fake-it)
		// Generate fake stores
		SubcommandSpec{
			Name:  "stores",
			Left:  emptySideSpec(),
			Right: fakeitStoresRightSideSpec(),
		},
		// > Subcommand: groups (fake-it)
		// Generate fake groups.
		SubcommandSpec{
			Name:  "groups",
			Left:  emptySideSpec(),
			Right: fakeitGroupsRightSideSpec(),
		},
		// > Subcommand: tags (fake-it)
		// Generate fake tags.
		SubcommandSpec{
			Name:  "tags",
			Left:  emptySideSpec(),
			Right: fakeitTagsRightSideSpec(),
		},
	),
}

// groupsCommandSpec specifies the command spec for the "groups" command.
// cash groups list
//
// Examples:
// cash groups list
var groupsCommandSpec = createSubcommandAlias(
	CommandSpec{
		Name:              "groups",
		DefaultSubcommand: "list",
		Subcommands: subcommands(
			// cash groups list
			SubcommandSpec{
				Name: "list",
				Left: sideSpec(
					[]TokenKind{TokenAttribute},
					settableOnlyAttribute("order").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("desc").SetShapes(AttributeValueShapeSingle),
				),
				Right: emptySideSpec().WithArgs(0, 0),
			},
			// cash groups add
			SubcommandSpec{
				Name: "add",
				Left: emptySideSpec(),
				Right: sideSpec(
					[]TokenKind{TokenText, TokenAttribute},
				).WithArgs(2, 0).
					WithKindRule(TokenText, atMostOne()).
					WithAttributeRule("group", atMostOne()).
					WithAtLeastOneOf(
						PresenceRule{
							Kinds:      []TokenKind{TokenText},
							Attributes: []string{"group"},
							Message:    "groups add requires a group name",
						},
					).
					WithAtLeastOneOf(
						PresenceRule{
							Kinds:      []TokenKind{TokenAttribute},
							Attributes: []string{"id", "identifier", "T"},
							Message:    "groups add requires at least one transaction id",
						},
					),
			},
			// cash groups modify
			SubcommandSpec{
				Name: "modify",
				Left: emptySideSpec(),
				Right: sideSpec(
					[]TokenKind{TokenText, TokenAttribute},
				).WithArgs(2, 2).
					WithKindRule(TokenAttribute, atMostOne()).
					WithKindRule(TokenText, atMostOne()),
			},
			// cash groups delete
			SubcommandSpec{
				Name: "delete",
				Left: emptySideSpec(),
				Right: sideSpec(
					[]TokenKind{TokenText},
				).
					WithArgs(1, 1).
					WithKindRule(TokenText, exactlyOne()),
			},
			// cash groups remove <transaction> <group>
			SubcommandSpec{
				Name: "remove",
				Left: emptySideSpec(),
				Right: sideSpec(
					[]TokenKind{TokenAttribute},
				).
					WithArgs(2, 2).
					WithAttributeRule("id", atMostOne()).
					WithAttributeRule("identifier", atMostOne()).
					WithAttributeRule("T", atMostOne()).
					WithAttributeRule("group", exactlyOne()).
					WithAtLeastOneOf(
						PresenceRule{
							Kinds:      []TokenKind{TokenAttribute},
							Attributes: []string{"id", "identifier", "T", "group"},
							Message:    "groups remove requires at least one transaction id",
						},
					),
			},
		),
	},
	[]SubcommandAlias{
		{SubcommandName: "list", SubcommandAlias: "ls"},
		{SubcommandName: "modify", SubcommandAlias: "rename"},
		{SubcommandName: "modify", SubcommandAlias: "rn"},
		{SubcommandName: "delete", SubcommandAlias: "rm"},
	},
)

// importCommandSpec specifies the command spec for the "import" command.
var importCommandSpec = CommandSpec{
	Name:              "import",
	DefaultSubcommand: "csv",
	Subcommands: subcommands(
		// > Subcommand: csv (import)
		// $ cash import csv <file>
		SubcommandSpec{
			Name:  "csv",
			Left:  emptySideSpec(),
			Right: sideSpec([]TokenKind{TokenText}).WithArgs(1, 1),
		},
	),
}

// listCommandSpec specifies the command spec for the "list" command.
// cash list <account> <category> <group> <store>
// account is optional
// category is optional
// group is optional
// store is optional
//
// Examples:
// cash list account:groceries
// cash list account:groceries category:groceries
var listCommandSpec = createSubcommandAlias(
	CommandSpec{
		Name:              "list",
		DefaultSubcommand: "transactions",
		Subcommands: subcommands(
			// > Subcommand: transactions (list)
			// $ cash list transactions
			SubcommandSpec{
				Name: "transactions",
				Left: transactionFilterSideSpec(
					settableOnlyAttribute("order").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("desc").SetShapes(AttributeValueShapeSingle),
				),
				Right: genericSideSpec(),
			},
			// > Subcommand: accounts (list)
			// $ cash list accounts
			SubcommandSpec{
				Name: "accounts",
				Left: accountFilterSideSpec(
					settableOnlyAttribute("order").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("desc").SetShapes(AttributeValueShapeSingle),
				),
			},
			// > Subcommand: stores (list)
			// $ cash list stores
			SubcommandSpec{
				Name: "groups",
				Left: groupFilterSideSpec(
					settableOnlyAttribute("order").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("desc").SetShapes(AttributeValueShapeSingle),
				),
			},
			// > Subcommand: categories (list)
			// $ cash list categories
			SubcommandSpec{
				Name: "tags",
				Left: tagFilterSideSpec(
					settableOnlyAttribute("order").SetShapes(AttributeValueShapeSingle),
					settableOnlyAttribute("desc").SetShapes(AttributeValueShapeSingle),
				),
			},
		),
	},
	[]SubcommandAlias{
		{"transactions", "t"},
		{"accounts", "a"},
		{"groups", "g"},
		{"tags", "ta"},
	},
)

// modifyCommandSpec specifies the command spec for the "modify" command.
// cash modify <account> <category> <group> <store> <attribute> <value>
// account is optional
// category is optional
// group is optional
// store is optional
// attribute is required
// value is required
//
// Examples:
var modifyCommandSpec = CommandSpec{
	Name:              "modify",
	DefaultSubcommand: "default",
	Subcommands: subcommands(
		// > Subcommand: default (modify)
		// Modify a transaction.
		// $ cash modify <transaction attributes>
		SubcommandSpec{
			Name:  "default",
			Left:  transactionFilterSideSpec(),
			Right: modifyRightSideSpec(),
		},
	),
}

// storesCommandSpec specifies the command spec for the "stores" command.
var storesCommandSpec = createSubcommandAlias(
	CommandSpec{
		Name:              "stores",
		DefaultSubcommand: "list",
		Subcommands: subcommands(
			// > Subcommand: list (stores)
			// $ cash stores list
			SubcommandSpec{
				Name:  "list",
				Left:  emptySideSpec(),
				Right: emptySideSpec().WithArgs(0, 0),
			},
			// > Subcommand: add
			// $ cash stores add <store name>
			SubcommandSpec{
				Name: "add",
				Left: emptySideSpec(),
				Right: sideSpec(
					[]TokenKind{TokenText},
				).WithArgs(1, 1),
			},
			// > Subcommand: rename
			// $ cash stores modify <store name> <new name>
			SubcommandSpec{
				Name: "rename",
				Left: emptySideSpec(),
				Right: sideSpec([]TokenKind{TokenText}).
					WithArgs(2, 2).
					WithKindRule(TokenText, countRule(2, 2)),
			},
			// > Subcommand: delete
			// $ cash stores delete <store name>
			SubcommandSpec{
				Name: "delete",
				Left: emptySideSpec(),
				Right: sideSpec([]TokenKind{TokenText}).
					WithArgs(1, 1).
					WithKindRule(TokenText, exactlyOne()),
			},
		),
	},
	[]SubcommandAlias{
		{"list", "ls"},
		{"rename", "rn"},
		{"delete", "rm"},
	},
)

// purgeCommandSpec specifies the command spec for the "purge" command.
// cash purge <account> <category> <group> <store>
// account is optional
// category is optional
// group is optional
// store is optional
//
// Examples:
// cash purge account:groceries
// cash purge account:groceries category:groceries
var purgeCommandSpec = CommandSpec{
	Name:              "purge",
	DefaultSubcommand: "default",
	Subcommands: subcommands(
		// cash purge list
		SubcommandSpec{
			Name:  "default",
			Left:  transactionFilterSideSpec(),
			Right: emptySideSpec(),
		},
	),
}

// restoreCommandSpec
var restoreCommandSpec = CommandSpec{
	Name:              "restore",
	DefaultSubcommand: "default",
	Subcommands: subcommands(
		SubcommandSpec{
			Name: "default",
			Left: emptySideSpec(),
			Right: sideSpec(
				[]TokenKind{},
			).
				WithArgs(1, 1).
				WithAttributeRule("id", exactlyOne()),
		},
		SubcommandSpec{
			Name:  "list",
			Left:  emptySideSpec(),
			Right: emptySideSpec().WithArgs(0, 0),
		},
	),
}

// showCommandSpec specifies the command spec for the "show" command.
// cash show <account> <category> <group> <store> <attribute>
//
// Subcommands :
//
// - transaction: show transaction informations
// - accounts: show accounts
// - tags: show tags
// - categories: show categories
var showCommandSpec = CommandSpec{
	Name:              "show",
	DefaultSubcommand: "transaction",
	Subcommands: subcommands(
		// cash show infos <id|T|identifier>
		SubcommandSpec{
			Name: "transaction",
			Left: emptySideSpec(),
			Right: sideSpec(
				[]TokenKind{},
			).
				WithArgs(1, 1).
				WithAttributeRule("id", atMostOne()).
				WithAttributeRule("T", atMostOne()).
				WithAttributeRule("identifier", atMostOne()).
				WithAtLeastOneOf(
					PresenceRule{
						Kinds:      []TokenKind{TokenAttribute},
						Attributes: []string{"id", "T", "identifier"},
						Message:    "show requires an id",
					},
				),
		},
		// cash show categories
		categoriesListSubcommandSpec("categories"),
		// cash show accounts
		accountsListSubcommandSpec("accounts"),
	),
}

// summaryCommandSpec
var summaryCommandSpec = CommandSpec{
	Name:              "summary",
	DefaultSubcommand: "",
	Subcommands: subcommands(
		SubcommandSpec{
			Name:  "days",
			Left:  transactionFilterSideSpec(),
			Right: transactionFilterSideSpec(),
		},
	),
}

// tagsCommandSpec specifies the command spec for the "tags" command.
//
// Subcommands:
//
//   - list
//   - add
//   - modify
//   - delete
var tagsCommandSpec = CommandSpec{
	Name:              "tags",
	DefaultSubcommand: "list",
	Subcommands: subcommands(
		// cash tags list
		SubcommandSpec{
			Name:  "list",
			Left:  emptySideSpec(),
			Right: emptySideSpec().WithArgs(0, 0),
		},
		// add
		SubcommandSpec{
			Name: "add",
			Left: emptySideSpec(),
			Right: sideSpec(
				[]TokenKind{TokenText, TokenAttribute},
				settableOnlyAttribute("tag").SetShapes(AttributeValueShapeSingle),
			).
				WithArgs(1, 1).
				WithAtLeastOneOf(
					PresenceRule{
						Kinds:      []TokenKind{TokenText},
						Attributes: []string{"tag"},
						Message:    "tags add requires a tag name",
					},
				),
		},
		// cash tags modify <tag name> [desc]
		SubcommandSpec{
			Name: "modify",
			Left: emptySideSpec(),
			Right: sideSpec(
				[]TokenKind{TokenText, TokenAttribute},
				settableOnlyAttribute("tag").SetShapes(AttributeValueShapeSingle),
			).
				WithKindRule(TokenText, exactlyOne()).
				WithArgs(2, 2).
				WithAttributeRule("tag", exactlyOne()).
				WithAtLeastOneOf(
					PresenceRule{
						Attributes: []string{"tag"},
						Message:    "tags modify requires a new tag name",
					},
				),
		},
		// cash tags delete <tag name>
		SubcommandSpec{
			Name: "delete",
			Left: emptySideSpec(),
			Right: sideSpec(
				[]TokenKind{TokenText, TokenAttribute},
				settableOnlyAttribute("tag").SetShapes(AttributeValueShapeSingle),
			).WithArgs(1, 1).
				WithAtLeastOneOf(
					PresenceRule{
						Kinds:      []TokenKind{TokenText},
						Attributes: []string{"tag"},
						Message:    "tags delete requires a tag name"},
				),
		},
	),
}

// themeCommandSpec specifies specs for the command "theme" and its subcommand.
// cash theme [theme name] : list theme if empty, set theme otherwise
//
// Subcommands:
//
//   - default: list all themes or set
//   - list: list all themes
//   - set: set theme to the specified theme name
var themeCommandSpec = CommandSpec{
	Name:              "theme",
	DefaultSubcommand: "default",
	Subcommands: subcommands(
		// default
		// cash theme default <theme name>
		SubcommandSpec{
			Name: "default",
			Left: emptySideSpec(),
			Right: sideSpec(
				[]TokenKind{TokenText},
			).WithArgs(0, 1),
		},
		// cash theme list
		SubcommandSpec{
			Name:  "list",
			Left:  emptySideSpec(),
			Right: emptySideSpec().WithArgs(0, 0),
		},
		// cash theme set <theme name>
		SubcommandSpec{
			Name: "set",
			Left: emptySideSpec(),
			Right: sideSpec(
				[]TokenKind{TokenText},
			).WithArgs(1, 1),
		},
	),
}

// transferCommandSpec specifies the command spec for the "transfer" command.
// cash transfer [add|delete] <account> <amount> <desc> [group] [store] [date]
//
// Subcommands:
//
//   - add
//   - delete
//   - list
var transferCommandSpec = CommandSpec{
	Name:              "transfer",
	DefaultSubcommand: "add",
	Subcommands: subcommands(
		// cash transfer add <account> <amount> <desc> [group] [store] [date]
		SubcommandSpec{
			Name:  "add",
			Left:  emptySideSpec(),
			Right: transferRightSideSpec(),
		},
		// cash transfer delete <id>
		// similar to cash delete
		SubcommandSpec{
			Name: "delete",
			Left: emptySideSpec(),
			Right: sideSpec([]TokenKind{TokenAttribute}).
				WithArgs(1, 1).
				WithKindRule(TokenAttribute, exactlyOne()).
				WithAttributeRule("id", atMostOne()).
				WithAttributeRule("identifier", atMostOne()).
				WithAttributeRule("T", atMostOne()).
				WithAtLeastOneOf(
					PresenceRule{
						Kinds:      []TokenKind{TokenAttribute},
						Attributes: []string{"id"},
						Message:    "transfer delete requires an id",
					},
				),
		},
		// cash transfer list
		SubcommandSpec{
			Name:  "list",
			Left:  emptySideSpec(),
			Right: emptySideSpec().WithArgs(0, 0),
		},
	),
}

// CommandSpecs is a map of command names to their specifications.
// accounts: list, balance, add, modify, initial-balance, rename, delete
// add: add a transaction
// backup: backup the database
// budget: list, add, check budgets
// categories: list, add, modify, delete categories
// config: set or get config values
// delete: delete transaction(s)
// fakeit: create fake data
// import: import data from a file
// group: default
// groups: list
var CommandSpecs = map[string]CommandSpec{
	"accounts":    accountsCommandSpec,
	"add":         addCommandSpec,
	"backup":      backupCommandSpec,
	"balance":     defaultCommandSpec("balance"),
	"budget":      budgetCommandSpec,
	"by":          defaultCommandSpec("by"),
	"categories":  categoriesCommandSpec,
	"config":      configCommandSpec,
	"delete":      deleteCommandSpec,
	"export":      defaultCommandSpec("export"),
	"fakeit":      fakeitCommandSpec,
	"groups":      groupsCommandSpec,
	"import":      importCommandSpec,
	"list":        listCommandSpec,
	"modify":      modifyCommandSpec,
	"stores":      storesCommandSpec,
	"purge":       purgeCommandSpec,
	"report":      defaultCommandSpec("report"),
	"restore":     restoreCommandSpec,
	"set-balance": defaultCommandSpec("set-balance"),
	"show":        showCommandSpec,
	"stats":       defaultCommandSpec("stats"),
	"sum":         defaultCommandSpec("sum"),
	"summary":     summaryCommandSpec,
	"tags":        tagsCommandSpec,
	"theme":       themeCommandSpec,
	"transfer":    transferCommandSpec,
	"undo":        defaultCommandSpec("undo"),
}

// endregion CommandSpec

func GetCommandSpec(name string) (CommandSpec, bool) {
	spec, ok := CommandSpecs[name]
	return spec, ok
}

func IsKnownCommand(command string) bool {
	_, ok := CommandSpecs[command]
	return ok
}

func KnownCommands() []string {
	commands := make([]string, 0, len(CommandSpecs))
	for name := range CommandSpecs {
		commands = append(commands, name)
	}
	sort.Strings(commands)
	return commands
}

func GetSubcommandSpec(command string, subcommand string) (SubcommandSpec, bool) {
	commandSpec, ok := GetCommandSpec(command)
	if !ok {
		return SubcommandSpec{}, false
	}
	subcommandSpec, ok := commandSpec.Subcommands[subcommand]
	return subcommandSpec, ok
}
