package parser

import (
	"strconv"
	"strings"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/domain"
	"github.com/nschaetti/cashwarrior/internal/utils"
)

func trimFlagPrefix(raw string) string {
	if strings.HasPrefix(raw, "--") {
		raw = strings.TrimPrefix(raw, "--")
	} else {
		raw = strings.TrimPrefix(raw, "-")
	}
	return raw
}

func ParseTokenFlag(raw string, config config.Config) (Token, error) {
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

func ParseTokenTag(raw string, config config.Config) (Token, error) {
	return Token{Kind: TokenTag, Raw: raw[1:]}, nil
}

func ParseTokenTagNegative(raw string, config config.Config) (Token, error) {
	return Token{Kind: TokenTagNegative, Raw: raw[2:]}, nil
}

type TokenAttributeValueParser func(raw string, config config.Config) (AttributeValue, error)

var tokenAttributeParsers = map[AttributeValueType]TokenAttributeValueParser{
	AttributeValueTypeString:  parseTokenAttributeStringValue,
	AttributeValueTypeInteger: parseTokenAttributeIntegerValue,
	AttributeValueTypeFloat:   parseTokenAttributeFloatValue,
	AttributeValueTypeFile:    parseTokenAttributeFileValue,
	AttributeValueTypeDate:    parseTokenAttributeDateValue,
	AttributeValueTypeBool:    parseTokenAttributeBoolValue,
}

func isRange(raw string) bool {
	return strings.Contains(raw, "..")
}

func isList(raw string) bool {
	return strings.Contains(raw, ",")
}

func parseTokenAttributeStringValue(raw string, config config.Config) (AttributeValue, error) {
	if isRange(raw) {
		parts := strings.Split(raw, "..")
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

func checkFloatValue(raw string) bool {
	_, err := strconv.ParseFloat(raw, 64)
	return err == nil
}

func parseTokenAttributeIntegerValue(raw string, config config.Config) (AttributeValue, error) {
	if isRange(raw) {
		parts := strings.Split(raw, "..")
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

func parseTokenAttributeFloatValue(raw string, config config.Config) (AttributeValue, error) {
	if isRange(raw) {
		parts := strings.Split(raw, "..")
		startValue := parts[0]
		endValue := parts[1]
		if !checkFloatValue(startValue) {
			return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid float range start"}
		}
		if !checkFloatValue(endValue) {
			return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid float range end"}
		}
		return AttributeValue{Kind: AttributeValueKindRange, Start: parts[0], End: parts[1]}, nil
	} else if isList(raw) {
		items := strings.Split(raw, ",")
		for _, item := range items {
			if !checkFloatValue(item) {
				return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid float list item"}
			}
		}
		return AttributeValue{Kind: AttributeValueKindList, Items: items}, nil
	}
	if !checkFloatValue(raw) {
		return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid float"}
	}
	return AttributeValue{Kind: AttributeValueKindSingle, Raw: raw}, nil
}

func IsValidPath(path string) bool {
	return path != "" && !strings.ContainsRune(path, '\x00')
}

func parseTokenAttributeFileValue(raw string, config config.Config) (AttributeValue, error) {
	if IsValidPath(raw) {
		return AttributeValue{Kind: AttributeValueKindSingle, Raw: utils.ExpandPath(raw)}, nil
	}
	return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid file path"}
}

func checkDateValue(raw string, config config.Config) bool {
	_, err := time.Parse(config.Display.DateFormat, raw)
	isTS := domain.IsTimeShortcut(raw)
	return err == nil || isTS
}

func parseTokenAttributeDateValue(raw string, config config.Config) (AttributeValue, error) {
	if isRange(raw) {
		parts := strings.Split(raw, "..")
		startValue := parts[0]
		endValue := parts[1]
		if !checkDateValue(startValue, config) {
			return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid date range start"}
		}
		if !checkDateValue(endValue, config) {
			return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid date range end"}
		}
		return AttributeValue{Kind: AttributeValueKindRange, Start: parts[0], End: parts[1]}, nil
	} else if isList(raw) {
		items := strings.Split(raw, ",")
		for _, item := range items {
			if !checkDateValue(item, config) {
				return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid integer list item"}
			}
		}
		return AttributeValue{Kind: AttributeValueKindList, Items: items}, nil
	}
	if !checkDateValue(raw, config) {
		return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid date"}
	}
	return AttributeValue{Kind: AttributeValueKindSingle, Raw: raw}, nil
}

func parseTokenAttributeBoolValue(raw string, config config.Config) (AttributeValue, error) {
	return AttributeValue{Kind: AttributeValueKindSingle, Raw: raw}, nil
}

func ParseTokenAttribute(raw string, config config.Config) (Token, error) {
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
	attributeValue, err = tokenAttributeParsers[attrSpec.Type](attributeValueRaw, config)
	if err != nil {
		return Token{}, &ParseError{Code: ParseErrorInvalidInput, Message: err.Error()}
	}

	return createAttributeToken(
		strings.SplitN(raw, ":", 2)[0],
		attributeValue,
	), nil
}

func ParseTokenText(raw string, config config.Config) (Token, error) {
	return Token{Kind: TokenText, Raw: raw}, nil
}
