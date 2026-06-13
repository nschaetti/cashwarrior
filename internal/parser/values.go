package parser

import (
	"fmt"
	"strings"
	"time"
)

//type AttributeValueKind string
//
//const (
//	AttributeValueKindSingle AttributeValueKind = "single"
//	AttributeValueKindList   AttributeValueKind = "list"
//	AttributeValueKindRange  AttributeValueKind = "range"
//)

//type AttributeValue struct {
//	Raw   string
//	Kind  AttributeValueKind
//	Items []string
//	Start string
//	End   string
//	Clear bool
//}

type AttributeItem interface {
	Kind() AttributeValueType
	RawString() string
}

type StringItem struct {
	Raw   string
	Value string
}

func (s StringItem) Kind() AttributeValueType { return AttributeValueTypeString }
func (s StringItem) RawString() string        { return s.Raw }

type IntItem struct {
	Raw   string
	Value int64
}

func (i IntItem) Kind() AttributeValueType { return AttributeValueTypeInteger }
func (i IntItem) RawString() string        { return i.Raw }

type FloatItem struct {
	Raw   string
	Value float64
}

func (f FloatItem) Kind() AttributeValueType { return AttributeValueTypeFloat }
func (f FloatItem) RawString() string        { return f.Raw }

type TimeItem struct {
	Raw   string
	Value time.Time
}

func (t TimeItem) Kind() AttributeValueType { return AttributeValueTypeDate }
func (t TimeItem) RawString() string        { return t.Raw }

type BoolItem struct {
	Raw   string
	Value bool
}

func (b BoolItem) Kind() AttributeValueType { return AttributeValueTypeBool }
func (b BoolItem) RawString() string        { return b.Raw }

type FileItem struct {
	Raw   string
	Value string
}

func (f FileItem) Kind() AttributeValueType { return AttributeValueTypeFile }
func (f FileItem) RawString() string        { return f.Raw }

type AttributeRange struct {
	Start AttributeItem
	End   AttributeItem
}

type AttributeValue struct {
	Raw        string
	ValueShape AttributeValueShape
	Value      AttributeItem
	Items      []AttributeItem
	Range      AttributeRange
}

func (v AttributeValue) IsEmpty() bool {
	return v.Raw == ""
}

func (v AttributeValue) IsSingle() bool {
	return v.ValueShape == AttributeValueShapeSingle
}

func (v AttributeValue) IsList() bool {
	return v.ValueShape == AttributeValueShapeList
}

func (v AttributeValue) IsRange() bool {
	return v.ValueShape == AttributeValueShapeRange
}

func createSingleAttributeValue(raw string, value AttributeItem) AttributeValue {
	return AttributeValue{ValueShape: AttributeValueShapeSingle, Raw: raw, Value: value}
}

func createListAttributeValue(raw string, items ...AttributeItem) AttributeValue {
	return AttributeValue{ValueShape: AttributeValueShapeList, Raw: raw, Items: items}
}

func createRangeAttributeValue(start AttributeItem, end AttributeItem) AttributeValue {
	return AttributeValue{ValueShape: AttributeValueShapeRange, Range: AttributeRange{Start: start, End: end}}
}

func (v AttributeValue) String() string {

	switch v.ValueShape {
	case AttributeValueShapeList:
		items := make([]string, 0, len(v.Items))
		for _, item := range v.Items {
			items = append(items, item.RawString())
		}
		return fmt.Sprintf("list(%s)", strings.Join(items, ","))
	case AttributeValueShapeRange:
		return fmt.Sprintf("range(%s,%s)", v.Range.Start, v.Range.End)
	default:
		return fmt.Sprintf("single(%s)", v.Raw)
	}
}

func ParseAttributeValue(raw string) (AttributeValue, error) {
	if raw == "" {
		return AttributeValue{}, fmt.Errorf("attribute value cannot be empty")
	}

	if strings.Contains(raw, ",") {
		parts := strings.Split(raw, ",")
		items := make([]AttributeItem, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				return AttributeValue{}, fmt.Errorf("attribute list contains an empty item")
			}
			items = append(items, StringItem{Raw: part, Value: part})
		}
		return AttributeValue{Raw: raw, ValueShape: AttributeValueShapeList, Items: items}, nil
	}

	if isAttributeRange(raw) {
		parts := strings.SplitN(raw, "-", 2)
		return AttributeValue{
			Raw:        raw,
			ValueShape: AttributeValueShapeRange,
			Range: AttributeRange{
				Start: StringItem{Raw: parts[0], Value: parts[0]},
				End:   StringItem{Raw: parts[1], Value: parts[1]},
			},
		}, nil
	}

	return AttributeValue{Raw: raw, ValueShape: AttributeValueShapeSingle}, nil
}

func isAttributeRange(raw string) bool {
	if strings.Count(raw, "-") != 1 {
		return false
	}
	parts := strings.SplitN(raw, "-", 2)
	return parts[0] != "" && parts[1] != ""
}

func (s AttributeValueShape) Allows(value AttributeValue) bool {
	switch value.ValueShape {
	case AttributeValueShapeSingle:
		return s&AttributeValueShapeSingle != 0
	case AttributeValueShapeList:
		return s&AttributeValueShapeList != 0
	case AttributeValueShapeRange:
		return s&AttributeValueShapeRange != 0
	default:
		return false
	}
}
