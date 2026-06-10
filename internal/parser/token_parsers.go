package parser

import (
	"strconv"
	"strings"
)

func trimFlagPrefix(raw string) string {
	if strings.HasPrefix(raw, "--") {
		raw = strings.TrimPrefix(raw, "--")
	} else {
		raw = strings.TrimPrefix(raw, "-")
	}
	return raw
}

func ParseTokenFlag(raw string) (Token, error) {
	var flagKey string
	var flagValue string
	var singleFlag bool

	if strings.Contains(raw, "=") {
		flagValue = strings.SplitN(raw, "=", 2)[1]
		singleFlag = false
	} else {
		flagValue = "true"
		singleFlag = true
	}

	if singleFlag {
		flagKey = trimFlagPrefix(raw)
	} else {
		flagKey = trimFlagPrefix(strings.SplitN(raw, "=", 2)[0])
	}

	return createFlagTokenWithRaw(raw, flagKey, flagValue), nil
}

func ParseTokenTag(raw string) (Token, error) {
	return Token{Kind: TokenTag, Raw: raw[1:]}, nil
}

func ParseTokenTagNegative(raw string) (Token, error) {
	return Token{Kind: TokenTagNegative, Raw: raw[2:]}, nil
}

type TokenAttributeValueParser func(raw string) (AttributeValue, error)

var tokenAttributeParsers = map[AttributeValueType]TokenAttributeValueParser{
	AttributeValueTypeString:  parseTokenAttributeStringValue,
	AttributeValueTypeInteger: parseTokenAttributeIntegerValue,
	AttributeValueTypeFloat:   parseTokenAttributeFloatValue,
	AttributeValueTypeFile:    parseTokenAttributeFileValue,
	AttributeValueTypeDate:    parseTokenAttributeDateValue,
	AttributeValueTypeBool:    parseTokenAttributeBoolValue,
}

func isRange(raw string) bool {
	return strings.HasPrefix(raw, "[") && strings.Contains(raw, "-") && strings.HasSuffix(raw, "]")
}

func isList(raw string) bool {
	return strings.Contains(raw, ",")
}

func parseTokenAttributeStringValue(raw string) (AttributeValue, error) {
	if isRange(raw) {
		parts := strings.Split(raw, "-")
		return AttributeValue{Kind: AttributeValueKindRange, Start: parts[0], End: parts[1]}, nil
	} else if isList(raw) {
		items := strings.Split(raw, ",")
		return AttributeValue{Kind: AttributeValueKindList, Items: items}, nil
	}
	return AttributeValue{Kind: AttributeValueKindSingle, Raw: raw}, nil
}

func checkIntegerValue(raw string) bool {
	_, err := strconv.Atoi(raw)
	return err == nil
}

func parseTokenAttributeIntegerValue(raw string) (AttributeValue, error) {
	if isRange(raw) {
		parts := strings.Split(raw, "-")
		startValue := parts[0]
		endValue := parts[1]
		if !checkIntegerValue(startValue) {
			return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid integer range start"}
		}
		if !checkIntegerValue(endValue) {
			return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid integer range end"}
		}
		return AttributeValue{Kind: AttributeValueKindRange, Start: parts[0], End: parts[1]}, nil
	} else if isList(raw) {
		items := strings.Split(raw, ",")
		for _, item := range items {
			if !checkIntegerValue(item) {
				return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid integer list item"}
			}
		}
		return AttributeValue{Kind: AttributeValueKindList, Items: items}, nil
	}
	if !checkIntegerValue(raw) {
		return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid integer"}
	}
	return AttributeValue{Kind: AttributeValueKindSingle, Raw: raw}, nil
}

func parseTokenAttributeFloatValue(raw string) (AttributeValue, error) {
	return AttributeValue{Kind: AttributeValueKindSingle, Raw: raw}, nil
}

func parseTokenAttributeFileValue(raw string) (AttributeValue, error) {
	return AttributeValue{Kind: AttributeValueKindSingle, Raw: raw}, nil
}

func parseTokenAttributeDateValue(raw string) (AttributeValue, error) {
	return AttributeValue{Kind: AttributeValueKindSingle, Raw: raw}, nil
}

func parseTokenAttributeBoolValue(raw string) (AttributeValue, error) {
	return AttributeValue{Kind: AttributeValueKindSingle, Raw: raw}, nil
}

func ParseTokenAttribute(raw string) (Token, error) {
	var attributeKey string
	var attributeValue AttributeValue
	var err error

	// Key and values
	attributeKey = strings.SplitN(raw, ":", 2)[0]
	attributeValueRaw := strings.SplitN(raw, ":", 2)[1]

	// Get attribute spec
	attrSpec, ok := AttributeSpecs[attributeKey]
	if !ok {
		return Token{}, &ParseError{Code: ParseErrorUnknownToken, Message: "unknown attribute"}
	}

	// Parse attribute value
	attributeValue, err = tokenAttributeParsers[attrSpec.Type](attributeValueRaw)
	if err != nil {
		return Token{}, &ParseError{Code: ParseErrorInvalidInput, Message: err.Error()}
	}

	return createAttributeToken(
		strings.SplitN(raw, ":", 2)[0],
		attributeValue,
	), nil
}

func ParseTokenText(raw string) (Token, error) {
	return Token{Kind: TokenText, Raw: raw}, nil
}
