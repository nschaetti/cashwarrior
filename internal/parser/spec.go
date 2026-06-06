package parser

import (
	"fmt"
	"sort"
)

type AttributeValueShape uint8

const (
	AttributeValueShapeSingle AttributeValueShape = 1 << iota
	AttributeValueShapeList
	AttributeValueShapeRange
	AttributeValueShapeOperator
)

type AttributeValueType uint8

const (
	AttributeValueTypeString AttributeValueType = iota
	AttributeValueTypeInteger
	AttributeValueTypeFloat
	AttributeValueTypeDate
	AttributeValueTypeBool
	AttributeValueTypeFile
)

type AttributeSpec struct {
	Name      string
	Shapes    AttributeValueShape
	Type      AttributeValueType
	Settable  bool
	Clearable bool
}

func (s AttributeSpec) AllowsSet() bool {
	if !s.Settable && !s.Clearable {
		return true
	}
	return s.Settable
}

func (s AttributeSpec) AllowsClear() bool {
	return s.Clearable
}

var AttributeSpecs = map[string]AttributeSpec{
	"amount": {
		Name:      "amount",
		Shapes:    AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange,
		Type:      AttributeValueTypeFloat,
		Settable:  true,
		Clearable: false,
	},
	"account": {
		Name:      "account",
		Shapes:    AttributeValueShapeSingle | AttributeValueShapeList,
		Type:      AttributeValueTypeString,
		Settable:  true,
		Clearable: false,
	},
	"category": {
		Name:      "category",
		Shapes:    AttributeValueShapeSingle | AttributeValueShapeList,
		Type:      AttributeValueTypeString,
		Settable:  true,
		Clearable: true,
	},
	"currency": {
		Name:      "currency",
		Shapes:    AttributeValueShapeSingle | AttributeValueShapeList,
		Type:      AttributeValueTypeString,
		Settable:  false,
		Clearable: false,
	},
	"date": {
		Name:      "date",
		Shapes:    AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange,
		Type:      AttributeValueTypeDate,
		Settable:  true,
	},
	"desc": {
		Name:      "desc",
		 Shapes:    AttributeValueShapeSingle,
		 Type:      AttributeValueTypeBool,
		 Settable:  true,
		 Clearable: false,
	},
},

type CountRule struct {
	Min int
	Max int
}

type PresenceRule struct {
	Kinds      []TokenKind
	Attributes []string
	Message    string
}

type SideSpec struct {
	AllowedKinds   map[TokenKind]bool
	AllowAnyAttr   bool
	Attributes     map[string]AttributeSpec
	MinArgs        int
	MaxArgs        int
	KindRules      map[TokenKind]CountRule
	AttributeRules map[string]CountRule
	AtLeastOneOf   []PresenceRule
}

type SubcommandSpec struct {
	Name  string
	Left  SideSpec
	Right SideSpec
}

type CommandSpec struct {
	Name              string
	DefaultSubcommand string
	Subcommands       map[string]SubcommandSpec
}

func (c CommandSpec) String() string {
	return fmt.Sprintf(
		"<CommandSpec name=%s, default=%s, subcommands=%v>",
		c.Name,
		c.DefaultSubcommand,
		c.Subcommands,
	)
}

func allowKinds(kinds ...TokenKind) map[TokenKind]bool {
	allowed := make(map[TokenKind]bool, len(kinds))
	for _, kind := range kinds {
		allowed[kind] = true
	}
	return allowed
}

func attributes(specs ...AttributeSpec) map[string]AttributeSpec {
	attrs := make(map[string]AttributeSpec, len(specs))
	for _, spec := range specs {
		attrs[spec.Name] = spec
	}
	return attrs
}

func subcommands(specs ...SubcommandSpec) map[string]SubcommandSpec {
	items := make(map[string]SubcommandSpec, len(specs))
	for _, spec := range specs {
		items[spec.Name] = spec
	}
	return items
}

func allSupportedKinds() map[TokenKind]bool {
	return allowKinds(
		TokenTag,
		TokenTagNegative,
		TokenAttribute,
		TokenAttributeClear,
		TokenText,
	)
}

func genericSideSpec() SideSpec {
	return SideSpec{
		AllowedKinds:   allSupportedKinds(),
		AllowAnyAttr:   true,
		KindRules:      map[TokenKind]CountRule{},
		AttributeRules: map[string]CountRule{},
		AtLeastOneOf:   []PresenceRule{},
	}
}

