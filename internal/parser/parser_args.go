package parser

import "fmt"

type ArgKind int

const (
	ArgKindUnknown ArgKind = iota
	ArgKindTag
	ArgKindTagNegative
	ArgKindAttribute
	ArgKindText
	ArgKindFlag
)

func (k ArgKind) String() string {
	switch k {
	case ArgKindUnknown:
		return "unknown"
	case ArgKindTag:
		return "tag"
	case ArgKindTagNegative:
		return "tag-negative"
	case ArgKindAttribute:
		return "attribute"
	case ArgKindText:
		return "text"
	case ArgKindFlag:
		return "flag"
	default:
		return "unknown"
	}
}

type Arg interface {
	ArgKind() ArgKind
	RawString() string
	IsAttribute() bool
	IsAttributeClear() bool
	IsTag() bool
	IsTagNegative() bool
	IsText() bool
	IsFlag() bool
}

/*
 * ArgText represents a text argument.
 */

type ArgText struct {
	Raw  string
	Text string
}

func (t ArgText) ArgKind() ArgKind       { return ArgKindText }
func (t ArgText) RawString() string      { return t.Raw }
func (t ArgText) String() string         { return fmt.Sprintf("argtext(%s)", t.Text) }
func (t ArgText) IsText() bool           { return true }
func (t ArgText) IsAttribute() bool      { return false }
func (t ArgText) IsAttributeClear() bool { return false }
func (t ArgText) IsTag() bool            { return false }
func (t ArgText) IsTagNegative() bool    { return false }
func (t ArgText) IsFlag() bool           { return false }

/*
 * ArgAttribute represents an attribute argument.
 */

type ArgAttribute struct {
	Raw   string
	Key   string
	Value AttributeValue
	Clear bool
}

func (a ArgAttribute) ArgKind() ArgKind  { return ArgKindAttribute }
func (a ArgAttribute) RawString() string { return a.Raw }
func (a ArgAttribute) String() string {
	if a.Clear {
		return fmt.Sprintf("argattr-clear(%s)", a.Key)
	}
	return fmt.Sprintf("argattr(%s=%s)", a.Key, a.Value)
}
func (a ArgAttribute) IsAttribute() bool      { return true }
func (a ArgAttribute) IsAttributeClear() bool { return a.Clear }
func (a ArgAttribute) IsTag() bool            { return false }
func (a ArgAttribute) IsTagNegative() bool    { return false }
func (a ArgAttribute) IsText() bool           { return false }
func (a ArgAttribute) IsFlag() bool           { return false }

func createClearAttribute(key string) ArgAttribute {
	return ArgAttribute{Key: key, Clear: true}
}

func createAttribute(key string, value AttributeValue) ArgAttribute {
	return ArgAttribute{Key: key, Value: value}
}

/*
 * ArgTag represents a tag argument.
 */

type ArgTag struct {
	Raw      string
	Tag      string
	Negative bool
}

func (t ArgTag) ArgKind() ArgKind  { return ArgKindTag }
func (t ArgTag) RawString() string { return t.Raw }
func (t ArgTag) String() string {
	if t.Negative {
		return fmt.Sprintf("argtag-negative(%s)", t.Tag)
	}
	return fmt.Sprintf("argtag(%s)", t.Tag)
}
func (t ArgTag) IsTag() bool            { return true }
func (t ArgTag) IsTagNegative() bool    { return t.Negative }
func (t ArgTag) IsAttribute() bool      { return false }
func (t ArgTag) IsAttributeClear() bool { return false }
func (t ArgTag) IsText() bool           { return false }
func (t ArgTag) IsFlag() bool           { return false }

/*
 * ArgFlag represents a flag argument.
 */

type ArgFlag struct {
	Raw   string
	Key   string
	Value FlagValueItem
}

func (f ArgFlag) ArgKind() ArgKind       { return ArgKindFlag }
func (f ArgFlag) RawString() string      { return f.Raw }
func (f ArgFlag) IsTag() bool            { return false }
func (f ArgFlag) IsTagNegative() bool    { return false }
func (f ArgFlag) IsAttribute() bool      { return false }
func (f ArgFlag) IsAttributeClear() bool { return false }
func (f ArgFlag) IsText() bool           { return false }
func (f ArgFlag) IsFlag() bool           { return true }
func (f ArgFlag) String() string {
	return fmt.Sprintf("<%s=%s>", f.Key, f.Value)
}
