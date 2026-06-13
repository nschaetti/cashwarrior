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

func isRange(raw string) bool {
	return strings.Contains(raw, "..")
}

func isList(raw string) bool {
	return strings.Contains(raw, ",")
}

func checkIntegerValue(raw string) bool {
	_, err := strconv.Atoi(raw)
	return err == nil
}

func checkFloatValue(raw string) bool {
	_, err := strconv.ParseFloat(raw, 64)
	return err == nil
}

func checkDateValue(raw string, config config.Config) bool {
	_, err := time.Parse(config.Display.DateFormat, raw)
	isTS := domain.IsTimeShortcut(raw)
	return err == nil || isTS
}

func parseDateValue(raw string, config config.Config) (time.Time, error) {
	parsedDate, err := time.Parse(config.Display.DateFormat, raw)
	if err != nil {
		return time.Time{}, err
	}
	return parsedDate, nil
}

func IsValidPath(path string) bool {
	return path != "" && !strings.ContainsRune(path, '\x00')
}

func ParseArgFlag(raw string, config config.Config) (Arg, error) {
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
	return ArgFlag{Raw: raw, Key: flagKey, Value: flagValue}, nil
}

/*func ParseTokenFlag(raw string, config config.Config) (Token, error) {
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
}*/

/*
 * Parse text arguments
 */

func ParseArgText(raw string, config config.Config) (Arg, error) {
	return ArgText{Raw: raw, Text: raw}, nil
}

/*
 * Parse tag arguments
 */

func ParseArgTag(raw string, config config.Config) (Arg, error) {
	return ArgTag{Raw: raw, Tag: raw, Negative: false}, nil
}

/*func ParseTokenTag(raw string, config config.Config) (Token, error) {
	return Token{Kind: TokenTag, Raw: raw[1:]}, nil
}*/

/*
 * Parse tag negative arguments
 */

func ParseArgTagNegative(raw string, config config.Config) (Arg, error) {
	return ArgTag{Raw: raw, Tag: raw, Negative: true}, nil
}

/*func ParseTokenTagNegative(raw string, config config.Config) (Token, error) {
	return Token{Kind: TokenTagNegative, Raw: raw[2:]}, nil
}*/

/*
 * Parse attribute arguments
 */

type TokenAttributeValueParser func(raw string, config config.Config) (AttributeValue, error)

var tokenAttributeParsers = map[AttributeValueType]TokenAttributeValueParser{
	AttributeValueTypeString:  parseTokenAttributeStringValue,
	AttributeValueTypeInteger: parseTokenAttributeIntegerValue,
	AttributeValueTypeFloat:   parseTokenAttributeFloatValue,
	AttributeValueTypeFile:    parseTokenAttributeFileValue,
	AttributeValueTypeDate:    parseTokenAttributeDateValue,
	AttributeValueTypeBool:    parseTokenAttributeBoolValue,
}

func parseTokenAttributeStringValue(raw string, config config.Config) (AttributeValue, error) {
	if isRange(raw) {
		parts := strings.Split(raw, "..")
		return AttributeValue{
			ValueShape: AttributeValueShapeRange,
			Range: AttributeRange{
				Start: StringItem{Raw: parts[0], Value: parts[0]},
				End:   StringItem{Raw: parts[1], Value: parts[1]},
			},
		}, nil
	} else if isList(raw) {
		sitems := strings.Split(raw, ",")
		items := make([]AttributeItem, 0, len(sitems))
		for _, item := range sitems {
			items = append(items, StringItem{Raw: item, Value: item})
		}
		return AttributeValue{
			ValueShape: AttributeValueShapeList,
			Items:      items,
		}, nil
	}
	return AttributeValue{
		ValueShape: AttributeValueShapeSingle,
		Raw:        raw,
		Value:      StringItem{Raw: raw, Value: raw},
	}, nil
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
		startInt, _ := strconv.Atoi(startValue)
		endInt, _ := strconv.Atoi(endValue)
		if startInt > endInt {
			return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid integer range"}
		}
		return AttributeValue{
			ValueShape: AttributeValueShapeRange,
			Range: AttributeRange{
				Start: IntItem{Raw: parts[0], Value: int64(startInt)},
				End:   IntItem{Raw: parts[1], Value: int64(endInt)},
			},
		}, nil
	} else if isList(raw) {
		sitems := strings.Split(raw, ",")
		for _, item := range sitems {
			if !checkIntegerValue(item) {
				return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid integer list item"}
			}
		}
		items := make([]AttributeItem, 0, len(sitems))
		for _, item := range sitems {
			intVal, _ := strconv.Atoi(item)
			items = append(items, IntItem{Raw: item, Value: int64(intVal)})
		}
		return AttributeValue{
			ValueShape: AttributeValueShapeList,
			Items:      items,
		}, nil
	}
	if !checkIntegerValue(raw) {
		return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid integer"}
	}
	intVal, _ := strconv.Atoi(raw)
	return AttributeValue{
		ValueShape: AttributeValueShapeSingle,
		Raw:        raw,
		Value:      IntItem{Raw: raw, Value: int64(intVal)},
	}, nil
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
		floatStart, _ := strconv.ParseFloat(startValue, 64)
		floatEnd, _ := strconv.ParseFloat(endValue, 64)
		if floatStart > floatEnd {
			return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid float range"}
		}
		return AttributeValue{
			ValueShape: AttributeValueShapeRange,
			Range: AttributeRange{
				Start: FloatItem{Raw: parts[0], Value: floatStart},
				End:   FloatItem{Raw: parts[1], Value: floatEnd},
			},
		}, nil
	} else if isList(raw) {
		items := strings.Split(raw, ",")
		for _, item := range items {
			if !checkFloatValue(item) {
				return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid float list item"}
			}
		}
		floatItems := make([]AttributeItem, 0, len(items))
		for _, item := range items {
			floatItem, _ := strconv.ParseFloat(item, 64)
			floatItems = append(floatItems, FloatItem{Raw: item, Value: floatItem})
		}
		return AttributeValue{
			ValueShape: AttributeValueShapeList,
			Items:      floatItems,
		}, nil
	}
	if !checkFloatValue(raw) {
		return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid float"}
	}
	floatValue, _ := strconv.ParseFloat(raw, 64)
	return AttributeValue{
		ValueShape: AttributeValueShapeSingle,
		Raw:        raw,
		Value:      FloatItem{Raw: raw, Value: floatValue},
	}, nil
}

