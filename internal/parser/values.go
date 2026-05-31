package parser

import (
	"fmt"
	"strings"
)

type AttributeValueKind string

const (
	AttributeValueKindSingle AttributeValueKind = "single"
	AttributeValueKindList   AttributeValueKind = "list"
	AttributeValueKindRange  AttributeValueKind = "range"
)

type AttributeValue struct {
	Raw   string
	Kind  AttributeValueKind
	Items []string
	Start string
	End   string
}

func (v AttributeValue) String() string {
	switch v.Kind {
	case AttributeValueKindList:
		return fmt.Sprintf("list(%s)", strings.Join(v.Items, ","))
	case AttributeValueKindRange:
		return fmt.Sprintf("range(%s-%s)", v.Start, v.End)
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
		items := make([]string, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				return AttributeValue{}, fmt.Errorf("attribute list contains an empty item")
			}
			items = append(items, part)
		}
		return AttributeValue{Raw: raw, Kind: AttributeValueKindList, Items: items}, nil
	}

	if isAttributeRange(raw) {
		parts := strings.SplitN(raw, "-", 2)
		return AttributeValue{Raw: raw, Kind: AttributeValueKindRange, Start: parts[0], End: parts[1]}, nil
	}

	return AttributeValue{Raw: raw, Kind: AttributeValueKindSingle}, nil
}

func isAttributeRange(raw string) bool {
	if strings.Count(raw, "-") != 1 {
		return false
	}
	parts := strings.SplitN(raw, "-", 2)
	return parts[0] != "" && parts[1] != ""
}

func (s AttributeValueShape) Allows(value AttributeValue) bool {
	switch value.Kind {
	case AttributeValueKindSingle:
		return s&AttributeValueShapeSingle != 0
	case AttributeValueKindList:
		return s&AttributeValueShapeList != 0
	case AttributeValueKindRange:
		return s&AttributeValueShapeRange != 0
	default:
		return false
	}
}