// emptySideSpec specifies a side that accepts no tokens.
func emptySideSpec() SideSpec {
	return SideSpec{
		AllowedKinds:   map[TokenKind]bool{},
		Attributes:     map[string]AttributeSpec{},
		KindRules:      map[TokenKind]CountRule{},
		AttributeRules: map[string]CountRule{},
		AtLeastOneOf:   []PresenceRule{},
	}
}

func sideSpec(kinds []TokenKind, attrs ...AttributeSpec) SideSpec {
	return SideSpec{
		AllowedKinds:   allowKinds(kinds...),
		Attributes:     attributes(attrs...),
		KindRules:      map[TokenKind]CountRule{},
		AttributeRules: map[string]CountRule{},
		AtLeastOneOf:   []PresenceRule{},
	}
}

func sideSpecWithAnyAttributes(kinds ...TokenKind) SideSpec {
	return SideSpec{
		AllowedKinds:   allowKinds(kinds...),
		AllowAnyAttr:   true,
		KindRules:      map[TokenKind]CountRule{},
		AttributeRules: map[string]CountRule{},
		AtLeastOneOf:   []PresenceRule{},
	}
}

func countRule(min int, max int) CountRule {
	return CountRule{Min: min, Max: max}
}

func exactlyOne() CountRule {
	return CountRule{Min: 1, Max: 1}
}

func atLeastOne() CountRule {
	return CountRule{Min: 1}
}

func atMostOne() CountRule {
	return CountRule{Max: 1}
}

func withArgs(side SideSpec, min int, max int) SideSpec {
	side.MinArgs = min
	side.MaxArgs = max
	return side
}

func withKindRule(side SideSpec, kind TokenKind, rule CountRule) SideSpec {
	if side.KindRules == nil {
		side.KindRules = map[TokenKind]CountRule{}
	}
	side.KindRules[kind] = rule
	return side
}

func withAttributeRule(side SideSpec, name string, rule CountRule) SideSpec {
	if side.AttributeRules == nil {
		side.AttributeRules = map[string]CountRule{}
	}
	side.AttributeRules[name] = rule
	return side
}

func withAtLeastOneOf(side SideSpec, rule PresenceRule) SideSpec {
	side.AtLeastOneOf = append(side.AtLeastOneOf, rule)
	return side
}

func setOnlyAttribute(name string, shapes AttributeValueShape) AttributeSpec {
	return AttributeSpec{Name: name, Shapes: shapes, Settable: true}
}

func setOrClearAttribute(name string, shapes AttributeValueShape) AttributeSpec {
	return AttributeSpec{Name: name, Shapes: shapes, Settable: true, Clearable: true}
}

func transactionFilterSideSpec(extraAttrs ...AttributeSpec) SideSpec {
	base := []AttributeSpec{
		{Name: "account", Shapes: AttributeValueShapeSingle | AttributeValueShapeList},
		{Name: "currency", Shapes: AttributeValueShapeSingle | AttributeValueShapeList},
		{Name: "store", Shapes: AttributeValueShapeSingle | AttributeValueShapeList},
		{Name: "desc", Shapes: AttributeValueShapeSingle},
		{Name: "date", Shapes: AttributeValueShapeSingle | AttributeValueShapeRange | AttributeValueShapeList},
		{Name: "period", Shapes: AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange},
		{Name: "group", Shapes: AttributeValueShapeSingle | AttributeValueShapeList},
		{Name: "identifier", Shapes: AttributeValueShapeSingle | AttributeValueShapeList},
	}
	base = append(base, extraAttrs...)
	return sideSpec([]TokenKind{TokenText, TokenAttribute}, base...)
}

func transactionFilterSideSpecWithoutAccount(extraAttrs ...AttributeSpec) SideSpec {
	base := []AttributeSpec{
		{Name: "currency", Shapes: AttributeValueShapeSingle},
		{Name: "store", Shapes: AttributeValueShapeSingle},
		{Name: "desc", Shapes: AttributeValueShapeSingle},
		{Name: "date", Shapes: AttributeValueShapeSingle | AttributeValueShapeRange},
		{Name: "period", Shapes: AttributeValueShapeSingle},
		{Name: "group", Shapes: AttributeValueShapeSingle},
		{Name: "identifier", Shapes: AttributeValueShapeSingle},
	}
	base = append(base, extraAttrs...)
	return sideSpec([]TokenKind{TokenText, TokenAttribute}, base...)
}

