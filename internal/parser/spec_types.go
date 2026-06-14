package parser

import "fmt"

type FlagValueType uint8

const (
	FlagValueTypeString FlagValueType = iota
	FlagValueTypeInteger
	FlagValueTypeFloat
	FlagValueTypeBool
)

// String returns a string representation of the flag value type.
func (t FlagValueType) String() string {
	switch t {
	case FlagValueTypeString:
		return "string"
	case FlagValueTypeInteger:
		return "integer"
	case FlagValueTypeFloat:
		return "float"
	case FlagValueTypeBool:
		return "bool"
	default:
		panic("unhandled default case")
	}
}

type FlagSpec struct {
	Name string
	Type FlagValueType
}

func (s FlagSpec) String() string {
	return fmt.Sprintf("<FlagSpec name=%s, type=%s>", s.Name, s.Type)
}

func (s FlagSpec) IsBool() bool {
	return s.Type == FlagValueTypeBool
}

func (s FlagSpec) IsString() bool {
	return s.Type == FlagValueTypeString
}

func (s FlagSpec) IsInteger() bool {
	return s.Type == FlagValueTypeInteger
}

func (s FlagSpec) IsFloat() bool {
	return s.Type == FlagValueTypeFloat
}

// AttributeValueShape are the shapes of an attribute value (single, list, range, operator)
type AttributeValueShape uint8

const (
	AttributeValueShapeSingle AttributeValueShape = 1 << iota // ArgAttribute value can be a single value
	AttributeValueShapeList                                   // ArgAttribute value can be a list of values
	AttributeValueShapeRange                                  // ArgAttribute value can be a range of values
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
	retText := "<Shapes: "
	if s&AttributeValueShapeSingle != 0 {
		retText += "single "
	}
	if s&AttributeValueShapeList != 0 {
		retText += "list "
	}
	if s&AttributeValueShapeRange != 0 {
		retText += "range "
	}
	retText += ">"
	return retText
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

func (s AttributeSpec) IsAllowedValue(shape AttributeValueShape) bool {
	return s.Shapes&shape != 0
}

func (s AttributeSpec) IsAllowedShape(shape AttributeValueShape) bool {
	return s.AllowedShapes&shape != 0
}

// String returns a string representation of the attribute spec.
func (s AttributeSpec) String() string {
	return fmt.Sprintf(
		"<AttributeSpec name=%s, allowed_shapes=%s, shapes=%s, type=%s, settable=%t, clearable=%t>",
		s.Name,
		s.AllowedShapes,
		s.Shapes,
		s.Type,
		s.Settable,
		s.Clearable,
	)
}

func (s AttributeSpec) SetShapes(shapes AttributeValueShape) AttributeSpec {
	if !s.IsAllowedShape(shapes) {
		panic(fmt.Sprintf("SetShapes: try to set %v but %v allowed", shapes, s.AllowedShapes))
	}
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

// CountRule specifies the number of times a token of a given kind or attribute
// must appear in a sequence.
type CountRule struct {
	Min int // Min number of times the token must appear.
	Max int // Max number of times the token may appear.
}

// PresenceRule specifies that at least one of the given kinds or attributes
// must appear in a sequence.
type PresenceRule struct {
	Kinds      []ArgKind // List of kinds that must appear.
	Attributes []string  // List of attributes that must appear.
	Message    string    // Error message to display if the rule is not met.
}

// SideSpec specifies the allowed tokens on either side of a command.
type SideSpec struct {
	AllowedKinds   map[ArgKind]bool         // Allowed token kinds.
	AllowAnyAttr   bool                     // Allow any attribute.
	Attributes     map[string]AttributeSpec // List of allowed attributes.
	MinArgs        int                      // Minimum number of arguments (if 0, then no constraint)
	MaxArgs        int                      // Maximum number of arguments (if 0, then no constraint)
	KindRules      map[ArgKind]CountRule    // Count rules for each kind.
	AttributeRules map[string]CountRule     // Count rules for each attribute.
	AtLeastOneOf   []PresenceRule           // List of PresenceRules.
}

func (s SideSpec) IsEmpty() bool {
	if len(s.AllowedKinds) == 0 {
		return true
	}

	if len(s.AllowedKinds) == 1 && s.AllowedKinds[ArgKindAttribute] && len(s.Attributes) == 0 && !s.AllowAnyAttr {
		return true
	}

	return false
}

// WithArgs specifies the minimum and maximum number of arguments in a SideSpec.
func (s SideSpec) WithArgs(min int, max int) SideSpec {
	s.MinArgs = min
	s.MaxArgs = max
	return s
}

func (s SideSpec) WithKindRule(kind ArgKind, rule CountRule) SideSpec {
	if s.KindRules == nil {
		s.KindRules = map[ArgKind]CountRule{}
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

// SubcommandSpec specifies a subcommand of a command.
type SubcommandSpec struct {
	Name        string
	Description string
	Left        SideSpec
	Right       SideSpec
	IsAlias     bool
}

func (s SubcommandSpec) HasFilter() bool {
	return !s.Left.IsEmpty()
}

func (s SubcommandSpec) HasArgument() bool {
	return !s.Right.IsEmpty()
}

// CommandSpec specifies a command.
type CommandSpec struct {
	Name              string
	Description       string
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

type SubcommandAlias struct {
	SubcommandName  string
	SubcommandAlias string
}