func parseTokenAttributeFileValue(raw string, config config.Config) (AttributeValue, error) {
	if IsValidPath(raw) {
		return AttributeValue{ValueShape: AttributeValueShapeSingle, Raw: utils.ExpandPath(raw)}, nil
	}
	return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid file path"}
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
		startDate, _ := parseDateValue(startValue, config)
		endDate, _ := parseDateValue(endValue, config)
		if startDate.After(endDate) {
			return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid date range"}
		}
		return AttributeValue{
			ValueShape: AttributeValueShapeRange,
			Range: AttributeRange{
				Start: TimeItem{Raw: parts[0], Value: startDate},
				End:   TimeItem{Raw: parts[1], Value: endDate},
			},
		}, nil
	} else if isList(raw) {
		items := strings.Split(raw, ",")
		for _, item := range items {
			if !checkDateValue(item, config) {
				return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid integer list item"}
			}
		}
		dateItems := make([]AttributeItem, 0, len(items))
		for _, item := range items {
			dateItem, _ := parseDateValue(item, config)
			dateItems = append(dateItems, TimeItem{Raw: item, Value: dateItem})
		}
		return AttributeValue{
			ValueShape: AttributeValueShapeList,
			Items:      dateItems,
		}, nil
	}
	if !checkDateValue(raw, config) {
		return AttributeValue{}, &ParseError{Code: ParseErrorInvalidInput, Message: "invalid date"}
	}
	dateValue, _ := parseDateValue(raw, config)
	return AttributeValue{
		ValueShape: AttributeValueShapeSingle,
		Raw:        raw,
		Value:      TimeItem{Raw: raw, Value: dateValue},
	}, nil
}

func parseTokenAttributeBoolValue(raw string, config config.Config) (AttributeValue, error) {
	return AttributeValue{
		ValueShape: AttributeValueShapeSingle,
		Raw:        raw,
	}, nil
}

func ParseArgAttribute(raw string, config config.Config) (Arg, error) {
	var attributeKey string
	var attributeValue AttributeValue
	var err error

	// Key and values
	attributeKey = strings.SplitN(raw, ":", 2)[0]
	attributeValueRaw := strings.SplitN(raw, ":", 2)[1]

	// Get attribute spec
	attrSpec, ok := AttributeSpecs[attributeKey]
	if !ok {
		return ArgAttribute{}, &ParseError{Code: ParseErrorUnknownToken, Message: "unknown attribute"}
	}

	// Parse attribute value
	attributeValue, err = tokenAttributeParsers[attrSpec.Type](attributeValueRaw, config)
	if err != nil {
		return ArgAttribute{}, &ParseError{Code: ParseErrorInvalidInput, Message: err.Error()}
	}

	return ArgAttribute{Raw: raw, Key: attributeKey, Value: attributeValue}, nil
}

/*func ParseTokenAttribute(raw string, config config.Config) (Token, error) {
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
}*/

/*func ParseTokenText(raw string, config config.Config) (Token, error) {
	return Token{Kind: TokenText, Raw: raw}, nil
}*/

/*
 * Parse clear attribute arguments
 */

func ParseArgAttributeClear(raw string, config config.Config) (Arg, error) {
	attributeKey := strings.SplitN(raw, ":", 2)[0]
	return ArgAttribute{Raw: raw, Key: attributeKey}, nil
}