// addRightSideSpec specifies the right side of the add command.
func addRightSideSpec() SideSpec {
	side := sideSpec(
		[]TokenKind{TokenTag, TokenAttribute, TokenText},
		setOnlyAttribute("date", AttributeValueShapeSingle),
		setOnlyAttribute("store", AttributeValueShapeSingle),
		setOnlyAttribute("account", AttributeValueShapeSingle),
		setOnlyAttribute("category", AttributeValueShapeSingle),
		setOnlyAttribute("group", AttributeValueShapeSingle),
	)
	side = withArgs(side, 2, 0)
	side = withAttributeRule(side, "amount", exactlyOne())
	for _, name := range []string{"datetime", "date", "time", "store", "account", "category", "group"} {
		side = withAttributeRule(side, name, atMostOne())
	}
	side = withAtLeastOneOf(side, PresenceRule{Kinds: []TokenKind{TokenText}, Message: "add requires a description"})
	return side
}

func modifyRightSideSpec() SideSpec {
	side := sideSpec(
		[]TokenKind{TokenAttribute, TokenAttributeClear, TokenTag, TokenTagNegative},
		setOnlyAttribute("identifier", AttributeValueShapeSingle),
		setOnlyAttribute("amount", AttributeValueShapeSingle),
		setOnlyAttribute("desc", AttributeValueShapeSingle),
		setOnlyAttribute("date", AttributeValueShapeSingle),
		setOnlyAttribute("account", AttributeValueShapeSingle),
		setOrClearAttribute("category", AttributeValueShapeSingle),
		setOnlyAttribute("store", AttributeValueShapeSingle),
		setOrClearAttribute("group", AttributeValueShapeSingle),
	)
	side = withArgs(side, 1, 0)
	for _, name := range []string{"identifier", "amount", "desc", "date", "time", "datetime", "account", "category", "store", "group"} {
		side = withAttributeRule(side, name, atMostOne())
	}
	return side
}

func transferRightSideSpec() SideSpec {
	side := sideSpec(
		[]TokenKind{TokenAttribute, TokenText},
		setOnlyAttribute("from", AttributeValueShapeSingle),
		setOnlyAttribute("to", AttributeValueShapeSingle),
		setOnlyAttribute("date", AttributeValueShapeSingle),
	)
	side = withAttributeRule(side, "amount", exactlyOne())
	side = withAttributeRule(side, "from", exactlyOne())
	side = withAttributeRule(side, "to", exactlyOne())
	side = withAttributeRule(side, "date", atMostOne())
	return side
}

func fakeitRightSideSpec() SideSpec {
	side := sideSpec(
		[]TokenKind{TokenAttribute, TokenText},
		setOnlyAttribute("account", AttributeValueShapeSingle),
		setOnlyAttribute("category", AttributeValueShapeSingle),
		setOnlyAttribute("type", AttributeValueShapeSingle),
		setOnlyAttribute("year", AttributeValueShapeSingle),
		setOnlyAttribute("month", AttributeValueShapeSingle),
	)
	side = withKindRule(side, TokenText, atMostOne())
	for _, name := range []string{"account", "category", "type", "year", "month"} {
		side = withAttributeRule(side, name, atMostOne())
	}
	return side
}

func budgetSideSpec() SideSpec {
	return SideSpec{
		AllowedKinds: allSupportedKinds(),
		Attributes: attributes(
			setOnlyAttribute("account", AttributeValueShapeSingle|AttributeValueShapeList),
			setOnlyAttribute("category", AttributeValueShapeSingle|AttributeValueShapeList),
			setOnlyAttribute("currency", AttributeValueShapeSingle|AttributeValueShapeList),
			setOnlyAttribute("date", AttributeValueShapeSingle|AttributeValueShapeRange),
			setOnlyAttribute("desc", AttributeValueShapeSingle),
			setOnlyAttribute("group", AttributeValueShapeSingle|AttributeValueShapeList),
			setOnlyAttribute("store", AttributeValueShapeSingle|AttributeValueShapeList),
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

var CommandSpecs = map[string]CommandSpec{
	"accounts": {
		Name:              "accounts",
		DefaultSubcommand: "list",
		Subcommands: subcommands(
			SubcommandSpec{Name: "list", Left: emptySideSpec(), Right: withArgs(emptySideSpec(), 0, 0)},
			SubcommandSpec{Name: "balance", Left: transactionFilterSideSpecWithoutAccount(), Right: withAtLeastOneOf(withAttributeRule(withArgs(sideSpec([]TokenKind{TokenText, TokenAttribute}, setOnlyAttribute("account", AttributeValueShapeSingle)), 1, 1), "account", atMostOne()), PresenceRule{Kinds: []TokenKind{TokenText}, Attributes: []string{"account"}, Message: "accounts balance requires an account name"})},
			SubcommandSpec{Name: "add", Left: emptySideSpec(), Right: withAtLeastOneOf(withAttributeRule(withAttributeRule(withArgs(sideSpec([]TokenKind{TokenText, TokenAttribute}, setOnlyAttribute("account", AttributeValueShapeSingle), setOnlyAttribute("currency", AttributeValueShapeSingle), setOnlyAttribute("initial_balance", AttributeValueShapeSingle)), 1, 3), "currency", atMostOne()), "initial_balance", atMostOne()), PresenceRule{Kinds: []TokenKind{TokenText}, Attributes: []string{"account"}, Message: "accounts add requires an account name"})},
			SubcommandSpec{Name: "modify", Left: emptySideSpec(), Right: withAtLeastOneOf(withAttributeRule(withAttributeRule(withAttributeRule(withKindRule(withArgs(sideSpec([]TokenKind{TokenText, TokenAttribute}, setOnlyAttribute("account", AttributeValueShapeSingle), setOnlyAttribute("currency", AttributeValueShapeSingle), setOnlyAttribute("initial_balance", AttributeValueShapeSingle)), 2, 4), TokenText, exactlyOne()), "account", atMostOne()), "currency", atMostOne()), "initial_balance", atMostOne()), PresenceRule{Attributes: []string{"account", "currency", "initial_balance"}, Message: "accounts modify requires at least one modification"})},
			SubcommandSpec{Name: "initial-balance", Left: emptySideSpec(), Right: withArgs(sideSpec([]TokenKind{TokenText, TokenAttribute}, setOnlyAttribute("account", AttributeValueShapeSingle), setOnlyAttribute("amount", AttributeValueShapeSingle), setOnlyAttribute("initial_balance", AttributeValueShapeSingle)), 2, 2)},
			SubcommandSpec{Name: "rename", Left: emptySideSpec(), Right: withKindRule(withArgs(sideSpec([]TokenKind{TokenText}), 2, 2), TokenText, countRule(2, 2))},
			SubcommandSpec{Name: "delete", Left: emptySideSpec(), Right: withAtLeastOneOf(withAttributeRule(withArgs(sideSpec([]TokenKind{TokenText, TokenAttribute}, setOnlyAttribute("account", AttributeValueShapeSingle)), 1, 1), "account", atMostOne()), PresenceRule{Kinds: []TokenKind{TokenText}, Attributes: []string{"account"}, Message: "accounts delete requires an account name"})},
		),
	},
	"add": {
		Name:              "add",
		DefaultSubcommand: "default",
		Subcommands:       subcommands(SubcommandSpec{Name: "default", Left: emptySideSpec(), Right: addRightSideSpec()}),
	},
	"backup": {
		Name:              "backup",
		DefaultSubcommand: "default",
		Subcommands:       subcommands(SubcommandSpec{Name: "default", Left: emptySideSpec(), Right: withAttributeRule(withArgs(sideSpec([]TokenKind{TokenAttribute}, setOnlyAttribute("output", AttributeValueShapeSingle)), 0, 1), "output", atMostOne())}),
	},
	"balance": defaultCommandSpec("balance"),
	"budget": {
		Name:              "budget",
		DefaultSubcommand: "list",
		Subcommands: subcommands(
			SubcommandSpec{Name: "list", Left: budgetSideSpec(), Right: budgetSideSpec()},
			SubcommandSpec{Name: "add", Left: budgetSideSpec(), Right: budgetSideSpec()},
		),
	},
	"by": defaultCommandSpec("by"),
	"categories": {
		Name:              "categories",
		DefaultSubcommand: "list",
		Subcommands: subcommands(
			SubcommandSpec{Name: "list", Left: emptySideSpec(), Right: withArgs(emptySideSpec(), 0, 0)},
			SubcommandSpec{Name: "add", Left: emptySideSpec(), Right: withAtLeastOneOf(withAttributeRule(withArgs(sideSpec([]TokenKind{TokenText, TokenAttribute}, setOnlyAttribute("category", AttributeValueShapeSingle), setOnlyAttribute("parent", AttributeValueShapeSingle)), 1, 2), "parent", atMostOne()), PresenceRule{Kinds: []TokenKind{TokenText}, Attributes: []string{"category"}, Message: "categories add requires a category name"})},
			SubcommandSpec{Name: "modify", Left: emptySideSpec(), Right: withAtLeastOneOf(withAttributeRule(withAttributeRule(withKindRule(withArgs(sideSpec([]TokenKind{TokenText, TokenAttribute}, setOnlyAttribute("category", AttributeValueShapeSingle), setOnlyAttribute("parent", AttributeValueShapeSingle)), 2, 3), TokenText, exactlyOne()), "category", atMostOne()), "parent", atMostOne()), PresenceRule{Attributes: []string{"category", "parent"}, Message: "categories modify requires at least one modification"})},
			SubcommandSpec{Name: "delete", Left: emptySideSpec(), Right: withAtLeastOneOf(withAttributeRule(withArgs(sideSpec([]TokenKind{TokenText, TokenAttribute}, setOnlyAttribute("category", AttributeValueShapeSingle)), 1, 1), "category", atMostOne()), PresenceRule{Kinds: []TokenKind{TokenText}, Attributes: []string{"category"}, Message: "categories delete requires a category name"})},
		),
	},
	"config": {
		Name:              "config",
		DefaultSubcommand: "default",
		Subcommands:       subcommands(SubcommandSpec{Name: "default", Left: emptySideSpec(), Right: withArgs(sideSpecWithAnyAttributes(TokenAttribute), 0, 1)}),
	},
	"delete": {
		Name:              "delete",
		DefaultSubcommand: "default",
		Subcommands: subcommands(
			SubcommandSpec{Name: "default", Left: emptySideSpec(), Right: withKindRule(withArgs(sideSpec([]TokenKind{TokenID}), 1, 1), TokenID, exactlyOne())},
			SubcommandSpec{Name: "default", Left: emptySideSpec(), Right: withAttributeRule()},
			SubcommandSpec{Name: "list", Left: emptySideSpec(), Right: withArgs(emptySideSpec(), 0, 0)},
		),
	},
	"fakeit": {
		Name:              "fakeit",
		DefaultSubcommand: "default",
		Subcommands:       subcommands(SubcommandSpec{Name: "default", Left: emptySideSpec(), Right: fakeitRightSideSpec()}),
	},
	"group": {
		Name:              "group",
		DefaultSubcommand: "default",
		Subcommands:       subcommands(SubcommandSpec{Name: "default", Left: emptySideSpec(), Right: withArgs(sideSpec([]TokenKind{TokenText}), 2, 0)}),
	},
	"groups": {
		Name:              "groups",
		DefaultSubcommand: "list",
		Subcommands: subcommands(
			SubcommandSpec{Name: "list", Left: sideSpec([]TokenKind{TokenAttribute},
				AttributeSpec{Name: "order", Shapes: AttributeValueShapeSingle},
				AttributeSpec{Name: "desc", Shapes: AttributeValueShapeSingle},
			), Right: withArgs(emptySideSpec(), 0, 0)},
		),
	},
	"import": {
		Name:              "import",
		DefaultSubcommand: "default",
		Subcommands:       subcommands(SubcommandSpec{Name: "default", Left: emptySideSpec(), Right: withArgs(sideSpec([]TokenKind{TokenText}), 1, 1)}),
	},
	"init": defaultCommandSpec("init"),
	"list": {
		Name:              "list",
		DefaultSubcommand: "default",
		Subcommands: subcommands(SubcommandSpec{Name: "default", Left: transactionFilterSideSpec(
			AttributeSpec{Name: "order", Shapes: AttributeValueShapeSingle},
			AttributeSpec{Name: "desc", Shapes: AttributeValueShapeSingle},
		), Right: genericSideSpec()}),
	},
	"modify": {
		Name:              "modify",
		DefaultSubcommand: "default",
		Subcommands:       subcommands(SubcommandSpec{Name: "default", Left: transactionFilterSideSpec(), Right: modifyRightSideSpec()}),
	},
	"places": {
		Name:              "places",
		DefaultSubcommand: "list",
		Subcommands: subcommands(
			SubcommandSpec{Name: "list", Left: emptySideSpec(), Right: withArgs(emptySideSpec(), 0, 0)},
			SubcommandSpec{Name: "rename", Left: emptySideSpec(), Right: withKindRule(withArgs(sideSpec([]TokenKind{TokenText}), 2, 2), TokenText, countRule(2, 2))},
		),
	},
	"purge": {
		Name:              "purge",
		DefaultSubcommand: "default",
		Subcommands:       subcommands(SubcommandSpec{Name: "default", Left: emptySideSpec(), Right: withKindRule(withArgs(sideSpec([]TokenKind{TokenID}), 1, 1), TokenID, exactlyOne())}),
	},
	"report": defaultCommandSpec("report"),
	"restore": {
		Name:              "restore",
		DefaultSubcommand: "default",
		Subcommands: subcommands(
			SubcommandSpec{Name: "default", Left: emptySideSpec(), Right: withKindRule(withArgs(sideSpec([]TokenKind{TokenID}), 1, 1), TokenID, exactlyOne())},
			SubcommandSpec{Name: "list", Left: emptySideSpec(), Right: withArgs(emptySideSpec(), 0, 0)},
		),
	},
	"set-balance": defaultCommandSpec("set-balance"),
	"show": {
		Name:              "show",
		DefaultSubcommand: "default",
		Subcommands:       subcommands(SubcommandSpec{Name: "default", Left: emptySideSpec(), Right: withKindRule(withArgs(sideSpec([]TokenKind{TokenID}), 1, 1), TokenID, exactlyOne())}),
	},
	"stats": defaultCommandSpec("stats"),
	"sum":   defaultCommandSpec("sum"),
	"summary": {
		Name:              "summary",
		DefaultSubcommand: "",
		Subcommands: subcommands(
			SubcommandSpec{Name: "days", Left: transactionFilterSideSpec(), Right: transactionFilterSideSpec()},
		),
	},
	"tags": {
		Name:              "tags",
		DefaultSubcommand: "list",
		Subcommands: subcommands(
			SubcommandSpec{Name: "list", Left: emptySideSpec(), Right: withArgs(emptySideSpec(), 0, 0)},
			SubcommandSpec{Name: "add", Left: emptySideSpec(), Right: withAtLeastOneOf(withArgs(sideSpec([]TokenKind{TokenText, TokenAttribute}, setOnlyAttribute("tag", AttributeValueShapeSingle)), 1, 1), PresenceRule{Kinds: []TokenKind{TokenText}, Attributes: []string{"tag"}, Message: "tags add requires a tag name"})},
			SubcommandSpec{Name: "modify", Left: emptySideSpec(), Right: withAtLeastOneOf(withAttributeRule(withKindRule(withArgs(sideSpec([]TokenKind{TokenText, TokenAttribute}, setOnlyAttribute("tag", AttributeValueShapeSingle)), 2, 2), TokenText, exactlyOne()), "tag", exactlyOne()), PresenceRule{Attributes: []string{"tag"}, Message: "tags modify requires a new tag name"})},
			SubcommandSpec{Name: "delete", Left: emptySideSpec(), Right: withAtLeastOneOf(withArgs(sideSpec([]TokenKind{TokenText, TokenAttribute}, setOnlyAttribute("tag", AttributeValueShapeSingle)), 1, 1), PresenceRule{Kinds: []TokenKind{TokenText}, Attributes: []string{"tag"}, Message: "tags delete requires a tag name"})},
		),
	},
	"theme": {
		Name:              "theme",
		DefaultSubcommand: "default",
		Subcommands:       subcommands(SubcommandSpec{Name: "default", Left: emptySideSpec(), Right: withArgs(sideSpec([]TokenKind{TokenText}), 0, 1)}),
	},
	"transfer": {
		Name:              "transfer",
		DefaultSubcommand: "default",
		Subcommands:       subcommands(SubcommandSpec{Name: "default", Left: emptySideSpec(), Right: transferRightSideSpec()}),
	},
	"undo": defaultCommandSpec("undo"),
}

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
